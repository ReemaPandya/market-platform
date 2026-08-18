package main

import (
	"log"
	"net/http"
	"os"

	"market-service/internal/cache"
	"market-service/internal/handlers"
	"market-service/internal/logger"
	"market-service/internal/websocket"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func main() {
	if err := logger.Init(); err != nil {
		log.Fatal("failed to initialize logger: ", err)
	}
	defer logger.Sync()

	if err := godotenv.Overload(); err != nil {
		logger.Log.Info("no .env file found, using environment variables")
	}

	if err := cache.ConnectRedis(); err != nil {
		logger.Log.Fatal(
			"redis connection failed",
			zap.Error(err),
		)
	}

	logger.Log.Info(
		"redis connected successfully",
		zap.String("service", "market-service"),
	)

	defer cache.Client.Close()

	r := gin.Default()

	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	r.GET("/health", func(c *gin.Context) {
		if err := cache.Client.Ping(cache.Ctx).Err(); err != nil {
			logger.Log.Error(
				"health check failed",
				zap.String("dependency", "redis"),
				zap.Error(err),
			)

			c.JSON(http.StatusServiceUnavailable, gin.H{
				"service": "market-service",
				"status":  "unhealthy",
				"redis":   "disconnected",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"service": "market-service",
			"status":  "healthy",
			"redis":   "connected",
		})
	})

	r.GET("/api/market/quote/:symbol", handlers.GetQuote)
	r.GET("/api/market/ws/:symbol", websocket.StreamMarket)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8083"
	}

	logger.Log.Info(
		"starting market service",
		zap.String("service", "market-service"),
		zap.String("port", port),
	)

	if err := r.Run(":" + port); err != nil {
		logger.Log.Fatal(
			"market service stopped",
			zap.Error(err),
		)
	}
}
