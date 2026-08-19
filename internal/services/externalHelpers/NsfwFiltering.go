package externalHelpers

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/LikeMindsCommunity/likeminds-swarm/internal/services/cache"
	"github.com/LikeMindsCommunity/likeminds-swarm/internal/services/environment"
	"github.com/LikeMindsCommunity/likeminds-swarm/internal/services/logging"
	"github.com/LikeMindsCommunity/likeminds-swarm/internal/utils"
	"github.com/gin-gonic/gin"
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
		inferdoFailureCacheKey := fmt.Sprintf(cache.InferdoApiFailsCountKey, communityId)

		// Log error
		logging.Error(fmt.Sprintf("NSFW Filtering | Error while sending request to Inferdo api: statusCode: %d %s",
			statusCode, parsedResponse))

		// Increment key in redis cache
		count, err := cacheHelper.Increment(inferdoFailureCacheKey)
		if err != nil {
			logging.Error(fmt.Sprintf("NSFW Filtering | Error while incrementing inferdo api fails count: %s", err.Error()))

		} else if count == 1 {

			// Set expiry for key to 1 day if first error
			_, err := cacheHelper.Expire(inferdoFailureCacheKey, time.Hour*24)
			if err != nil {
				logging.Error(fmt.Sprintf("NSFW Filtering | Error while setting expiry for inferdo api fails count: %s", err.Error()))
			}

		} else if count == 10 {

			// Send mail to team notifying about inferdo api errors [In background]
			utils.SafeGo(func() { sendMailtoTeamForInferdoAPIErrors(userId, communityId, parsedResponse) })
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

	subject := fmt.Sprintf("Multiple Inferdo API Errors for communityId: %d", communityId)
	body := fmt.Sprintf(`
						<h1>Multiple Inferdo API errors occured for CommunityId %d</h1>
						<h2>Date And time: %v </h2>
						<h2>Error Response: </h2>
						%s
						<br>
						<h3>Please contact team & check logs for more info<h3>`,
		communityId, time.Now().String(), errorResponse)

	teamMails := environment.GoDotEnvVariable("TEAM_ADMIN_MAILS")
	mails := strings.Split(teamMails, ",")

	// Send mail using caravan service
	SendMail(userId, mails, subject, body)
}
