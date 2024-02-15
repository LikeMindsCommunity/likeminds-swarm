package externalHelpers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/nateshr/likeminds-swarm/internal/services/cache"
	"github.com/nateshr/likeminds-swarm/internal/services/environment"
	"github.com/nateshr/likeminds-swarm/internal/services/logging"
)

type CommunityWebhook struct {
	Id          int    `json:"id"`
	CommunityId int    `json:"community"`
	IsActive    bool   `json:"is_active"`
	Url         string `json:"url"`
	WebhookType string `json:"webhook_type"`
}

type CommunityWebhooksResponse struct {
	Success           bool               `json:"success"`
	CommunityWebhooks []CommunityWebhook `json:"webhooks"`
}

func fetchCommunityWebhooksFromInternalService(apiKey string) []CommunityWebhook {

	// Get Bot Id using ApiKey
	botId := GetCommunityBotId(apiKey)
	if botId == "" {
		return nil
	}

	headers := map[string]interface{}{
		"x-member-id":     botId,
		"x-api-key":       apiKey,
		"x-platform-type": SwarmServiceHeader,
	}

	// Send Request to fetch community webhooks
	respBytes, statusCode, err := GetRequestResponse(CaravanService, CommunityWebhooksEndpoint, GETRequest, headers, nil, nil)
	if err != nil || statusCode != http.StatusOK {
		//If API fails or any other error
		logging.Error("Error fetching community webhooks from API: ", err, " Response: ", string(respBytes))
		return nil
	}

	// Unmarshal the response
	response := CommunityWebhooksResponse{}
	if err := json.Unmarshal(respBytes, &response); err != nil {
		//Internal unmarshal error
		logging.Error("Error unmarshalling community webhooks from API: ", err)
		return nil
	}

	return response.CommunityWebhooks

}

// internal method to fetch community webhooks (from cache if present else internal service)
func fetchCommunityWebhooks(cacheHelper cache.Helper, apiKey string) []CommunityWebhook {

	communityWebhooks := []CommunityWebhook{}

	// Fetch communityWebhooks from cache
	webhooksCacheKey := fmt.Sprintf(cache.CommunityWebhooksCacheKey, apiKey)
	cacheValue, exists, err := cacheHelper.GetWithKeyExists(webhooksCacheKey)
	if err != nil {
		logging.Error("Error fetching community webhooks from cache: ", err)
		return communityWebhooks
	}

	// If webhooks exist in cache, unmarshal the value Else fetch from internal service
	if exists {
		if err := json.Unmarshal([]byte(cacheValue), &communityWebhooks); err != nil {
			logging.Error("Error unmarshalling community webhooks from cache: ", err)
			return communityWebhooks
		}
	} else {

		// Fetch community webhooks from internal service
		communityWebhooks = fetchCommunityWebhooksFromInternalService(apiKey)
		if communityWebhooks == nil {
			logging.Error("Error fetching community webhooks from internal service")
			return communityWebhooks
		}

		webhooksByteValues, err := json.Marshal(communityWebhooks)
		if err != nil {
			logging.Error("Error marshalling community webhooks for cache: ", err)
			return communityWebhooks
		}

		// Save data to cache
		logging.Info("Saving community webhooks to cache for ApiKey: ", apiKey)
		cacheHelper.Set(webhooksCacheKey, webhooksByteValues, CommunityWebhooksCacheTTTLInHours*time.Hour)
	}

	return communityWebhooks
}

// internal method to fetch webhook id
func fetchWebhookId(cacheHelper cache.Helper, apiKey string, webhookType string, webhookUrl string) int {

	communityWebhooks := fetchCommunityWebhooks(cacheHelper, apiKey)

	for _, webhook := range communityWebhooks {
		if webhook.WebhookType == webhookType && webhook.Url == webhookUrl {
			return webhook.Id
		}
	}

	return 0
}

// Exposed method to fetch active webhook urls for a specific webhook type
func FetchActiveWebhookUrls(cacheHelper cache.Helper, apiKey string, webhookType string) []string {

	communityWebhooks := fetchCommunityWebhooks(cacheHelper, apiKey)

	// Filter active webhooks for the webhook type
	activeWebhookUrls := []string{}
	for _, webhook := range communityWebhooks {
		if webhook.WebhookType == webhookType && webhook.IsActive {
			activeWebhookUrls = append(activeWebhookUrls, webhook.Url)
		}
	}

	return activeWebhookUrls
}

// Exposed method check if a webhook url is active
func IsWebhookUrlActive(cacheHelper cache.Helper, apiKey string, webhookType string, webhookUrl string) bool {

	communityWebhooks := fetchCommunityWebhooks(cacheHelper, apiKey)

	for _, webhook := range communityWebhooks {
		if webhook.WebhookType == webhookType && webhook.Url == webhookUrl && webhook.IsActive {
			return true
		}
	}

	return false
}

// Exposed method to disable a webhook and send mail to team
func DisableWebhookAndSendMail(cacheHelper cache.Helper, apiKey string, webhookType string, webhookUrl string,
	webhookResponse string, webhookPayload string) {

	botId := GetCommunityBotId(apiKey)

	webhookId := fetchWebhookId(cacheHelper, apiKey, webhookType, webhookUrl)
	if webhookId == 0 {
		logging.Error("Error fetching webhook id for url: ", webhookUrl)
		return
	}

	headers := map[string]interface{}{
		"x-member-id":     botId,
		"x-api-key":       apiKey,
		"x-platform-type": SwarmServiceHeader,
	}

	requestBody := map[string]interface{}{
		"is_active": false,
	}

	UpdateWebhooksEndpoint := fmt.Sprint(CommunityWebhooksEndpoint, "/", webhookId)

	// Send request to disable webhook
	respBytes, statusCode, err := GetRequestResponse(CaravanService, UpdateWebhooksEndpoint, PATCHRequest, headers, nil, requestBody)
	if err != nil || statusCode != http.StatusOK {
		logging.Error("Error disabling webhook: ", err, " Response: ", string(respBytes))
		return
	}

	logging.Info("Webhook disabled successfully: ", webhookUrl)

	// Send mail to team for webhook disable
	sendMailToTeamForWebhookFailure(apiKey, botId, webhookType, webhookUrl, statusCode, webhookResponse, webhookPayload)
}

// internal method to send mail to team for webhook disable
func sendMailToTeamForWebhookFailure(apiKey string, userId string, webhookType string, webhookUrl string,
	statusCode int, response string, payload string) {

	subject := WebhookFailureBody
	body := fmt.Sprintf(WebhookFailureBody, webhookType, time.Now(), webhookUrl, time.Now(), statusCode, response, payload)

	teamMails := environment.GoDotEnvVariable("TEAM_ADMIN_MAILS")
	mails := strings.Split(teamMails, ",")

	// Send mail using caravan service
	SendMail(userId, mails, subject, body)
}
