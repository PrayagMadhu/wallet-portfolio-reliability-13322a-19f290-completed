package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"keyrail.local/wallet-portfolio/internal/cache"
	"keyrail.local/wallet-portfolio/internal/models"
	"keyrail.local/wallet-portfolio/internal/provider"
)

var (
	ErrInvalidAddress       = errors.New("invalid wallet address")
	ErrProviderUnavailable = errors.New("provider unavailable")
)

type PortfolioService struct {
	provider        provider.MarketProvider
	cache           *cache.TTLCache
	priceTimeout    time.Duration
	maxPriceLookups int
}

func NewPortfolioService(provider provider.MarketProvider, cache *cache.TTLCache) *PortfolioService {
	return &PortfolioService{
		provider:        provider,
		cache:           cache,
		priceTimeout:    120 * time.Millisecond,
		maxPriceLookups: 5,
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
		return models.PortfolioResponse{}, fmt.Errorf("load balances: %w", err)
	}

	response := models.PortfolioResponse{Address: address, Tokens: make([]models.TokenHolding, 0, len(balances))}
	for _, balance := range balances {
		holding := models.TokenHolding{
			Symbol: balance.Symbol,
			Name:   balance.Name,
			Chain:  balance.Chain,
			Amount: balance.Amount,
		}

		cacheKey := strings.ToUpper(balance.Symbol)
		price, ok := s.cache.Get(cacheKey)
		if !ok {
			price, err = s.provider.TokenPriceUSD(context.Background(), balance.Symbol)
			if err == nil {
				s.cache.Set(cacheKey, price)
			}
		}

		if err == nil {
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

func looksLikeAddress(address string) bool {
	return strings.HasPrefix(address, "0x") && len(address) >= 6
}
