package models

type TokenBalance struct {
	Symbol   string  `json:"symbol"`
	Name     string  `json:"name"`
	Chain    string  `json:"chain"`
	Contract string  `json:"contract"`
	Amount   float64 `json:"amount"`
}

type TokenHolding struct {
	Symbol         string   `json:"symbol"`
	Name           string   `json:"name"`
	Chain          string   `json:"chain"`
	Amount         float64  `json:"amount"`
	PriceUSD       *float64 `json:"price_usd,omitempty"`
	ValueUSD       *float64 `json:"value_usd,omitempty"`
	PriceAvailable bool     `json:"price_available"`
}

type PortfolioResponse struct {
	Address  string         `json:"address"`
	Tokens   []TokenHolding `json:"tokens"`
	TotalUSD float64        `json:"total_usd"`
}

type ErrorBody struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
