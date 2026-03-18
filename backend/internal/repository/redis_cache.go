package repository

import (
	"context"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(client *redis.Client) *RedisCache {
	return &RedisCache{client: client}
}

func (c *RedisCache) Get(key string) ([]byte, bool) {
	value, err := c.client.Get(context.Background(), key).Bytes()
	if err != nil {
		return nil, false
	}
	return value, true
}

func (c *RedisCache) Set(key string, value []byte, ttl time.Duration) {
	_ = c.client.Set(context.Background(), key, value, ttl).Err()
}

func (c *RedisCache) Delete(key string) {
	_ = c.client.Del(context.Background(), key).Err()
}

func (c *RedisCache) DeleteByPrefix(prefix string) {
	var cursor uint64
	pattern := prefix
	if !strings.HasSuffix(pattern, "*") {
		pattern += "*"
	}

	for {
		keys, nextCursor, err := c.client.Scan(context.Background(), cursor, pattern, 100).Result()
		if err != nil {
			return
		}
		if len(keys) > 0 {
			_ = c.client.Del(context.Background(), keys...).Err()
		}
		if nextCursor == 0 {
			return
		}
		cursor = nextCursor
	}
}

func (c *RedisCache) AcquireLock(key string, ttl time.Duration) bool {
	locked, err := c.client.SetNX(context.Background(), key, "1", ttl).Result()
	if err != nil {
		return false
	}
	return locked
}

func (c *RedisCache) ReleaseLock(key string) {
	_ = c.client.Del(context.Background(), key).Err()
}
