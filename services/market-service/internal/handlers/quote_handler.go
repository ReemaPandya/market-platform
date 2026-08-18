package handlers

import (
	"encoding/json"
	"market-service/internal/cache"
	"market-service/internal/indianapi"
	"market-service/internal/models"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func GetQuote(c *gin.Context) {
	symbol := strings.ToUpper(c.Param("symbol"))

	cacheKey := "quote:NSE:" + symbol

	// 1. Try Redis first
	cached, err := cache.Client.Get(cache.Ctx, cacheKey).Result()

	if err == nil {
		var quote models.Quote

		if json.Unmarshal([]byte(cached), &quote) == nil {
			quote.Source = "cache"

			c.JSON(http.StatusOK, quote)
			return
		}
	}

	if err != nil && err != redis.Nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "redis lookup failed",
		})
		return
	}

	// 2. Temporary mock NSE prices.
	// mockPrices := map[string]float64{
	// 	"RELIANCE": 1450.20,
	// 	"TCS":      3025.50,
	// 	"INFY":     1488.75,
	// 	"HDFCBANK": 965.40,
	// 	"SBIN":     812.30,
	// }

	// price, exists := mockPrices[symbol]

	// if !exists {
	// 	c.JSON(http.StatusNotFound, gin.H{
	// 		"error": "symbol not supported",
	// 	})
	// 	return
	// }
	price, err := indianapi.GetNSEPrice(symbol)

	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error":   "failed to fetch market quote",
			"details": err.Error(),
		})
		return
	}

	quote := models.Quote{
		Symbol:   symbol,
		Exchange: "NSE",
		LTP:      price,
		Source:   "indianapi",
	}

	// 3. Store quote in Redis.
	data, err := json.Marshal(quote)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to encode quote",
		})
		return
	}

	if err := cache.Client.Set(
		cache.Ctx,
		cacheKey,
		data,
		60*time.Second,
	).Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to cache quote",
		})
		return
	}

	c.JSON(http.StatusOK, quote)
}
