package handlers

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nateshr/likeminds-swarm/internal/api/enums"
	"github.com/nateshr/likeminds-swarm/internal/api/responses"
	"github.com/nateshr/likeminds-swarm/internal/services/externalHelpers"
	"github.com/nateshr/likeminds-swarm/internal/services/logging"
	"github.com/nateshr/likeminds-swarm/internal/utils"
)

func generatePostPayloadForWebhook(handlers *FeedHandlers, postId string) (*responses.WebhookPostPayload, error) {

	// Fetch post data
	postData, err := getPostByID(handlers.postHelper, postId)
	if err != nil {
		return nil, err
	}

	post := parsePostResponse(handlers.likeHelper, handlers.commentHelper, handlers.saveHelper, handlers.topicHelper,
		handlers.widgetHelper, *postData, postData.UserId, false, "", "", false, handlers.cacheHelper)

	// response["post"] = fetchPostData
	topics := getTopicDataFromPosts(handlers.topicHelper, post, post.CommunityId)
	widgets := getWidgetDataFromPosts(handlers, post, post.CommunityId, post.UserId)
	repostedPosts := getOriginalPostForReposts(handlers, post, post.CommunityId, post.UserId, false, "", "", false)

	// Fetch post creator data
	_, usersMeta := externalHelpers.FetchMemberMeta([]string{post.UserId}, post.UserId, post.CommunityId)
	if len(usersMeta.Members) == 0 {
		return nil, fmt.Errorf("error fetching post creator data for post id: %s", postId)
	}

	postPayload := responses.WebhookPostPayload{
		Post:          post,
		Topics:        topics,
		Widgets:       widgets,
		RepostedPosts: repostedPosts,
		PostCreator:   usersMeta.Members[0],
	}

	return &postPayload, nil
}

func generatePayloadForPostCreationWebhook(handlers *FeedHandlers, postId string) (*responses.WebhookPayload, error) {

	payload := &responses.WebhookPayload{
		Event:     enums.PostCreatedWebhookType,
		CreatedAt: time.Now().Unix(),
		ID:        uuid.New().String(),
		Source:    enums.WebhookSourceLMFeed,
		Data:      map[string]interface{}{},
	}

	// generate post payload
	postPaylod, err := generatePostPayloadForWebhook(handlers, postId)
	if err == nil {
		return nil, err
	}

	// Add post payload to payload data
	payload.Data["post_data"] = postPaylod

	return payload, nil
}

// Exposed method to send webhook request with payload to a url for feed processor
func SendWebhookRequestWithPayload(handlers *FeedHandlers, apiKey string, url string, payload *responses.WebhookPayload,
	webhookType string, secret string, retryCount int) error {

	if url == "" {
		logging.Error("URL is missing in the task payload")
		return nil
	}

	// Check if webhook url is active or not
	isActive := externalHelpers.IsWebhookUrlActive(handlers.cacheHelper, apiKey, webhookType, url)
	if !isActive {
		return nil
	}

	headers := map[string]interface{}{}

	// generate signature for the payload if secret is present
	if secret != "" {

		// Marshal payload to bytes
		payloadBytes, err := json.Marshal(*payload)
		if err != nil {
			logging.Error("Error marshalling payload for the webhook: ", err)
		}

		// Generate hexDigest signature for the payload
		signature, err := utils.GenerateMessageDigestOrSignature(payloadBytes, []byte(secret))
		if err != nil {
			logging.Error("Error generating signature for the payload: ", err)
		}

		// TODO: remove this print statement
		fmt.Println("Signature: ", signature, " Payload: ", string(payloadBytes))

		headers["x-signature"] = signature
	}

	// Send POST request to the webhook url with payload
	respbytes, statusCode, err := externalHelpers.SendPostRequestToExternalService(url, headers, *payload)
	if err != nil || statusCode != 200 {

		// Log error
		logging.Error("Error sending webhook request to URL: ", url, " Error: ", err, "status code: ", statusCode, " Response: ", string(respbytes))

		// Retry the request, else disable the webhook if retry limit reached
		if retryCount <= enums.WebhookRetryLimit {
			return fmt.Errorf("failed to send webhook request to URL: %s, retrying", url)
		} else {
			logging.Error("Disabling Webhook as retry limit reached for URL: ", url)

			// Disable the webhook and send mail to team
			go externalHelpers.DisableWebhookAndSendMail(handlers.cacheHelper, apiKey, webhookType, url)
		}
	} else {
		logging.Info("Successfully sent webhook request to URL: ", url, " for ApiKey: ", apiKey, " Response: ", string(respbytes))
	}

	return nil
}

// Exposed method to trigger post creation webhook for feed processor
func TriggerPostCreationWebhook(handlers *FeedHandlers, apiKey string, postId string) error {

	if apiKey == "" || postId == "" {
		logging.Error("API Key or Post ID is missing in the payload")
		return nil
	}

	// Fetch active webhook urls for type "post.created"
	activeWebhooksUrls := externalHelpers.FetchActiveWebhookUrls(handlers.cacheHelper, apiKey, enums.PostCreatedWebhookType)
	if len(activeWebhooksUrls) == 0 {
		return nil
	}

	// Create Payload for post creation webhook
	payload, err := generatePayloadForPostCreationWebhook(handlers, postId)
	if err != nil {
		return err
	}

	// Send webhook request to all active urls
	for _, webhookUrl := range activeWebhooksUrls {
		err := handlers.taskDistributor.SendWebhookRequestWithPayload(apiKey, webhookUrl, payload, enums.PostCreatedWebhookType, apiKey)
		if err != nil {
			logging.Error("error enqueuing task to send webhook request: ", err)
		}
	}

	return nil
}
