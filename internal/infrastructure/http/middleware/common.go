package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORSMiddleware creates CORS middleware
func CORSMiddleware() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}

// LoggerMiddleware creates a logging middleware
func LoggerMiddleware() gin.HandlerFunc {
	return gin.Logger()
}

// RecoveryMiddleware creates a panic recovery middleware
func RecoveryMiddleware() gin.HandlerFunc {
	return gin.Recovery()
}
