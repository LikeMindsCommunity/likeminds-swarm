package cache

import (
	"time"

	"github.com/go-redis/redis/v7"
)

// Helper | interface for CacheHelper
type Helper interface {
	Set(key string, object interface{}, expiration time.Duration) *redis.StatusCmd
	Get(key string) *redis.StringCmd
	DelMultiple(keys []string) []*redis.IntCmd
	Del(key string) *redis.IntCmd
	LPush(key string, object string, listMaxLength int) *redis.IntCmd
	LRem(key string, count int64, element interface{}) *redis.IntCmd
}
