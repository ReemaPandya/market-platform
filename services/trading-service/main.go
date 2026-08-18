package main

import (
	"log"
	"net/http"

	"trading-service/internal/config"
	"trading-service/internal/handlers"
	"trading-service/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	if err := godotenv.Overload(); err != nil {
		log.Println("no .env file found, using environment variables")
	}

	config.ConnectDB()
	defer config.DB.Close()

	router := gin.Default()
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	router.GET("/health", func(c *gin.Context) {
		if err := config.DB.Ping(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"service":  "trading-service",
				"status":   "unhealthy",
				"database": "disconnected",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"service":  "trading-service",
			"status":   "healthy",
			"database": "connected",
		})
	})
	protected := router.Group("/api")
	protected.Use(middleware.JWTMiddleware())
	protected.POST("/orders", handlers.CreateOrder)
	protected.GET("/orders", handlers.GetOrders)
	protected.GET("/portfolio", handlers.GetPortfolio)
	protected.GET("/test", func(c *gin.Context) {
		email, _ := c.Get("email")

		c.JSON(http.StatusOK, gin.H{
			"message": "trading service authentication successful",
			"email":   email,
		})
	})
	log.Println("starting trading service on :8082")

	if err := router.Run(":8082"); err != nil {
		log.Fatal(err)
	}
}
