package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/matt-j-deasy/betty-crokers-api/config"
)

func CORSMiddleware(cfg config.Environment) gin.HandlerFunc {
	return func(c *gin.Context) {
		allowedOrigins := map[string]bool{
			"http://localhost:3000": true,
			"http://localhost:3001": true,
			cfg.FrontEndURL:         true,
		}

		origin := c.Request.Header.Get("Origin")

		corsOrigin := cfg.FrontEndURL

		if origin != "" && allowedOrigins[origin] {
			corsOrigin = origin
		}

		c.Writer.Header().Set("Access-Control-Allow-Origin", corsOrigin)
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
