// cache/cache.go
package cache

import (
    "github.com/dgraph-io/ristretto"
    "log"
)

type GenericCache struct {
    cache *ristretto.Cache
}

func NewGenericCache(numCounters int64, maxCost int64, bufferItems int64) *GenericCache {
    cache, err := ristretto.NewCache(&ristretto.Config{
        NumCounters: numCounters, // number of keys to track frequency of (10M).
        MaxCost:     maxCost,     // maximum cost of cache (1GB).
        BufferItems: bufferItems, // number of keys per Get buffer.
    })
    if err != nil {
        log.Fatalf("failed to create cache: %v", err)
    }
    return &GenericCache{cache: cache}
}

func (c *GenericCache) Set(key string, value interface{}, cost int64) {
    c.cache.Set(key, value, cost)
    c.cache.Wait()
}

func (c *GenericCache) Get(key string) (interface{}, bool) {
    return c.cache.Get(key)
}

func (c *GenericCache) Delete(key string) {
    c.cache.Del(key)
}
