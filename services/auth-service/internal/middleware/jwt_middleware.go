package middleware

import (
	"auth-service/internal/metrics"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func JWTMiddleware() gin.HandlerFunc {

	return func(c *gin.Context) {

		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "missing token",
			})
			c.Abort()
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")

		if token != "valid-token" {

			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid token",
			})

			c.Abort()
			return
		}

		c.Next()
	}
}

func MetricsMiddleware() gin.HandlerFunc {

	return func(c *gin.Context) {

		metrics.HttpRequests.WithLabelValues(
			c.FullPath(),
			c.Request.Method,
		).Inc()

		c.Next()
	}
}
