package externalHelpers

import (
	"encoding/json"
	"fmt"
	"time"

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

	inferdoNsfwEndpoint := "https://api.inferdo.com/v1/nsfw"

	headers := map[string]interface{}{
		"X-RapidAPI-Key":  inferdoApiKey,
		"X-RapidAPI-Host": "api.inferdo.com",
	}

	body := map[string]string{
		"image_url": imageUrl,
	}

	resp, statusCode, err := SendPostRequestToExternalService(inferdoNsfwEndpoint, headers, body)
	if err != nil || statusCode != 200 {

		parsedResponse := string(resp)

		// if resp != nil {
		// 	err = json.Unmarshal(resp, &parsedResponse)
		// }

		// Log error
		logging.Error(fmt.Sprintf("NSFW Filtering | Error while sending request to Inferdo api: statusCode: %d %s",
			statusCode, parsedResponse))

		// Increment key in redis cache with expiry of 1 day
		count := cacheHelper.IncrWithExpiry(cache.InferdoApiFailsCountKey, 24*time.Hour)
		if count == 10 {
			//Send mail to team if count is 10
			sendMailToTeam(userId, communityId, "Inferdo API Error", fmt.Sprintf("Inferdo API Error: %s", parsedResponse))
		}

		return -1, err
	}

	var inr InferdoNsfwApiResponse
	err = json.Unmarshal(resp, &inr)
	if err != nil {
		logging.Error(fmt.Sprintf("NSFW Filtering | Error while unmarshalling inferdo api response: %s", err.Error()))
		return -1, err
	}

	if inr.NsfwScore == nil {
		return -1, fmt.Errorf("error while getting nsfw score from inferdo api")
	}

	return *inr.NsfwScore, nil
}

func sendMailToTeam(userId string, communityId int, subject string, body string) {

	// team mails
	teamMails := []string{
		"product@likeminds.community",
		"backend@likeminds.community",
	}

	// send mail using external service api
	fmt.Printf("Sending mail to team: %s", teamMails)
	return
}
