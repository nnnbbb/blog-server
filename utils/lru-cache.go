package utils

import (
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
)

type CacheItem[T any] struct {
	Data      T
	ExpiresAt time.Time
}

type Cache[T any] struct {
	store *lru.Cache[string, CacheItem[T]]
	ttl   time.Duration
}

// NewCache 新建一个带过期时间的缓存
// size: 最大容量
// ttl: 过期时间
func NewCache[T any](size int, ttl time.Duration) (*Cache[T], error) {
	l, err := lru.New[string, CacheItem[T]](size)
	if err != nil {
		return nil, err
	}
	return &Cache[T]{
		store: l,
		ttl:   ttl,
	}, nil
}

// Set 写入缓存
func (c *Cache[T]) Set(key string, value T) {
	c.store.Add(key, CacheItem[T]{
		Data:      value,
		ExpiresAt: time.Now().Add(c.ttl),
	})
}

// Get 读取缓存（未命中或过期返回 false）
func (c *Cache[T]) Get(key string) (T, bool) {
	if val, ok := c.store.Get(key); ok {
		if time.Now().Before(val.ExpiresAt) {
			return val.Data, true
		}
		// 过期删除
		c.store.Remove(key)
	}
	var zero T
	return zero, false
}

// Remove 删除缓存
func (c *Cache[T]) Remove(key string) {
	c.store.Remove(key)
}
