package externalHelpers

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/services/logging"
	"github.com/nateshr/likeminds-swarm/internal/utils"
)

type FetchBotUserResponse struct {
	Success bool       `json:"success"`
	BotUser MemberMeta `json:"user"`
}

// Exposed Method to get Community ID from API Key
func GetCommunityId(c *gin.Context) int {

	defer utils.Timer("GetCommunityId")()

	//Send Request
	respBytes, statusCode, err := GetRequestResponse(CaravanService, SdkAuthenticateEndPoint, GETRequest, CreateHeaders(c, ""), nil, nil)
	if respBytes == nil {
		//If API fails or any other error
		GeneralAPIError(c, err.Error())
		return DefaultCommunityId
	}

	//Validate response
	apiCR := ValidateClientResponse(c, respBytes, statusCode)
	if apiCR == nil {
		return DefaultCommunityId
	}

	//If flow succeeds
	dataResponse := apiCR.Response

	return int(dataResponse["community_id"].(float64))
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
