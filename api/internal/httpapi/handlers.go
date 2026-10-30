package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"keyrail.local/wallet-portfolio/internal/models"
	"keyrail.local/wallet-portfolio/internal/provider"
	"keyrail.local/wallet-portfolio/internal/service"
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
			writeError(w, http.StatusNotFound, "wallet_not_found", "Wallet not found")
			return
		}
		if errors.Is(err, service.ErrInvalidAddress) {
			writeError(w, http.StatusBadRequest, "invalid_address", "Invalid wallet address")
			return
		}
		if errors.Is(err, service.ErrProviderUnavailable) {
			writeError(w, http.StatusServiceUnavailable, "provider_unavailable", "Wallet data is temporarily unavailable")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "Unable to load portfolio")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(portfolio)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(models.ErrorBody{
		Error: models.ErrorDetail{Code: code, Message: message},
	})
}
