package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/api/constants"
	"github.com/nateshr/likeminds-swarm/internal/api/requests"
	"github.com/nateshr/likeminds-swarm/internal/services/cache"
	"github.com/nateshr/likeminds-swarm/internal/services/externalHelpers"
	log "github.com/nateshr/likeminds-swarm/internal/services/logging"
	"github.com/nateshr/likeminds-swarm/internal/utils"
)

// Internal Method to create cache key name for user connection
func getUserConnectionCacheKeyName(userId string, communityId int) string {
	return fmt.Sprintf(cache.UserConnnectionCacheKey, userId, utils.ConvertNumberToString(communityId))
}

// Internal Method to get User connection data from cache
func getUserConnectionDataFromCache(handlers *FeedHandlers, userId string, communityId int) (map[string]bool, bool) {
	userConnectionData := map[string]bool{}

	// Getting data from cache for user
	userCacheKeyName := getUserConnectionCacheKeyName(userId, communityId)
	// get := handlers.cacheHelper.Get(userCacheKeyName)

	val, keyExists, err := handlers.cacheHelper.GetWithKeyExists(userCacheKeyName)
	if err != nil {
		log.Error(fmt.Sprintf("getUserConnectionDataFromCache() - Error while getting data from cache, %s: %s", userCacheKeyName, err.Error()))
		return userConnectionData, keyExists
	}

	if keyExists {
		err = json.Unmarshal([]byte(val), &userConnectionData)
		if err != nil {
			return userConnectionData, keyExists
		}
	}

	return userConnectionData, keyExists
}

// Internal Method to set User connection data in cache
func setUserConnectionDataInCache(handlers *FeedHandlers, userId string, communityId int, connectionData interface{}) {
	userCacheKeyName := getUserConnectionCacheKeyName(userId, communityId)
	marshalledData, _ := json.Marshal(connectionData)
	set := handlers.cacheHelper.Set(userCacheKeyName, marshalledData, 24*time.Hour)
	if set.Err() != nil {
		log.Error(fmt.Sprintf("setUserConnectionDataInCache() - Error while setting data in cache, %s: %s", userCacheKeyName, set.Err()))
	}
}

// Interal Method to Add user in connection list for a user
func addUserInConnectionListInCache(handlers *FeedHandlers, primaryUserId string, communityId int, secondaryUserId string) {
	primaryUserConnectionData, _ := getUserConnectionDataFromCache(handlers, primaryUserId, communityId)
	secondaryUserConnectionData, _ := getUserConnectionDataFromCache(handlers, secondaryUserId, communityId)

	primaryUserConnectionData[secondaryUserId] = true
	secondaryUserConnectionData[primaryUserId] = true

	setUserConnectionDataInCache(handlers, primaryUserId, communityId, primaryUserConnectionData)
	setUserConnectionDataInCache(handlers, secondaryUserId, communityId, secondaryUserConnectionData)
}

// Internal Method to Remove user from connection list for a user
func removeUserInConnectionListInCache(handlers *FeedHandlers, primaryUserId string, communityId int, secondaryUserId string) {
	primaryUserConnectionData, _ := getUserConnectionDataFromCache(handlers, primaryUserId, communityId)
	secondaryUserConnectionData, _ := getUserConnectionDataFromCache(handlers, secondaryUserId, communityId)

	delete(primaryUserConnectionData, secondaryUserId)
	delete(secondaryUserConnectionData, primaryUserId)

	setUserConnectionDataInCache(handlers, primaryUserId, communityId, primaryUserConnectionData)
	setUserConnectionDataInCache(handlers, secondaryUserId, communityId, secondaryUserConnectionData)
}

// Internal Method to warm up connection list for a user
func warmUpConnectionList(handlers *FeedHandlers, userId string, communityId int) {
	userCacheKeyName := getUserConnectionCacheKeyName(userId, communityId)
	handlers.cacheHelper.Del(userCacheKeyName)

	getConnections := true
	page := 1
	pageSize := 50

	for getConnections {
		success, userConnections := externalHelpers.FetchUserConnectionsByPage(userId, communityId, page, pageSize)

		if success {
			if len(userConnections.Connections) > 0 {
				for _, userConnection := range userConnections.Connections {
					addUserInConnectionListInCache(handlers, userConnection.User1UUID, communityId, userConnection.User2UUID)
				}

				page += 1
			} else {
				getConnections = false
			}
		} else {
			getConnections = false
		}
	}
}

// Internal Method to Update connection list for a user
func updateConnectionList(handlers *FeedHandlers, primaryUserId string, communityId int, secondaryUserId string, connected bool) {
	_, primaryUserCacheKeyExists := getUserConnectionDataFromCache(handlers, primaryUserId, communityId)
	if !primaryUserCacheKeyExists {
		warmUpConnectionList(handlers, primaryUserId, communityId)
	}

	if secondaryUserId != "" {
		_, secondaryUserCacheKeyExists := getUserConnectionDataFromCache(handlers, secondaryUserId, communityId)
		if !secondaryUserCacheKeyExists {
			warmUpConnectionList(handlers, secondaryUserId, communityId)
		} else {
			if connected {
				addUserInConnectionListInCache(handlers, primaryUserId, communityId, secondaryUserId)
			} else {
				removeUserInConnectionListInCache(handlers, primaryUserId, communityId, secondaryUserId)
			}
		}
	}
}

// Exposed Method to Update a Connection
func (handlers *FeedHandlers) UpdateConnection(c *gin.Context) {
	// fetch headers and url params
	headers := utils.GetHeaders(c)
	userId := c.Param("user_id")

	// validation of api_key
	communityId := externalHelpers.GetCommunityId(c)
	if communityId == externalHelpers.DefaultCommunityId {
		return
	}

	// validation of request body
	var updateConnectionRequest requests.UpdateConnectionRequest
	if err := c.ShouldBindJSON(&updateConnectionRequest); err != nil {
		utils.GeneralAPIValidationError(c, err.Error())
		return
	}

	if updateConnectionRequest.Status == constants.CONNECTION_CONNECTED {
		updateConnectionList(handlers, headers[utils.HeadersMemberId], communityId, userId, true)
	} else {
		updateConnectionList(handlers, headers[utils.HeadersMemberId], communityId, userId, false)
	}

	// return final response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}
