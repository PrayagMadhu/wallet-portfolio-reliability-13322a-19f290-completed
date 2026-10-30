package provider

import (
	"context"
	"errors"

	"keyrail.local/wallet-portfolio/internal/models"
)

var (
	ErrWalletNotFound    = errors.New("wallet not found")
	ErrPriceUnavailable  = errors.New("price unavailable")
	ErrProviderTemporary = errors.New("provider temporarily unavailable")
)

type MarketProvider interface {
	WalletBalances(ctx context.Context, address string) ([]models.TokenBalance, error)
	TokenPriceUSD(ctx context.Context, symbol string) (float64, error)
}
