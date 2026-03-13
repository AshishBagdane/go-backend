package middleware

import (
	"context"
	"math"
	"net/http"
	"strconv"
	"sync"
	"time"

	"backend/internal/api/handlers"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"golang.org/x/time/rate"
)

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type RateLimitConfig struct {
	RPS   float64
	Burst int
}

type RateLimiterConfig struct {
	Default     RateLimitConfig
	Routes      map[string]RateLimitConfig
	UseRedis    bool
	RedisClient *redis.Client
	RedisPrefix string
}

var (
	visitors = make(map[string]*visitor)
	mu       sync.Mutex
)

func getVisitor(key string, cfg RateLimitConfig) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	if v, ok := visitors[key]; ok {
		v.lastSeen = time.Now()
		return v.limiter
	}

	limiter := rate.NewLimiter(rate.Limit(cfg.RPS), cfg.Burst)
	visitors[key] = &visitor{limiter: limiter, lastSeen: time.Now()}
	return limiter
}

func cleanupVisitors() {
	for {
		time.Sleep(time.Minute)

		mu.Lock()
		for key, v := range visitors {
			if time.Since(v.lastSeen) > 3*time.Minute {
				delete(visitors, key)
			}
		}
		mu.Unlock()
	}
}

func init() {
	go cleanupVisitors()
}

func RateLimiter(cfg RateLimiterConfig) gin.HandlerFunc {

	if cfg.Default.RPS <= 0 {
		cfg.Default.RPS = 5
	}
	if cfg.Default.Burst <= 0 {
		cfg.Default.Burst = 10
	}

	redisEnabled := cfg.UseRedis && cfg.RedisClient != nil
	if cfg.RedisPrefix == "" {
		cfg.RedisPrefix = "rate_limit"
	}

	return func(c *gin.Context) {

		routePath := c.FullPath()
		if routePath == "" {
			routePath = c.Request.URL.Path
		}

		routeKey := c.Request.Method + " " + routePath
		routeCfg, ok := cfg.Routes[routeKey]
		if !ok {
			routeCfg = cfg.Default
		}

		key := c.ClientIP() + ":" + routeKey

		if redisEnabled {
			allowed, remaining, resetSeconds, err := allowRedis(c.Request.Context(), cfg.RedisClient, cfg.RedisPrefix+":"+key, routeCfg)
			if err == nil {
				writeRateLimitHeadersFixed(c, routeCfg.RPS, remaining, resetSeconds)
				if !allowed {
					handlers.Respond[any](c, http.StatusTooManyRequests, "rate limit exceeded", nil)
					c.Abort()
					return
				}
				c.Next()
				return
			}
		}

		limiter := getVisitor(key, routeCfg)
		writeRateLimitHeaders(c, limiter)

		if !limiter.Allow() {
			handlers.Respond[any](c, http.StatusTooManyRequests, "rate limit exceeded", nil)
			c.Abort()
			return
		}

		c.Next()
	}
}

func writeRateLimitHeaders(c *gin.Context, limiter *rate.Limiter) {
	limit := limiter.Limit()
	if limit <= 0 {
		return
	}

	remaining := int(limiter.Tokens())
	if remaining < 0 {
		remaining = 0
	}

	resetSeconds := 0
	if remaining < 1 {
		now := time.Now()
		r := limiter.ReserveN(now, 1)
		if r.OK() {
			resetSeconds = int(r.DelayFrom(now).Seconds())
			r.CancelAt(now)
		}
	}

	c.Header("X-RateLimit-Limit", strconv.FormatFloat(float64(limit), 'f', -1, 64))
	c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
	c.Header("X-RateLimit-Reset", strconv.Itoa(resetSeconds))
}

func writeRateLimitHeadersFixed(c *gin.Context, limit float64, remaining int, resetSeconds int) {
	if limit <= 0 {
		return
	}
	c.Header("X-RateLimit-Limit", strconv.FormatFloat(limit, 'f', -1, 64))
	c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
	c.Header("X-RateLimit-Reset", strconv.Itoa(resetSeconds))
}

var redisRateScript = redis.NewScript(`
local current = redis.call("INCR", KEYS[1])
if current == 1 then
  redis.call("EXPIRE", KEYS[1], ARGV[1])
end
local ttl = redis.call("TTL", KEYS[1])
return {current, ttl}
`)

func allowRedis(ctx context.Context, client *redis.Client, key string, cfg RateLimitConfig) (bool, int, int, error) {
	periodSeconds := 1
	limit := int(math.Max(math.Ceil(cfg.RPS), float64(cfg.Burst)))
	if limit < 1 {
		limit = 1
	}

	res, err := redisRateScript.Run(ctx, client, []string{key}, periodSeconds).Result()
	if err != nil {
		return true, 0, 0, err
	}

	values, ok := res.([]interface{})
	if !ok || len(values) < 2 {
		return true, 0, 0, nil
	}

	current, _ := values[0].(int64)
	ttl, _ := values[1].(int64)

	remaining := limit - int(current)
	if remaining < 0 {
		remaining = 0
	}
	reset := int(ttl)
	if reset < 0 {
		reset = 0
	}

	return int(current) <= limit, remaining, reset, nil
}
