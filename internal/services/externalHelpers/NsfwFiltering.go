package externalHelpers

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nateshr/likeminds-swarm/internal/services/cache"
	"github.com/nateshr/likeminds-swarm/internal/services/logging"
)

type InferdoNsfwApiResponse struct {
	Success   *bool    `json:"success,omitempty"`
	NsfwScore *float64 `json:"NSFW_Prob,omitempty"`
	Message   *string  `json:"message,omitempty"`
}

func GetNsfwScoreForImage(cacheHelper cache.Helper, userId string, communityId int,
	imageUrl string, inferdoApiKey string) (float64, error) {

	inferdoNsfwEndpoint := InferdoNsfwApiEndpoint

	headers := gin.H{
		"X-RapidAPI-Key":  inferdoApiKey,
		"X-RapidAPI-Host": InferdoApiHeaderHost,
	}

	body := gin.H{
		"url": imageUrl,
	}

	resp, statusCode, err := SendPostRequestToExternalService(inferdoNsfwEndpoint, headers, body)

	if err != nil || statusCode != 200 {

		parsedResponse := string(resp)

		// Log error
		logging.Error(fmt.Sprintf("NSFW Filtering | Error while sending request to Inferdo api: statusCode: %d %s",
			statusCode, parsedResponse))

		// Increment key in redis cache with expiry of 1 day
		count, err := cacheHelper.IncrWithExpiry(cache.InferdoApiFailsCountKey, 24*time.Hour)
		if err != nil {
			logging.Error(fmt.Sprintf("NSFW Filtering | Error while incrementing inferdo api fails count: %s", err.Error()))
		} else if count == 10 {

			//Send mail to team notifying about inferdo api errors
			sendMailtoTeamForInferdoAPIErrors(userId, communityId, parsedResponse)
		}

		return -1, err
	}

	var response InferdoNsfwApiResponse
	err = json.Unmarshal(resp, &response)
	if err != nil {
		logging.Error(fmt.Sprintf("NSFW Filtering | Error while unmarshalling inferdo api response: %s", err.Error()))
		return -1, err
	}

	if response.NsfwScore == nil {
		return -1, fmt.Errorf("error while getting nsfw score from inferdo api")
	}

	return *response.NsfwScore, nil
}

func sendMailtoTeamForInferdoAPIErrors(userId string, communityId int, errorResponse string) {

	// Get community name
	// communityName, err := GetCommunityName(userId, communityId)

	mails := []string{"abc"}
	subject := "Multiple Inferdo API Errors for community: "
	body := fmt.Sprintf("Inferdo API Error: %s", errorResponse)

	// Send mail using caravan service
	SendMail(userId, mails, subject, body)

}
