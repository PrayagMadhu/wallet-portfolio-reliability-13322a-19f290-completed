package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"keyrail.local/wallet-portfolio/internal/cache"
	"keyrail.local/wallet-portfolio/internal/provider"
	"keyrail.local/wallet-portfolio/internal/service"
)

func TestPortfolioInvalidAddressUsesStableErrorContract(t *testing.T) {
	h := NewHandler(service.NewPortfolioService(provider.NewFakeMarketProvider(), cache.NewTTLCache(30*time.Second)))
	req := httptest.NewRequest(http.MethodGet, "/api/wallets/not-valid/portfolio", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d with body %q", rr.Code, rr.Body.String())
	}
	var body struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("expected JSON error body: %v", err)
	}
	if body.Error.Code != "invalid_address" || body.Error.Message == "" {
		t.Fatalf("unexpected error body: %+v", body.Error)
	}
}

func TestPortfolioUnknownWalletStatus(t *testing.T) {
	h := NewHandler(service.NewPortfolioService(provider.NewFakeMarketProvider(), cache.NewTTLCache(30*time.Second)))
	req := httptest.NewRequest(http.MethodGet, "/api/wallets/0xmissing/portfolio", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d with body %q", rr.Code, rr.Body.String())
	}
	var body struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("expected JSON error body: %v", err)
	}
	if body.Error.Code != "wallet_not_found" {
		t.Fatalf("unexpected error body: %+v", body.Error)
	}
}

func TestPortfolioSuccessShape(t *testing.T) {
	h := NewHandler(service.NewPortfolioService(provider.NewFakeMarketProvider(), cache.NewTTLCache(30*time.Second)))
	req := httptest.NewRequest(http.MethodGet, "/api/wallets/0xabc123/portfolio", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d with body %q", rr.Code, rr.Body.String())
	}
	var body struct {
		Address  string `json:"address"`
		Tokens   []any  `json:"tokens"`
		TotalUSD int    `json:"total_usd"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("expected JSON success body: %v", err)
	}
	if body.Address != "0xabc123" || len(body.Tokens) == 0 {
		t.Fatalf("unexpected success body: %+v", body)
	}
}
