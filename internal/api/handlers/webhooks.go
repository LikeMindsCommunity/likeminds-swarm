package handlers

import (
	"fmt"

	"github.com/nateshr/likeminds-swarm/internal/api/enums"
	"github.com/nateshr/likeminds-swarm/internal/services/externalHelpers"
	"github.com/nateshr/likeminds-swarm/internal/services/logging"
)

func createWebhookPayloadForPostCreation(postId string, userId string) string {

	return ""
}

// Exposed method to send webhook request with payload to a url
func SendWebhookRequestWithPayload(handlers *FeedHandlers, apiKey string, url string, payload map[string]interface{},
	webhookType string, secret string, retryCount int) error {

	if url == "" || payload == nil {
		logging.Error("URL or Payload is missing in the payload")
		return nil
	}

	// Check if webhook url is active or not
	isActive := externalHelpers.IsWebhookUrlActive(handlers.cacheHelper, apiKey, webhookType, url)
	if !isActive {
		return nil
	}

	headers := map[string]interface{}{}

	// Add secret to the payload if present
	if secret != "" {
		signature := "" // TODO: Create utility method to create hex encoded HMAC signature
		headers["x-signature"] = signature

	}

	// Send POST request to the webhook url with payload
	respbytes, statusCode, err := externalHelpers.SendPostRequestToExternalService(url, headers, payload)
	if err != nil || statusCode != 200 {

		// Log error
		logging.Error("Error sending webhook request to URL: ", url, " Error: ", err, "status code: ", statusCode, " Response: ", string(respbytes))

		// Retry the request if retry count is less else disable the webhook and send mail
		if retryCount <= enums.WebhookRetryLimit {

			return fmt.Errorf("failed to send webhook request to URL: %s, retrying", url)

		} else {

			logging.Error("Disabling Webhook as retry limit reached for URL: ", url)

			// Disable the webhook and send mail to team
			go externalHelpers.DisableWebhookAndSendMail(handlers.cacheHelper, apiKey, webhookType, url)
		}
	}

	return nil
}

// Exposed method to trigger post creation webhook for feed processor
func TriggerPostCreationWebhook(handlers *FeedHandlers, apiKey string, postId string) error {

	if apiKey == "" || postId == "" {
		logging.Error("API Key, Post ID or User ID is missing in the payload")
		return nil
	}

	// Fetch active webhook urls for type "post.created"
	activeWebhooks := externalHelpers.FetchActiveWebhookUrls(handlers.cacheHelper, apiKey, enums.PostCreatedWebhookType)
	if len(activeWebhooks) == 0 {
		return nil
	}

	// Create Payload for the webhook
	// TODO: Call utility method - "createWebhookPayloadForPostCreation()"
	payload := map[string]interface{}{}

	// Trigger webhooks with the payload for each webhook url
	for _, webhookUrl := range activeWebhooks {
		err := handlers.taskDistributor.SendWebhookRequestWithPayload(apiKey, webhookUrl, payload, enums.PostCreatedWebhookType, apiKey)
		if err != nil {
			logging.Error("Error sending webhook request to URL: ", webhookUrl, " Error: ", err)
		}
	}

	return nil
}
