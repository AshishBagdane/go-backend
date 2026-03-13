package cache

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client  *redis.Client
	enabled bool
}

var ctx = context.Background()

func NewRedisCache(enabled bool, addr string) *RedisCache {
	if !enabled {
		return &RedisCache{enabled: false}
	}

	if addr == "" {
		addr = "localhost:6379"
	}

	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	return &RedisCache{client: client, enabled: true}
}

func (r *RedisCache) Get(key string) (string, error) {
	if !r.enabled {
		return "", errors.New("redis disabled")
	}
	return r.client.Get(ctx, key).Result()
}

func (r *RedisCache) Set(key string, value string) {
	if !r.enabled {
		return
	}
	r.client.Set(ctx, key, value, time.Minute*10)
}

func (r *RedisCache) Delete(key string) {
	if !r.enabled {
		return
	}
	r.client.Del(ctx, key)
}

func (r *RedisCache) Ping() error {
	if !r.enabled {
		return nil
	}
	return r.client.Ping(ctx).Err()
}

func (r *RedisCache) Client() *redis.Client {
	if !r.enabled {
		return nil
	}
	return r.client
}
