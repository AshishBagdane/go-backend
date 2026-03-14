package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AshishBagdane/go-backend/internal/api/handlers"
	"github.com/AshishBagdane/go-backend/internal/api/middleware"
	"github.com/AshishBagdane/go-backend/internal/cache"
	"github.com/AshishBagdane/go-backend/internal/config"
	"github.com/AshishBagdane/go-backend/internal/database"
	"github.com/AshishBagdane/go-backend/internal/repository"
	"github.com/AshishBagdane/go-backend/internal/service"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/AshishBagdane/go-backend/docs"
)

// @title Go Backend API
// @version 1.0
// @description Production-grade Go backend boilerplate built with Gin.
// @termsOfService https://example.com/terms

// @contact.name Go Backend Team
// @contact.email dev@example.com

// @license.name MIT

// @host localhost:8080
// @BasePath /
func main() {

	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	db := database.NewDB(cfg.DB.Driver, cfg.DB.DSN)

	repo := repository.NewTodoRepository(db)

	localCache := cache.NewLocalCache()

	redisCache := cache.NewRedisCache(cfg.Redis.Enabled, cfg.Redis.Addr)

	service := service.NewTodoService(repo, localCache, redisCache)

	handler := handlers.NewTodoHandler(service)
	healthHandler := handlers.NewHealthHandler(db, redisCache)

	router := gin.New()
	router.Use(gin.CustomRecovery(func(c *gin.Context, recovered any) {
		handlers.Respond[any](c, http.StatusInternalServerError, "internal server error", nil)
	}))
	router.Use(middleware.RequestID())
	router.Use(middleware.Logger(logger))
	router.Use(middleware.SecurityHeaders())
	router.Use(middleware.BodyLimit(cfg.Server.BodyLimitBytes))

	if cfg.CORS.Enabled {
		corsCfg := cors.Config{
			AllowOrigins:     cfg.CORS.AllowedOrigins,
			AllowMethods:     cfg.CORS.AllowedMethods,
			AllowHeaders:     cfg.CORS.AllowedHeaders,
			ExposeHeaders:    cfg.CORS.ExposeHeaders,
			MaxAge:           time.Duration(cfg.CORS.MaxAgeSeconds) * time.Second,
			AllowCredentials: false,
		}
		router.Use(cors.New(corsCfg))
	}

	if cfg.Metrics.Enabled {
		router.Use(middleware.MetricsMiddleware())
		router.GET(cfg.Metrics.Path, middleware.MetricsHandler())
	}

	routeLimits := map[string]middleware.RateLimitConfig{}
	for route, rl := range cfg.Rate.Routes {
		routeLimits[route] = middleware.RateLimitConfig{RPS: rl.RPS, Burst: rl.Burst}
	}
	if len(routeLimits) == 0 {
		routeLimits = map[string]middleware.RateLimitConfig{
			"GET /todos":     {RPS: 15, Burst: 30},
			"GET /todos/:id": {RPS: 15, Burst: 30},
			"POST /todos":    {RPS: 5, Burst: 10},
			"PUT /todos/:id": {RPS: 5, Burst: 10},
		}
	}

	router.Use(middleware.RateLimiter(middleware.RateLimiterConfig{
		Default:     middleware.RateLimitConfig{RPS: cfg.Rate.Default.RPS, Burst: cfg.Rate.Default.Burst},
		Routes:      routeLimits,
		UseRedis:    cfg.Rate.UseRedis,
		RedisClient: redisCache.Client(),
		RedisPrefix: cfg.Rate.RedisPrefix,
	}))

	router.GET("/health", healthHandler.Live)
	router.GET("/ready", healthHandler.Ready)

	router.NoRoute(func(c *gin.Context) {
		handlers.Respond[any](c, http.StatusNotFound, "route not found", nil)
	})

	if cfg.Env != "prod" {
		router.GET("/swagger-ui/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	api := router.Group("/")
	if cfg.Auth.Enabled {
		api.Use(middleware.APIKeyAuth(cfg.Auth.APIKey))
	}

	api.GET("/todos", handler.GetTodos)
	api.POST("/todos", handler.CreateTodo)
	api.GET("/todos/:id", handler.GetTodo)
	api.PUT("/todos/:id", handler.UpdateTodo)

	server := &http.Server{
		Addr:         cfg.Server.Address,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server failed", slog.Any("error", err))
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("server shutdown failed", slog.Any("error", err))
	}
}
