package websocket

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"market-service/internal/cache"
	"market-service/internal/indianapi"
	"market-service/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func StreamMarket(c *gin.Context) {
	symbol := strings.ToUpper(c.Param("symbol"))

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	sendQuote := func() error {
		cacheKey := "quote:NSE:" + symbol

		cached, err := cache.Client.Get(cache.Ctx, cacheKey).Result()

		if err == nil {
			var quote models.Quote

			if json.Unmarshal([]byte(cached), &quote) == nil {
				quote.Source = "cache"
				return conn.WriteJSON(quote)
			}
		}

		if err != nil && err != redis.Nil {
			return err
		}

		price, err := indianapi.GetNSEPrice(symbol)
		if err != nil {
			return err
		}

		quote := models.Quote{
			Symbol:   symbol,
			Exchange: "NSE",
			LTP:      price,
			Source:   "indianapi",
		}

		data, err := json.Marshal(quote)
		if err != nil {
			return err
		}

		if err := cache.Client.Set(
			cache.Ctx,
			cacheKey,
			data,
			60*time.Second,
		).Err(); err != nil {
			return err
		}

		return conn.WriteJSON(quote)
	}

	// Send immediately when client connects.
	if err := sendQuote(); err != nil {
		return
	}

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if err := sendQuote(); err != nil {
			return
		}
	}
}
