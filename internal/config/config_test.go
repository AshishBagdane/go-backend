package config

import "testing"

func TestValidateRequiresAuthKey(t *testing.T) {
	cfg := Config{
		Env: "test",
		Server: ServerConfig{
			Address: ":8080",
		},
		DB: DBConfig{
			Driver: "sqlite3",
			DSN:    ":memory:",
		},
		Redis: RedisConfig{
			Enabled: false,
		},
		Rate: RateConfig{
			Default: RateLimitConfig{RPS: 1, Burst: 1},
		},
		Auth: AuthConfig{
			Enabled: true,
			APIKey:  "",
		},
	}

	if err := cfg.Validate(); err == nil {
		t.Fatalf("expected validation error when auth is enabled without api_key")
	}
}
