package externalHelpers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/services/cache"
	"github.com/nateshr/likeminds-swarm/internal/services/logging"
	"github.com/nateshr/likeminds-swarm/internal/utils"
)

type FetchBotUserResponse struct {
	Success bool       `json:"success"`
	BotUser MemberMeta `json:"user"`
}

func getCommunityIdAgainstApiKeyFromCache(cacheHelper cache.Helper, apiKey string,
) int {

	communityId := DefaultCommunityId

	cacheKey := fmt.Sprintf(cache.CommunityIdAgainstApiKeyCacheKey, apiKey)
	value, exists, err := cacheHelper.GetWithKeyExists(cacheKey)
	if err != nil {
		logging.Error(fmt.Sprintf("error fetching community_id from cache for api-key: %s", apiKey))
		return communityId
	}

	if !exists {
		logging.Info(fmt.Sprintf("community_id not found in cache for api-key: %s", apiKey))
		return communityId
	}

	communityId, err = strconv.Atoi(string(value))
	if err != nil {
		logging.Error(fmt.Sprintf("error parsing community_id from cache %v", err))
	}

	return communityId
}

func setCommunityIdAgainstApiKeyInCache(cacheHelper cache.Helper, apiKey string, communityId int) {

	cacheKey := fmt.Sprintf(cache.CommunityIdAgainstApiKeyCacheKey, apiKey)
	if err := cacheHelper.Set(cacheKey, communityId, cache.CommunityIdAgainstApiKeyCacheTTL).Err(); err != nil {
		logging.Error(fmt.Sprintf("error setting community_id in cache for api-key: %s", apiKey))
	}
}

// Exposed Method to get Community ID from API Key
func GetCommunityId(c *gin.Context, cacheHelper cache.Helper) int {

	defer utils.Timer("GetCommunityId")()

	apiKey := c.GetHeader(utils.HeadersApiKey)

	// Fetch community_id from cache
	communityId := getCommunityIdAgainstApiKeyFromCache(cacheHelper, apiKey)
	if communityId != DefaultCommunityId {
		return communityId
	}

	// Fetch community_id from API
	respBytes, statusCode, err := GetRequestResponse(CaravanService, SdkAuthenticateEndPoint, GETRequest, CreateHeaders(c, ""), nil, nil)
	if respBytes == nil {
		//If API fails or any other error
		GeneralAPIError(c, err.Error())
		return DefaultCommunityId
	}

	// Validate response
	apiCR := ValidateClientResponse(c, respBytes, statusCode)
	if apiCR == nil {
		return DefaultCommunityId
	}

	// If flow succeeds
	communityId = int(apiCR.Response["community_id"].(float64))

	// Set community_id in cache
	utils.SafeGo(func() { setCommunityIdAgainstApiKeyInCache(cacheHelper, apiKey, communityId) })

	return communityId
}

// Exposed method to get community bot/owner id using API Key
func GetCommunityBotId(apiKey string, communityId string) string {

	headers := map[string]interface{}{
		"x-api-key":       apiKey,
		"x-platform-type": SwarmServiceHeader,
	}

	params := map[string]string{}

	if communityId != "" {
		params["community_id"] = communityId
	}

	// Send GET Request to api/sdk/user/bot
	respBytes, statusCode, err := GetRequestResponse(CaravanService, SdkBotUserEndpoint, GETRequest, headers, params, nil)
	if err != nil || statusCode != http.StatusOK {
		logging.Error("Error fetching bot id from API: ", err, " Response: ", string(respBytes))
		return ""
	}

	// Unmarshal the response
	response := FetchBotUserResponse{}
	if err := json.Unmarshal(respBytes, &response); err != nil {
		logging.Error("Error unmarshalling bot id from API: ", err)
		return ""
	}

	return response.BotUser.UserUniqueId
}
