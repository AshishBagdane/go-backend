package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "no-referrer")

		// Allow Swagger UI assets to load in non-prod environments.
		if strings.HasPrefix(c.Request.URL.Path, "/swagger-ui") {
			c.Header("Content-Security-Policy", "default-src 'self'; style-src 'self' 'unsafe-inline'; script-src 'self'; img-src 'self' data:; connect-src 'self'; font-src 'self'")
		} else {
			c.Header("Content-Security-Policy", "default-src 'none'")
		}
		c.Next()
	}
}
