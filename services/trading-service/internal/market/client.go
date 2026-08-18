package market

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

type Quote struct {
	Symbol   string  `json:"symbol"`
	Exchange string  `json:"exchange"`
	LTP      float64 `json:"ltp"`
	Source   string  `json:"source"`
}

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

func GetNSEPrice(symbol string) (float64, error) {
	baseURL := os.Getenv("MARKET_SERVICE_URL")

	if baseURL == "" {
		return 0, fmt.Errorf("MARKET_SERVICE_URL is not set")
	}

	url := strings.TrimRight(baseURL, "/") +
		"/api/market/quote/" +
		strings.ToUpper(symbol)

	resp, err := httpClient.Get(url)
	if err != nil {
		return 0, fmt.Errorf("market service unavailable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf(
			"market service returned status %d",
			resp.StatusCode,
		)
	}

	var quote Quote

	if err := json.NewDecoder(resp.Body).Decode(&quote); err != nil {
		return 0, err
	}

	if quote.LTP <= 0 {
		return 0, fmt.Errorf("invalid market price")
	}

	return quote.LTP, nil
}
