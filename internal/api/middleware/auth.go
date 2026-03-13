package middleware

import (
	"net/http"
	"strings"

	"backend/internal/api/handlers"

	"github.com/gin-gonic/gin"
)

func APIKeyAuth(apiKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if apiKey == "" {
			handlers.Respond[any](c, http.StatusInternalServerError, "auth misconfigured", nil)
			c.Abort()
			return
		}

		key := c.GetHeader("X-API-Key")
		if key == "" {
			authHeader := c.GetHeader("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				key = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		if key != apiKey {
			handlers.Respond[any](c, http.StatusUnauthorized, "unauthorized", nil)
			c.Abort()
			return
		}

		c.Next()
	}
}
