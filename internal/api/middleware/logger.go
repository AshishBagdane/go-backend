package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

func Logger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		reqID, _ := c.Get(RequestIDHeader)

		logger.Info(
			"request",
			slog.String("method", c.Request.Method),
			slog.String("path", c.FullPath()),
			slog.Int("status", status),
			slog.String("ip", c.ClientIP()),
			slog.Duration("latency", latency),
			slog.Any("request_id", reqID),
		)
	}
}
