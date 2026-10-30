package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"keyrail.local/wallet-portfolio/internal/cache"
	"keyrail.local/wallet-portfolio/internal/models"
	"keyrail.local/wallet-portfolio/internal/provider"
)

var (
	ErrInvalidAddress      = errors.New("invalid wallet address")
	ErrProviderUnavailable = errors.New("provider unavailable")
)

type PortfolioService struct {
	provider        provider.MarketProvider
	cache           *cache.TTLCache
	priceTimeout    time.Duration
	maxPriceLookups int
	priceSlots      chan struct{}
	priceMu         sync.Mutex
	priceCalls      map[string]*priceCall
}

type priceCall struct {
	done  chan struct{}
	price float64
	err   error
}

func NewPortfolioService(provider provider.MarketProvider, cache *cache.TTLCache) *PortfolioService {
	return &PortfolioService{
		provider:        provider,
		cache:           cache,
		priceTimeout:    120 * time.Millisecond,
		maxPriceLookups: 5,
		priceSlots:      make(chan struct{}, 5),
		priceCalls:      make(map[string]*priceCall),
	}
}

func (s *PortfolioService) Summary(ctx context.Context, address string) (models.PortfolioResponse, error) {
	address = strings.ToLower(strings.TrimSpace(address))
	if !looksLikeAddress(address) {
		return models.PortfolioResponse{}, ErrInvalidAddress
	}

	balances, err := s.provider.WalletBalances(ctx, address)
	if err != nil {
		if errors.Is(err, provider.ErrWalletNotFound) {
			return models.PortfolioResponse{}, err
		}
		return models.PortfolioResponse{}, fmt.Errorf("%w: %w", ErrProviderUnavailable, err)
	}

	response := models.PortfolioResponse{Address: address, Tokens: make([]models.TokenHolding, 0, len(balances))}
	type tokenPrice struct {
		price float64
		ok    bool
	}
	prices := make([]tokenPrice, len(balances))
	uniqueSymbols := make(map[string][]int)
	for i, balance := range balances {
		key := strings.ToUpper(balance.Symbol)
		uniqueSymbols[key] = append(uniqueSymbols[key], i)
	}

	type lookup struct {
		key     string
		symbol  string
		indexes []int
	}
	lookups := make([]lookup, 0, len(uniqueSymbols))
	for key, indexes := range uniqueSymbols {
		lookups = append(lookups, lookup{key: key, symbol: balances[indexes[0]].Symbol, indexes: indexes})
	}

	priceCtx, cancelPrices := context.WithTimeout(ctx, s.priceTimeout)
	defer cancelPrices()

	workers := s.maxPriceLookups
	if workers > len(lookups) {
		workers = len(lookups)
	}
	jobs := make(chan lookup)
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for item := range jobs {
				price, err := s.lookupPrice(priceCtx, item.key, item.symbol)
				if err == nil {
					for _, index := range item.indexes {
						prices[index] = tokenPrice{price: price, ok: true}
					}
				}
			}
		}()
	}
	for _, item := range lookups {
		jobs <- item
	}
	close(jobs)
	wg.Wait()

	for i, balance := range balances {
		holding := models.TokenHolding{
			Symbol: balance.Symbol,
			Name:   balance.Name,
			Chain:  balance.Chain,
			Amount: balance.Amount,
		}

		if prices[i].ok {
			price := prices[i].price
			value := balance.Amount * price
			holding.PriceUSD = &price
			holding.ValueUSD = &value
			holding.PriceAvailable = true
			response.TotalUSD += value
		}
		response.Tokens = append(response.Tokens, holding)
	}

	return response, nil
}

func (s *PortfolioService) lookupPrice(ctx context.Context, key, symbol string) (float64, error) {
	if price, ok := s.cache.Get(key); ok {
		return price, nil
	}

	s.priceMu.Lock()
	if price, ok := s.cache.Get(key); ok {
		s.priceMu.Unlock()
		return price, nil
	}
	if err := ctx.Err(); err != nil {
		s.priceMu.Unlock()
		return 0, err
	}
	if call, ok := s.priceCalls[key]; ok {
		s.priceMu.Unlock()
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case <-call.done:
			return call.price, call.err
		}
	}

	call := &priceCall{done: make(chan struct{})}
	s.priceCalls[key] = call
	s.priceMu.Unlock()

	go func() {
		priceCtx, cancel := context.WithTimeout(context.Background(), s.priceTimeout)
		defer cancel()

		select {
		case s.priceSlots <- struct{}{}:
			defer func() { <-s.priceSlots }()
		case <-priceCtx.Done():
			call.err = priceCtx.Err()
		}
		if call.err == nil {
			call.price, call.err = s.provider.TokenPriceUSD(priceCtx, symbol)
			if call.err == nil {
				s.cache.Set(key, call.price)
			}
		}

		s.priceMu.Lock()
		delete(s.priceCalls, key)
		close(call.done)
		s.priceMu.Unlock()
	}()

	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	case <-call.done:
		return call.price, call.err
	}
}

func looksLikeAddress(address string) bool {
	return strings.HasPrefix(address, "0x") && len(address) >= 6
}
