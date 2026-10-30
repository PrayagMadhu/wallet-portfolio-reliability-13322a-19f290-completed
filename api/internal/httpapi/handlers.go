package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"keyrail.local/wallet-portfolio/internal/models"
	"keyrail.local/wallet-portfolio/internal/provider"
)

type portfolioService interface {
	Summary(ctx context.Context, address string) (models.PortfolioResponse, error)
}

type Handler struct {
	portfolio portfolioService
}

func NewHandler(portfolio portfolioService) *Handler {
	return &Handler{portfolio: portfolio}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	address := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/wallets/"), "/portfolio")

	portfolio, err := h.portfolio.Summary(r.Context(), address)
	if err != nil {
		if errors.Is(err, provider.ErrWalletNotFound) {
			w.WriteHeader(http.StatusOK)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(portfolio)
}
