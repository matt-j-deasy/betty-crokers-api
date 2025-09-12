package middleware

import (
	"context"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// SlogMiddleware logs all HTTP requests with appropriate levels
func SlogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		c.Next() // Process the request

		latency := time.Since(startTime)
		status := c.Writer.Status()
		method := c.Request.Method
		path := c.Request.URL.Path
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()

		// Determine log level based on response status
		var logLevel slog.Level
		switch {
		case status >= 500:
			logLevel = slog.LevelError // Server errors
		case status >= 400:
			logLevel = slog.LevelWarn // Client errors
		default:
			logLevel = slog.LevelDebug // Successful requests
		}

		// Log the request at the appropriate level with context
		slog.LogAttrs(context.Background(), logLevel, "HTTP Request",
			slog.String("method", method),
			slog.String("path", path),
			slog.Int("status", status),
			slog.Duration("latency", latency),
			slog.String("clientIP", clientIP),
			slog.String("userAgent", userAgent),
		)
	}
}
