package handlers

import (
	"net/http"
	"strings"

	"trading-service/internal/config"
	"trading-service/internal/market"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

var ordersTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "trading_orders_total",
		Help: "Total number of successfully filled trading orders.",
	},
	[]string{"side"},
)

var orderFailuresTotal = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "trading_order_failures_total",
		Help: "Total number of order requests that failed.",
	},
)

func init() {
	prometheus.MustRegister(
		ordersTotal,
		orderFailuresTotal,
	)

	ordersTotal.WithLabelValues("BUY").Add(0)
	ordersTotal.WithLabelValues("SELL").Add(0)
}

type CreateOrderRequest struct {
	Symbol   string `json:"symbol" binding:"required"`
	Exchange string `json:"exchange"`
	Side     string `json:"side" binding:"required"`
	Quantity int    `json:"quantity" binding:"required"`
}

func CreateOrder(c *gin.Context) {
	orderSucceeded := false

	defer func() {
		if !orderSucceeded {
			orderFailuresTotal.Inc()
		}
	}()
	var req CreateOrderRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid order request",
		})
		return
	}

	req.Symbol = strings.ToUpper(req.Symbol)
	req.Exchange = strings.ToUpper(req.Exchange)
	req.Side = strings.ToUpper(req.Side)

	if req.Exchange == "" {
		req.Exchange = "NSE"
	}

	if req.Side != "BUY" && req.Side != "SELL" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "side must be BUY or SELL",
		})
		return
	}

	if req.Quantity <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "quantity must be greater than zero",
		})
		return
	}

	emailValue, exists := c.Get("email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user identity missing",
		})
		return
	}

	email := emailValue.(string)

	// Temporary until Market Service + Upstox are connected.
	executedPrice, err := market.GetNSEPrice(req.Symbol)

	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error":   "unable to obtain market price",
			"details": err.Error(),
		})
		return
	}

	tx, err := config.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to start transaction",
		})
		return
	}

	defer tx.Rollback()

	var userID int64

	err = tx.QueryRow(
		"SELECT id FROM users WHERE email = $1",
		email,
	).Scan(&userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "unable to resolve user",
		})
		return
	}

	// For SELL, verify that the user owns enough shares.
	if req.Side == "SELL" {
		var currentQuantity int

		err = tx.QueryRow(`
			SELECT quantity
			FROM positions
			WHERE user_id = $1
			  AND symbol = $2
			  AND exchange = $3
			FOR UPDATE
		`,
			userID,
			req.Symbol,
			req.Exchange,
		).Scan(&currentQuantity)

		if err != nil || currentQuantity < req.Quantity {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "insufficient position",
			})
			return
		}
	}

	var orderID int64

	err = tx.QueryRow(`
		INSERT INTO orders (
			user_id,
			symbol,
			exchange,
			side,
			quantity,
			executed_price,
			status
		)
		VALUES ($1, $2, $3, $4, $5, $6, 'FILLED')
		RETURNING id
	`,
		userID,
		req.Symbol,
		req.Exchange,
		req.Side,
		req.Quantity,
		executedPrice,
	).Scan(&orderID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create order",
		})
		return
	}

	var tradeID int64

	err = tx.QueryRow(`
		INSERT INTO trades (
			order_id,
			user_id,
			symbol,
			exchange,
			side,
			quantity,
			price
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`,
		orderID,
		userID,
		req.Symbol,
		req.Exchange,
		req.Side,
		req.Quantity,
		executedPrice,
	).Scan(&tradeID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create trade",
		})
		return
	}

	if req.Side == "BUY" {
		_, err = tx.Exec(`
			INSERT INTO positions (
				user_id,
				symbol,
				exchange,
				quantity,
				average_price
			)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (user_id, symbol, exchange)
			DO UPDATE SET
				average_price =
					(
						(positions.quantity * positions.average_price)
						+
						(EXCLUDED.quantity * EXCLUDED.average_price)
					)
					/
					(positions.quantity + EXCLUDED.quantity),
				quantity = positions.quantity + EXCLUDED.quantity,
				updated_at = NOW()
		`,
			userID,
			req.Symbol,
			req.Exchange,
			req.Quantity,
			executedPrice,
		)
	} else {
		_, err = tx.Exec(`
			UPDATE positions
			SET
				quantity = quantity - $4,
				updated_at = NOW()
			WHERE user_id = $1
			  AND symbol = $2
			  AND exchange = $3
		`,
			userID,
			req.Symbol,
			req.Exchange,
			req.Quantity,
		)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to update position",
		})
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to commit trade",
		})
		return
	}

	ordersTotal.WithLabelValues(req.Side).Inc()
	orderSucceeded = true

	c.JSON(http.StatusCreated, gin.H{
		"order_id":       orderID,
		"trade_id":       tradeID,
		"symbol":         req.Symbol,
		"exchange":       req.Exchange,
		"side":           req.Side,
		"quantity":       req.Quantity,
		"executed_price": executedPrice,
		"status":         "FILLED",
	})
}

func GetOrders(c *gin.Context) {
	emailValue, exists := c.Get("email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user identity missing",
		})
		return
	}

	email := emailValue.(string)

	rows, err := config.DB.Query(`
		SELECT
			o.id,
			o.symbol,
			o.exchange,
			o.side,
			o.quantity,
			o.executed_price,
			o.status,
			o.created_at
		FROM orders o
		JOIN users u ON o.user_id = u.id
		WHERE u.email = $1
		ORDER BY o.created_at DESC
	`, email)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to fetch orders",
		})
		return
	}

	defer rows.Close()

	type Order struct {
		ID            int64   `json:"id"`
		Symbol        string  `json:"symbol"`
		Exchange      string  `json:"exchange"`
		Side          string  `json:"side"`
		Quantity      int     `json:"quantity"`
		ExecutedPrice float64 `json:"executed_price"`
		Status        string  `json:"status"`
		CreatedAt     string  `json:"created_at"`
	}

	orders := []Order{}

	for rows.Next() {
		var order Order

		err := rows.Scan(
			&order.ID,
			&order.Symbol,
			&order.Exchange,
			&order.Side,
			&order.Quantity,
			&order.ExecutedPrice,
			&order.Status,
			&order.CreatedAt,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to read orders",
			})
			return
		}

		orders = append(orders, order)
	}

	c.JSON(http.StatusOK, gin.H{
		"orders": orders,
	})
}

func GetPortfolio(c *gin.Context) {
	emailValue, exists := c.Get("email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user identity missing",
		})
		return
	}

	email := emailValue.(string)

	rows, err := config.DB.Query(`
		SELECT
			p.symbol,
			p.exchange,
			p.quantity,
			p.average_price
		FROM positions p
		JOIN users u ON p.user_id = u.id
		WHERE u.email = $1
		  AND p.quantity > 0
		ORDER BY p.symbol
	`, email)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to fetch portfolio",
		})
		return
	}

	defer rows.Close()

	type Position struct {
		Symbol        string  `json:"symbol"`
		Exchange      string  `json:"exchange"`
		Quantity      int     `json:"quantity"`
		AveragePrice  float64 `json:"average_price"`
		CurrentPrice  float64 `json:"current_price"`
		InvestedValue float64 `json:"invested_value"`
		CurrentValue  float64 `json:"current_value"`
		UnrealizedPnL float64 `json:"unrealized_pnl"`
	}

	positions := []Position{}

	for rows.Next() {
		var position Position

		err := rows.Scan(
			&position.Symbol,
			&position.Exchange,
			&position.Quantity,
			&position.AveragePrice,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to read portfolio",
			})
			return
		}

		position.InvestedValue =
			float64(position.Quantity) * position.AveragePrice

		currentPrice, err := market.GetNSEPrice(position.Symbol)

		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{
				"error":   "failed to fetch current market price",
				"symbol":  position.Symbol,
				"details": err.Error(),
			})
			return
		}

		position.CurrentPrice = currentPrice

		position.CurrentValue =
			float64(position.Quantity) * position.CurrentPrice

		position.UnrealizedPnL =
			position.CurrentValue - position.InvestedValue

		positions = append(positions, position)
	}

	c.JSON(http.StatusOK, gin.H{
		"positions": positions,
	})
}
