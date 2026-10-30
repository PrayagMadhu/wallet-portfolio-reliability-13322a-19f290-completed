package provider

import (
	"context"
	"strings"
	"sync"
	"time"

	"keyrail.local/wallet-portfolio/internal/models"
)

type FakeMarketProvider struct {
	mu       sync.Mutex
	balances map[string][]models.TokenBalance
	prices   map[string]float64
	delays   map[string]time.Duration
	failures map[string]error
	calls    map[string]int
}

func NewFakeMarketProvider() *FakeMarketProvider {
	return &FakeMarketProvider{
		balances: map[string][]models.TokenBalance{
			"0xabc123": {
				{Symbol: "ETH", Name: "Ethereum", Chain: "ethereum", Contract: "native", Amount: 1.25},
				{Symbol: "USDC", Name: "USD Coin", Chain: "ethereum", Contract: "0xa0b86991", Amount: 240.00},
				{Symbol: "SOL", Name: "Solana", Chain: "solana", Contract: "native", Amount: 12.00},
				{Symbol: "KEY", Name: "Keyrail Test Token", Chain: "base", Contract: "0xkeyrail", Amount: 500.00},
			},
			"0xempty": {},
		},
		prices: map[string]float64{
			"ETH":  3200.00,
			"USDC": 1.00,
			"SOL":  145.00,
		},
		delays: map[string]time.Duration{
			"ETH":  25 * time.Millisecond,
			"USDC": 20 * time.Millisecond,
			"SOL":  30 * time.Millisecond,
			"KEY":  35 * time.Millisecond,
		},
		failures: map[string]error{
			"KEY": ErrPriceUnavailable,
		},
		calls: map[string]int{},
	}
}

func (f *FakeMarketProvider) WalletBalances(ctx context.Context, address string) ([]models.TokenBalance, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	normalized := strings.ToLower(address)
	balances, ok := f.balances[normalized]
	if !ok {
		return nil, ErrWalletNotFound
	}
	out := make([]models.TokenBalance, len(balances))
	copy(out, balances)
	return out, nil
}

func (f *FakeMarketProvider) TokenPriceUSD(ctx context.Context, symbol string) (float64, error) {
	symbol = strings.ToUpper(symbol)
	f.mu.Lock()
	f.calls[symbol]++
	delay := f.delays[symbol]
	failure := f.failures[symbol]
	price, hasPrice := f.prices[symbol]
	f.mu.Unlock()

	if delay > 0 {
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return 0, ctx.Err()
		}
	}

	if failure != nil {
		return 0, failure
	}
	if !hasPrice {
		return 0, ErrPriceUnavailable
	}
	return price, nil
}

func (f *FakeMarketProvider) Calls(symbol string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls[strings.ToUpper(symbol)]
}
