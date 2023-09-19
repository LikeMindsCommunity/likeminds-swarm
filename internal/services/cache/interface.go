package cache

import (
	"time"

	"github.com/go-redis/redis/v7"
)

// Helper | interface for CacheHelper
type Helper interface {
	Set(key string, object interface{}, expiration time.Duration) *redis.StatusCmd
	Get(key string) *redis.StringCmd
	DeleteMultiple(keys []string)
	Delete(key string)
	LPush(key string, object string, listMaxLength int)
	LRem(key string, count int64, element interface{}) *redis.IntCmd
}
