package cache

import (
	"time"

	"github.com/go-redis/redis/v7"
)

// LPush | left push into a cache list, trims list if listMaxLength is valid
func (cacheHelper *cacheHelper) LPush(key string, object string, listMaxLength int) *redis.IntCmd {
	intCMD := cacheHelper.redisClient.LPush(key, object)

	if listMaxLength != -1 {
		cacheHelper.redisClient.LTrim(key, 0, int64(listMaxLength))
	}
	return intCMD
}

// Set | set the key with object value into cache storage, set expiration = 0 for no expiry
func (cacheHelper *cacheHelper) Set(key string, object interface{}, expiration time.Duration) *redis.StatusCmd {
	statusCMD := cacheHelper.redisClient.Set(key, object, expiration)
	return statusCMD
}

// Get | get the key object value from cache storage
func (cacheHelper *cacheHelper) Get(key string) *redis.StringCmd {
	stringCMD := cacheHelper.redisClient.Get(key)
	return stringCMD
}

// DeleteMultiple | delete muktiple keys in the list from cache storage
func (cacheHelper *cacheHelper) DelMultiple(keys []string) []*redis.IntCmd {
	intCMDs := [](*redis.IntCmd){}
	for _, key := range keys {
		intCMD := cacheHelper.redisClient.Del(key)
		intCMDs = append(intCMDs, intCMD)
	}
	return intCMDs
}

// Delete | delete key from the cache storage
func (cacheHelper *cacheHelper) Del(key string) *redis.IntCmd {
	intCMd := cacheHelper.redisClient.Del(key)
	return intCMd
}

// LRem | remove first count occurence from the element
func (cacheHelper *cacheHelper) LRem(key string, count int64, element interface{}) *redis.IntCmd {
	intCMD := cacheHelper.redisClient.LRem(key, count, element)
	return intCMD
}

// Structure for ElasticSearch Helper
type cacheHelper struct {
	redisClient *redis.Client
}

// NewCacheHelper | method to create and expose cache helper
func NewCacheHelper(redisCache *redis.Client) Helper {
	return &cacheHelper{
		redisClient: redisCache,
	}
}
