package externalHelpers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/services/cache"
	"github.com/nateshr/likeminds-swarm/internal/services/logging"
	"github.com/nateshr/likeminds-swarm/internal/utils"
)

type UserCommunityFollowedChannels struct {
	ID int `json:"id"  binding:"required"`
}

type UserCommunityChannelsResponse struct {
	Success               bool                            `json:"success"`
	UserCommunityChannels []UserCommunityFollowedChannels `json:"chatrooms_data"`
}

func fetchUserCommunityChannelsFromCache(cacheHelper cache.Helper, userId string, communityId int) []int {

	userCommunityChannelsCacheKey := fmt.Sprintf(cache.UserCommunityChannelsCacheKey, userId, communityId)
	userCommunityChannelsCacheValue := cacheHelper.Get(userCommunityChannelsCacheKey)

	if userCommunityChannelsCacheValue.Val() == "" || userCommunityChannelsCacheValue.Val() == "null" {
		logging.Error(fmt.Sprintf("Channels not found in cache for communityId: %d, userId: %s", communityId, userId))
		return nil
	}

	var userCommunityChannels []int
	err := json.Unmarshal([]byte(userCommunityChannelsCacheValue.Val()), &userCommunityChannels)
	if err != nil {
		return nil
	}

	return userCommunityChannels
}

func saveUserCommunityChannelsInCache(cacheHelper cache.Helper, userId string, communityId int, userCommunityChannels []int) error {

	userCommunityChannelsCacheKey := fmt.Sprintf(cache.UserCommunityChannelsCacheKey, userId, communityId)
	parsedUserCommunityChannels, err := json.Marshal(userCommunityChannels)
	if err != nil {
		return err
	}

	userCommunityChannelsSet := cacheHelper.Set(userCommunityChannelsCacheKey, parsedUserCommunityChannels, cache.UserCommunityChannelsCacheTTLInHours*time.Hour)
	if userCommunityChannelsSet.Err() != nil {
		logging.Error(fmt.Sprintf("Error while saving user community channels in cache for communityId: %d, userId: %s, err: %v", communityId, userId, err))
		return err
	}

	logging.Info(fmt.Sprintf("Saved user community channels in cache for communityId: %d, userId: %s", communityId, userId))
	return nil
}

// fetch user community channels for application internal use
func getUserCommunityChannelsFromAPI(userId string, apiKey string) ([]UserCommunityFollowedChannels, error) {

	headers := gin.H{
		"Content-Type":        "application/json",
		utils.HeadersMemberId: userId,
		utils.HeadersApiKey:   apiKey,
	}

	// Params to be sent in the api/sync/chatrooms
	params := map[string]string{
		ParamsIsLocalDb:        fmt.Sprintf("%t", true),
		ParamsMinimumTimestamp: fmt.Sprintf("%d", 0),
		ParamsMaximumTimestamp: fmt.Sprintf("%d", time.Now().Unix()),
		ParamsChatroomTypes:    fmt.Sprintf("[%d]", FeedroomChatroomType),
	}

	// Send request to internal service
	respBytes, statusCode, err := GetRequestResponse(CaravanService, SyncChatroomsEndpoint, GETRequest, headers, params, nil)
	if err != nil || statusCode != http.StatusOK {
		return nil, err
	}

	// Parse response
	var uccr UserCommunityChannelsResponse
	if err = json.Unmarshal(respBytes, &uccr); err != nil {
		return nil, err
	}

	if len(uccr.UserCommunityChannels) == 0 {
		return nil, errors.New("User community channels not found.")
	}

	return uccr.UserCommunityChannels, nil
}

// Method to fetch user community channels
func FetchUserCommunityChannels(cacheHelper cache.Helper, userId string, communityId int, apiKey string) ([]int, error) {

	// fetch user community channels from cache
	userCommunityChannels := fetchUserCommunityChannelsFromCache(cacheHelper, userId, communityId)
	if userCommunityChannels == nil {

		// fetch from api if not found in cache
		userCommunityChannels, err := getUserCommunityChannelsFromAPI(userId, apiKey)
		if err != nil {
			return nil, err
		}

		// Save the channel ids in list
		channelIdsList := []int{}

		for _, userCommunityChannel := range userCommunityChannels {
			channelIdsList = append(channelIdsList, userCommunityChannel.ID)
		}

		// save in cache
		go saveUserCommunityChannelsInCache(cacheHelper, userId, communityId, channelIdsList)
	}

	return userCommunityChannels, nil
}
