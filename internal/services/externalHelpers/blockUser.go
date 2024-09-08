package externalHelpers

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/services/cache"
	"github.com/nateshr/likeminds-swarm/internal/services/logging"
)

type BlockUserResponse struct {
	BlockedUsers  []MemberMeta `json:"blocked_users"`
	BlockingUsers []MemberMeta `json:"blocking_users"`
}

type BlockUserCache struct {
	BlockingUsers []string `json:"blocking_users"`
	BlockedUsers  []string `json:"blocked_users"`
}

// Internal method to fetch block user list against userId from caravan service
func GetUserBlockList(cacheHelper cache.Helper, userId string, communityId int) (*BlockUserCache, error) {
	var blockUserResponse BlockUserResponse
	blockUserValue := BlockUserCache{
		BlockingUsers: []string{},
		BlockedUsers:  []string{},
	}

	// Get data from cache
	blockUserCacheKey := fmt.Sprintf(cache.BlockedUsersCacheKey, communityId, userId)
	blockUserCacheValue := cacheHelper.Get(blockUserCacheKey)

	if blockUserCacheValue.Val() != "" && blockUserCacheValue.Val() != "null" {
		err := json.Unmarshal([]byte(blockUserCacheValue.Val()), &blockUserValue)

		if err != nil {
			return nil, err
		}

		return &blockUserValue, nil
	}

	headers := gin.H{
		"Content-Type":    "application/json",
		"x-platform-type": SwarmServiceHeader,
		"x-member-id":     userId,
	}

	// Params to be sent in the api/community/configurations
	params := map[string]string{
		ParamBlockUserType: fmt.Sprintf("[%s, %s]", BlockingUserType, BlockedUserType),
		ParamCommunityId:   fmt.Sprintf("%d", communityId),
		ParamPage:          fmt.Sprintf("%d", DefaultGetBlockUserPageValue),
		ParamPageSize:      fmt.Sprintf("%d", DefaultGetBlockUserPageSizeValue),
	}

	apiEndpoint := fmt.Sprintf(BlockUserEndpoint, userId)

	//Send Request
	respBytes, _, err := GetRequestResponse(CaravanService, apiEndpoint, GETRequest, headers, params, nil)

	if respBytes == nil {
		//If API fails or any other error
		return nil, err
	}

	if err := json.Unmarshal(respBytes, &blockUserResponse); err != nil {
		//Internal unmarshal error
		return nil, err
	}

	for _, blockingUser := range blockUserResponse.BlockingUsers {
		blockUserValue.BlockingUsers = append(blockUserValue.BlockingUsers, blockingUser.UUID)
	}

	for _, blockedUser := range blockUserResponse.BlockedUsers {
		blockUserValue.BlockedUsers = append(blockUserValue.BlockedUsers, blockedUser.UUID)
	}

	// Save data to cache
	logging.Info(fmt.Sprintf("Saving the block users list for user: %s in cache for %d", userId, communityId))
	blockUserCacheBytesValue, _ := json.Marshal(blockUserValue)
	cacheHelper.Set(blockUserCacheKey, blockUserCacheBytesValue, cache.BlockUserCacheTTLInHours*time.Hour)

	return &blockUserValue, nil
}
