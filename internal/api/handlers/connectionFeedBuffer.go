package handlers

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/services/cache"
	log "github.com/nateshr/likeminds-swarm/internal/services/logging"
	"github.com/nateshr/likeminds-swarm/internal/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Internal Method to create cache key name for connection feed buffer
func getConnectionFeedBufferCacheKeyName(userId string, communityId int) string {
	return fmt.Sprintf(cache.ConnectionFeedBufferCacheKey, userId, utils.ConvertNumberToString(communityId))
}

// Internal Method to get User Connection Feed Buffer data from cache
func getConnectionFeedBufferDataFromCache(handlers *FeedHandlers, userId string, communityId int) (map[string]bool, bool) {
	connectionFeedBufferData := map[string]bool{}

	// Getting data from cache for user
	userConnectionFeedCacheKeyName := getConnectionFeedBufferCacheKeyName(userId, communityId)

	val, keyExists, err := handlers.cacheHelper.GetWithKeyExists(userConnectionFeedCacheKeyName)
	if err != nil {
		log.Error(fmt.Sprintf("getConnectionFeedBufferDataFromCache() - Error while conversion of data from cache, %s: %s", userConnectionFeedCacheKeyName, err.Error()))
	}

	// If key exists in cache for the User
	if keyExists {
		err = json.Unmarshal([]byte(val), &connectionFeedBufferData)
		if err != nil {
			log.Error(fmt.Sprintf("getConnectionFeedBufferDataFromCache() - Error while getting data conversion, %s: %s", userConnectionFeedCacheKeyName, err.Error()))
			return connectionFeedBufferData, keyExists
		}
	}

	return connectionFeedBufferData, keyExists
}

// Internal Method to set User Connection Feed Buffer data in cache
func setConnectionFeedBufferDataInCache(handlers *FeedHandlers, userId string, communityId int, connectionFeedBufferData interface{}) {
	userConnectionFeedCacheKeyName := getConnectionFeedBufferCacheKeyName(userId, communityId)
	marshalledData, _ := json.Marshal(connectionFeedBufferData)
	set := handlers.cacheHelper.Set(userConnectionFeedCacheKeyName, marshalledData, 24*time.Hour)
	if set.Err() != nil {
		log.Error(fmt.Sprintf("setConnectionFeedBufferDataInCache() - Error while setting data in cache, %s: %s", userConnectionFeedCacheKeyName, set.Err()))
	}
}

// Interal Method to Add post in Connection Feed Buffer list for a user
func addPostInConnectionFeedBufferInCache(handlers *FeedHandlers, userId string, communityId int, postId string) {
	connectionFeedBufferData, _ := getConnectionFeedBufferDataFromCache(handlers, userId, communityId)

	connectionFeedBufferData[postId] = true

	setConnectionFeedBufferDataInCache(handlers, userId, communityId, connectionFeedBufferData)
}

// Internal Method to Remove post from Connection Feed Buffer list for a user
func removePostInConnectionFeedBufferInCache(handlers *FeedHandlers, userId string, communityId int, postId string) {
	connectionFeedBufferData, _ := getConnectionFeedBufferDataFromCache(handlers, userId, communityId)

	delete(connectionFeedBufferData, postId)

	setConnectionFeedBufferDataInCache(handlers, userId, communityId, connectionFeedBufferData)
}

// Internal Method to warm up Connection Feed Buffer list for a user
func warmUpConnectionFeedBuffer(handlers *FeedHandlers, userId string, communityId int) {
	userConnectionFeedCacheKeyName := getConnectionFeedBufferCacheKeyName(userId, communityId)
	handlers.cacheHelper.Del(userConnectionFeedCacheKeyName)

	userConnectionData, _ := getUserConnectionDataFromCache(handlers, userId, communityId)
	if len(userConnectionData) == 0 {
		updateConnectionList(handlers, userId, communityId, "", false)
	}

	userConnectionFeedFilter := gin.H{
		"community_id": communityId,
		"user_id":      userId,
	}

	userConnectionFeedResults, err := handlers.connectionFeedHelper.FindConnectionFeedHelper(userConnectionFeedFilter, gin.H{})
	if err != nil {
		log.Error(fmt.Sprintf("warmUpConnectionFeedBuffer() - Error while fetching connection feed data from DB, %s", err.Error()))
		return
	}

	existingPostIds := []primitive.ObjectID{}
	for _, userConnectionFeedResult := range userConnectionFeedResults {
		existingPostIds = append(existingPostIds, userConnectionFeedResult.PostId)
	}

	userConnectionData, _ = getUserConnectionDataFromCache(handlers, userId, communityId)
	userConnectionIds := []string{}

	for userConnectionId := range userConnectionData {
		userConnectionIds = append(userConnectionIds, userConnectionId)
	}

	userPostsFilter := gin.H{
		"is_deleted":   false,
		"community_id": communityId,
		"user_id": gin.H{
			"$in": userConnectionIds,
		},
		"post_id": gin.H{
			"$nin": existingPostIds,
		},
	}

	userPostFilterOptions := addSortingOptions(map[string]interface{}{}, "created_at", OrderTypeDescending)

	// fetch user posts using helper method
	userPostResults, err := handlers.postHelper.FindPostHelper(userPostsFilter, userPostFilterOptions)
	if err != nil {
		log.Error(fmt.Sprintf("warmUpConnectionFeedBuffer() - Error while fetching post data from DB, %s", err.Error()))
		return
	}

	for _, userPost := range userPostResults {
		addPostInConnectionFeedBufferInCache(handlers, userId, communityId, userPost.ID.Hex())
	}
}

// Internal Method to Update Connection Feed Buffer list for a user
func updateConnectionFeedBuffer(handlers *FeedHandlers, userId string, communityId int, postId string, add bool) {
	_, connectionFeedBufferCacheKeyExists := getConnectionFeedBufferDataFromCache(handlers, userId, communityId)
	if !connectionFeedBufferCacheKeyExists {
		warmUpConnectionFeedBuffer(handlers, userId, communityId)
	} else {
		if postId != "" && add {
			addPostInConnectionFeedBufferInCache(handlers, userId, communityId, postId)
		} else if postId != "" && !add {
			removePostInConnectionFeedBufferInCache(handlers, userId, communityId, postId)
		}
	}
}
