package main

import (
	"auth-service/internal/config"
	"auth-service/internal/handlers"
	"auth-service/internal/logger"
	"auth-service/internal/metrics"
	"auth-service/internal/middleware"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func main() {

	// Load environment variables
	godotenv.Overload()

	// Initialize logger
	logger.InitLogger()

	defer logger.Log.Sync()

	// Connect PostgreSQL
	config.ConnectDB()

	// Register Prometheus metrics
	metrics.RegisterMetrics()

	// Create Gin router
	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:4200"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))
	// Metrics middleware
	router.Use(middleware.MetricsMiddleware())

	// Health check
	router.GET("/health", func(c *gin.Context) {

		err := config.DB.Ping()

		if err != nil {

			logger.Log.Error(
				"database health check failed",
				zap.String("service", "auth-service"),
				zap.Error(err),
			)

			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "database down",
			})

			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
		})
	})

	// Prometheus metrics endpoint
	router.GET(
		"/metrics",
		gin.WrapH(promhttp.Handler()),
	)

	// Public routes
	router.POST("/signup", handlers.SignupHandler)

	router.POST("/login", handlers.LoginHandler)

	// Protected routes
	protected := router.Group("/")

	protected.Use(middleware.JWTMiddleware())

	protected.GET("/profile", func(c *gin.Context) {

		c.JSON(http.StatusOK, gin.H{
			"message": "protected profile route",
		})
	})

	// Start server
	port := os.Getenv("PORT")

	logger.Log.Info(
		"starting auth service",
		zap.String("service", "auth-service"),
		zap.String("port", port),
	)

	err := router.Run(":" + port)

	if err != nil {

		logger.Log.Fatal(
			"failed to start server",
			zap.Error(err),
		)
	}
}
