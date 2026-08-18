package indianapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type FlexibleFloat float64

func (f *FlexibleFloat) UnmarshalJSON(data []byte) error {
	// Handle normal JSON number: 2447.60
	var number float64

	if err := json.Unmarshal(data, &number); err == nil {
		*f = FlexibleFloat(number)
		return nil
	}

	// Handle JSON string: "2447.60"
	var text string

	if err := json.Unmarshal(data, &text); err != nil {
		return err
	}

	text = strings.ReplaceAll(text, ",", "")

	number, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return err
	}

	*f = FlexibleFloat(number)

	return nil
}

type StockResponse struct {
	CompanyName string `json:"companyName"`

	CurrentPrice struct {
		BSE FlexibleFloat `json:"BSE"`
		NSE FlexibleFloat `json:"NSE"`
	} `json:"currentPrice"`
}

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

func GetNSEPrice(symbol string) (float64, error) {
	apiKey := os.Getenv("INDIAN_API_KEY")

	if apiKey == "" {
		return 0, fmt.Errorf("INDIAN_API_KEY is not set")
	}

	endpoint := "https://stock.indianapi.in/stock?name=" +
		url.QueryEscape(symbol)

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return 0, err
	}

	req.Header.Set("x-api-key", apiKey)

	resp, err := httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf(
			"IndianAPI returned status %d",
			resp.StatusCode,
		)
	}

	var stock StockResponse

	if err := json.NewDecoder(resp.Body).Decode(&stock); err != nil {
		return 0, err
	}

	price := float64(stock.CurrentPrice.NSE)

	if price <= 0 {
		return 0, fmt.Errorf("NSE price unavailable for %s", symbol)
	}

	return price, nil
}
