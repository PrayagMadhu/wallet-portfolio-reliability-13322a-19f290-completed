package main

import (
	"log"
	"net/http"
	"time"

	"keyrail.local/wallet-portfolio/internal/cache"
	"keyrail.local/wallet-portfolio/internal/httpapi"
	"keyrail.local/wallet-portfolio/internal/provider"
	"keyrail.local/wallet-portfolio/internal/service"
)

func main() {
	market := provider.NewFakeMarketProvider()
	priceCache := cache.NewTTLCache(30 * time.Second)
	portfolio := service.NewPortfolioService(market, priceCache)
	handler := httpapi.NewHandler(portfolio)

	mux := http.NewServeMux()
	mux.Handle("/api/wallets/", handler)

	log.Println("wallet portfolio API listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
