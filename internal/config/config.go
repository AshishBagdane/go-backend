package config

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Env     string        `mapstructure:"env"`
	Server  ServerConfig  `mapstructure:"server"`
	DB      DBConfig      `mapstructure:"db"`
	Redis   RedisConfig   `mapstructure:"redis"`
	Rate    RateConfig    `mapstructure:"rate_limit"`
	CORS    CORSConfig    `mapstructure:"cors"`
	Auth    AuthConfig    `mapstructure:"auth"`
	Metrics MetricsConfig `mapstructure:"metrics"`
}

type ServerConfig struct {
	Address         string        `mapstructure:"address"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	IdleTimeout     time.Duration `mapstructure:"idle_timeout"`
	BodyLimitBytes  int64         `mapstructure:"body_limit_bytes"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

type DBConfig struct {
	Driver string `mapstructure:"driver"`
	DSN    string `mapstructure:"dsn"`
}

type RedisConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Addr    string `mapstructure:"addr"`
}

type RateConfig struct {
	Default     RateLimitConfig            `mapstructure:"default"`
	Routes      map[string]RateLimitConfig `mapstructure:"routes"`
	UseRedis    bool                       `mapstructure:"use_redis"`
	RedisPrefix string                     `mapstructure:"redis_prefix"`
}

type RateLimitConfig struct {
	RPS   float64 `mapstructure:"rps"`
	Burst int     `mapstructure:"burst"`
}

type CORSConfig struct {
	Enabled        bool     `mapstructure:"enabled"`
	AllowedOrigins []string `mapstructure:"allowed_origins"`
	AllowedMethods []string `mapstructure:"allowed_methods"`
	AllowedHeaders []string `mapstructure:"allowed_headers"`
	ExposeHeaders  []string `mapstructure:"expose_headers"`
	MaxAgeSeconds  int      `mapstructure:"max_age_seconds"`
}

type AuthConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	APIKey  string `mapstructure:"api_key"`
}

type MetricsConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Path    string `mapstructure:"path"`
}

func Load() (Config, error) {
	v := viper.New()

	setDefaults(v)

	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("/etc/idea-finder")

	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return Config{}, err
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return Config{}, err
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("env", "dev")
	v.SetDefault("server.address", ":8080")
	v.SetDefault("server.read_timeout", "5s")
	v.SetDefault("server.write_timeout", "10s")
	v.SetDefault("server.idle_timeout", "60s")
	v.SetDefault("server.body_limit_bytes", 1048576)
	v.SetDefault("server.shutdown_timeout", "10s")

	v.SetDefault("db.driver", "sqlite3")
	v.SetDefault("db.dsn", "./data/todo.db")

	v.SetDefault("redis.enabled", true)
	v.SetDefault("redis.addr", "localhost:6379")

	v.SetDefault("rate_limit.default.rps", 10)
	v.SetDefault("rate_limit.default.burst", 20)
	v.SetDefault("rate_limit.use_redis", false)
	v.SetDefault("rate_limit.redis_prefix", "rate_limit")

	v.SetDefault("cors.enabled", false)
	v.SetDefault("cors.allowed_origins", []string{"*"})
	v.SetDefault("cors.allowed_methods", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	v.SetDefault("cors.allowed_headers", []string{"Authorization", "Content-Type", "X-Request-ID"})
	v.SetDefault("cors.expose_headers", []string{"X-RateLimit-Limit", "X-RateLimit-Remaining", "X-RateLimit-Reset"})
	v.SetDefault("cors.max_age_seconds", 600)

	v.SetDefault("auth.enabled", false)
	v.SetDefault("auth.api_key", "")

	v.SetDefault("metrics.enabled", true)
	v.SetDefault("metrics.path", "/metrics")
}

func (c Config) Validate() error {
	var errs []string

	if c.Server.Address == "" {
		errs = append(errs, "server.address is required")
	}
	if c.DB.Driver == "" {
		errs = append(errs, "db.driver is required")
	}
	if c.DB.DSN == "" {
		errs = append(errs, "db.dsn is required")
	}
	if c.Auth.Enabled && c.Auth.APIKey == "" {
		errs = append(errs, "auth.api_key is required when auth.enabled=true")
	}
	if c.Rate.Default.RPS <= 0 || c.Rate.Default.Burst <= 0 {
		errs = append(errs, "rate_limit.default requires rps>0 and burst>0")
	}
	if c.Rate.UseRedis && !c.Redis.Enabled {
		errs = append(errs, "rate_limit.use_redis requires redis.enabled=true")
	}
	for route, rl := range c.Rate.Routes {
		if rl.RPS <= 0 || rl.Burst <= 0 {
			errs = append(errs, fmt.Sprintf("rate_limit.routes.%s requires rps>0 and burst>0", route))
		}
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}
