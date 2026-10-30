package service

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"keyrail.local/wallet-portfolio/internal/cache"
	"keyrail.local/wallet-portfolio/internal/models"
	"keyrail.local/wallet-portfolio/internal/provider"
)

type scriptProvider struct {
	balances []models.TokenBalance
	prices   map[string]float64
	delays   map[string]time.Duration
	fails    map[string]error

	mu       sync.Mutex
	calls    map[string]int
	inFlight int
	maxSeen  int
}

func (s *scriptProvider) WalletBalances(ctx context.Context, address string) ([]models.TokenBalance, error) {
	if address == "0xmissing" {
		return nil, provider.ErrWalletNotFound
	}
	out := make([]models.TokenBalance, len(s.balances))
	copy(out, s.balances)
	return out, nil
}

func (s *scriptProvider) TokenPriceUSD(ctx context.Context, symbol string) (float64, error) {
	s.mu.Lock()
	if s.calls == nil {
		s.calls = map[string]int{}
	}
	s.calls[symbol]++
	s.inFlight++
	if s.inFlight > s.maxSeen {
		s.maxSeen = s.inFlight
	}
	delay := s.delays[symbol]
	failure := s.fails[symbol]
	price := s.prices[symbol]
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		s.inFlight--
		s.mu.Unlock()
	}()

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
	return price, nil
}

func (s *scriptProvider) count(symbol string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.calls[symbol]
}

func TestSummaryKeepsSlowPriceLookupTimely(t *testing.T) {
	sp := &scriptProvider{
		balances: []models.TokenBalance{
			{Symbol: "ETH", Name: "Ethereum", Chain: "ethereum", Amount: 1},
			{Symbol: "SLOW", Name: "Slow Token", Chain: "ethereum", Amount: 2},
		},
		prices: map[string]float64{"ETH": 3000, "SLOW": 10},
		delays: map[string]time.Duration{"ETH": 10 * time.Millisecond, "SLOW": 350 * time.Millisecond},
		fails:  map[string]error{},
	}
	svc := NewPortfolioService(sp, cache.NewTTLCache(30*time.Second))

	started := time.Now()
	got, err := svc.Summary(context.Background(), "0xabc123")
	elapsed := time.Since(started)
	if err != nil {
		t.Fatalf("Summary returned error: %v", err)
	}
	if elapsed > 250*time.Millisecond {
		t.Fatalf("summary took too long: %s", elapsed)
	}
	if len(got.Tokens) != 2 {
		t.Fatalf("expected 2 tokens, got %d", len(got.Tokens))
	}
	if got.Tokens[1].PriceAvailable {
		t.Fatalf("slow token should remain listed without an available price")
	}
	if got.TotalUSD != 3000 {
		t.Fatalf("total should exclude unavailable price, got %.2f", got.TotalUSD)
	}
}

func TestSummaryUsesCachedPricesOnRepeatedRequest(t *testing.T) {
	sp := &scriptProvider{
		balances: []models.TokenBalance{{Symbol: "ETH", Name: "Ethereum", Chain: "ethereum", Amount: 2}},
		prices:   map[string]float64{"ETH": 3000},
		delays:   map[string]time.Duration{},
		fails:    map[string]error{},
	}
	svc := NewPortfolioService(sp, cache.NewTTLCache(30*time.Second))

	for i := 0; i < 2; i++ {
		got, err := svc.Summary(context.Background(), "0xabc123")
		if err != nil {
			t.Fatalf("Summary returned error: %v", err)
		}
		if got.TotalUSD != 6000 {
			t.Fatalf("unexpected total: %.2f", got.TotalUSD)
		}
	}
	if sp.count("ETH") != 1 {
		t.Fatalf("expected one provider price call, got %d", sp.count("ETH"))
	}
}

func TestConcurrentSummariesSharePriceLookup(t *testing.T) {
	sp := &scriptProvider{
		balances: []models.TokenBalance{{Symbol: "ETH", Name: "Ethereum", Chain: "ethereum", Amount: 1}},
		prices:   map[string]float64{"ETH": 3000},
		delays:   map[string]time.Duration{"ETH": 25 * time.Millisecond},
		fails:    map[string]error{},
	}
	svc := NewPortfolioService(sp, cache.NewTTLCache(30*time.Second))

	const requests = 12
	start := make(chan struct{})
	errs := make(chan error, requests)
	var wg sync.WaitGroup
	for i := 0; i < requests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			got, err := svc.Summary(context.Background(), "0xabc123")
			if err == nil && got.TotalUSD != 3000 {
				err = errors.New("unexpected portfolio total")
			}
			errs <- err
		}()
	}
	close(start)
	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatalf("Summary returned error: %v", err)
		}
	}
	if sp.count("ETH") != 1 {
		t.Fatalf("expected one shared provider price call, got %d", sp.count("ETH"))
	}
}

func TestSummaryLooksUpDifferentPricesInParallel(t *testing.T) {
	sp := &scriptProvider{
		balances: []models.TokenBalance{
			{Symbol: "ETH", Name: "Ethereum", Chain: "ethereum", Amount: 1},
			{Symbol: "SOL", Name: "Solana", Chain: "solana", Amount: 1},
		},
		prices: map[string]float64{"ETH": 3000, "SOL": 100},
		delays: map[string]time.Duration{"ETH": 30 * time.Millisecond, "SOL": 30 * time.Millisecond},
		fails:  map[string]error{},
	}
	svc := NewPortfolioService(sp, cache.NewTTLCache(30*time.Second))

	if _, err := svc.Summary(context.Background(), "0xabc123"); err != nil {
		t.Fatalf("Summary returned error: %v", err)
	}
	if sp.maxSeen < 2 {
		t.Fatalf("expected parallel price calls, maximum in flight was %d", sp.maxSeen)
	}
}

func TestInvalidAddress(t *testing.T) {
	sp := &scriptProvider{}
	svc := NewPortfolioService(sp, cache.NewTTLCache(30*time.Second))
	_, err := svc.Summary(context.Background(), "not-an-address")
	if !errors.Is(err, ErrInvalidAddress) {
		t.Fatalf("expected invalid address error, got %v", err)
	}
}
