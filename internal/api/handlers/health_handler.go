package handlers

import (
	"net/http"

	"backend/internal/cache"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

type HealthHandler struct {
	db    *sqlx.DB
	redis *cache.RedisCache
}

func NewHealthHandler(db *sqlx.DB, redis *cache.RedisCache) *HealthHandler {
	return &HealthHandler{db: db, redis: redis}
}

func (h *HealthHandler) Live(c *gin.Context) {
	Respond(c, http.StatusOK, "ok", map[string]string{"status": "ok"})
}

func (h *HealthHandler) Ready(c *gin.Context) {
	if err := h.db.Ping(); err != nil {
		Respond[any](c, http.StatusServiceUnavailable, "db not ready", nil)
		return
	}
	if err := h.redis.Ping(); err != nil {
		Respond[any](c, http.StatusServiceUnavailable, "redis not ready", nil)
		return
	}
	Respond(c, http.StatusOK, "ok", map[string]string{"status": "ok"})
}
