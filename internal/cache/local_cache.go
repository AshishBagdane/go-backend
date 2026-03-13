package cache

import (
	"github.com/dgraph-io/ristretto"
)

type LocalCache struct {
	cache *ristretto.Cache
}

func NewLocalCache() *LocalCache {

	cache, _ := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,
		MaxCost:     1 << 30,
		BufferItems: 64,
	})

	return &LocalCache{cache: cache}
}

func (c *LocalCache) Get(key string) (interface{}, bool) {
	return c.cache.Get(key)
}

func (c *LocalCache) Set(key string, value interface{}) {
	c.cache.Set(key, value, 1)
}

func (c *LocalCache) Delete(key string) {
	c.cache.Del(key)
}
