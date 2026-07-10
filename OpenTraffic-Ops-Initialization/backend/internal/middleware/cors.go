package middleware

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORSMiddleware CORS中间件
func CORSMiddleware() gin.HandlerFunc {
	config := cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:3000", "http://localhost:18080"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * 60 * 60, // 12 hours
	}

	return cors.New(config)
}
