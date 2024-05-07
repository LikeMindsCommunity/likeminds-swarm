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

	// Fetch post response
	post, err := FetchSinglePostResponse(handlers, postId)
	if err != nil {
		return nil, err
	}

	// Fetch topics for the post
	topics, err := fetchAndParseTopicsForResponse(handlers.topicHelper, post.Topics, post.CommunityId)
	if err != nil {
		return nil, err
	}

	// Fetch post creator data
	_, usersMeta := externalHelpers.FetchMemberMeta([]string{post.UserId}, post.UserId, post.CommunityId)
	if len(usersMeta.Members) == 0 {
		return nil, fmt.Errorf("error fetching post creator data for post id: %s", postId)
	}

	// Fetch widgets for the post
	widgets, err := parseWidgetsResponse(handlers, getWidgetIdsFromAttachments(post.Attachments), post.CommunityId, enums.IsCM(usersMeta.Members[0].State), post.UserId)
	if err != nil {
		return nil, err
	}

	postPayload := responses.WebhookPostPayload{
		Post:        *post,
		Topics:      topics,
		Widgets:     widgets,
		PostCreator: usersMeta.Members[0],
	}

	return &postPayload, nil
}

func generateUsersPayloadForWebhook(userIds []string, communityId int) ([]externalHelpers.MemberMeta, error) {

	if len(userIds) == 0 {
		return nil, fmt.Errorf("user ids are missing in the payload")
	}

	// Fetch users meta
	_, usersMeta := externalHelpers.FetchMemberMeta(userIds, userIds[0], communityId)
	if len(usersMeta.Members) == 0 {
		return nil, fmt.Errorf("error fetching user data for user ids: %v", userIds)
	}

	return usersMeta.Members, nil
}

func generateCommentPayloadForWebhook(handlers *FeedHandlers, commentId string) (*responses.WebhookCommentPayload, error) {

	// Fetch comment response
	comment, err := FetchSingleCommentWithParentResponse(handlers, commentId)
	if err != nil {
		return nil, err
	}

	// Fetch comment creator data
	_, usersMeta := externalHelpers.FetchMemberMeta([]string{comment.UserId}, comment.UserId, comment.CommunityId)
	if len(usersMeta.Members) == 0 {
		return nil, fmt.Errorf("error fetching comment creator data for comment id: %s", commentId)
	}

	commentPayload := responses.WebhookCommentPayload{
		Comment:        *comment,
		CommentCreator: usersMeta.Members[0],
	}

	return &commentPayload, nil
}

// method to generate payload for different webhook types
func generatePayloadForWebhooks(handlers *FeedHandlers, postId string, userIds []string, webhookType string) (*responses.WebhookPayload, error) {

	payload := &responses.WebhookPayload{
		Event:     webhookType,
		CreatedAt: time.Now().UnixMilli(),
		ID:        uuid.New().String(),
		Source:    enums.WebhookSourceLMFeed,
		Data:      map[string]interface{}{},
	}

	switch webhookType {
	case enums.PostCreatedWebhookType:

		// generate post payload
		postPaylod, err := generatePostPayloadForWebhook(handlers, postId)
		if err != nil {
			return nil, err
		}

		// Add post payload to payload data
		payload.Data["post_data"] = postPaylod

	case enums.PostLikedWebhookType:

		// generate post payload
		postPaylod, err := generatePostPayloadForWebhook(handlers, postId)
		if err != nil {
			return nil, err
		}

		// Generate user meta
		userMeta, err := generateUsersPayloadForWebhook(userIds, postPaylod.Post.CommunityId)
		if err != nil {
			return nil, err
		}

		// Add to payload data
		payload.Data["post_data"] = postPaylod
		payload.Data["post_liked_by"] = userMeta[0]

	case enums.PostPinnedWebhookType:

		// generate post payload
		postPaylod, err := generatePostPayloadForWebhook(handlers, postId)
		if err != nil {
			return nil, err
		}

		// Generate user meta
		userMeta, err := generateUsersPayloadForWebhook(userIds, postPaylod.Post.CommunityId)
		if err != nil {
			return nil, err
		}

		// Add to payload data
		payload.Data["post_data"] = postPaylod
		payload.Data["pinned_by"] = userMeta[0]

	case enums.PostTaggedWebhookType:

		// generate post payload
		postPaylod, err := generatePostPayloadForWebhook(handlers, postId)
		if err != nil {
			return nil, err
		}

		// Add to payload data
		usersMeta, err := generateUsersPayloadForWebhook(userIds, postPaylod.Post.CommunityId)
		if err != nil {
			return nil, err
		}

		payload.Data["post_data"] = postPaylod
		payload.Data["tagged_users"] = usersMeta

	case enums.CommentAddedWebhook:

		// generate comment payload
		commentPayload, err := generateCommentPayloadForWebhook(handlers, postId)
		if err != nil {
			return nil, err
		}

		// Generate post payload
		postPaylod, err := generatePostPayloadForWebhook(handlers, commentPayload.Comment.PostId.Hex())
		if err != nil {
			return nil, err
		}

		// Add to payload data
		payload.Data["comment_data"] = commentPayload
		payload.Data["post_data"] = postPaylod

	case enums.CommentReactWebhook:

		// generate comment payload
		commentPayload, err := generateCommentPayloadForWebhook(handlers, postId)
		if err != nil {
			return nil, err
		}

		// Generate post payload
		postPaylod, err := generatePostPayloadForWebhook(handlers, commentPayload.Comment.PostId.Hex())
		if err != nil {
			return nil, err
		}

		// Generate user meta
		userMeta, err := generateUsersPayloadForWebhook(userIds, postPaylod.Post.CommunityId)
		if err != nil {
			return nil, err
		}

		// Add to payload data
		payload.Data["comment_data"] = commentPayload
		payload.Data["post_data"] = postPaylod
		payload.Data["liked_by_user"] = userMeta[0]

	case enums.CommentTaggedWebhook:

		// generate comment payload
		commentPayload, err := generateCommentPayloadForWebhook(handlers, postId)
		if err != nil {
			return nil, err
		}

		// Generate post payload
		postPaylod, err := generatePostPayloadForWebhook(handlers, commentPayload.Comment.PostId.Hex())
		if err != nil {
			return nil, err
		}

		// Generate users meta
		usersMeta, err := generateUsersPayloadForWebhook(userIds, postPaylod.Post.CommunityId)
		if err != nil {
			return nil, err
		}

		// Add to payload data
		payload.Data["comment_data"] = commentPayload
		payload.Data["post_data"] = postPaylod
		payload.Data["tagged_users"] = usersMeta

	default:
		return nil, fmt.Errorf("invalid webhook type: %s", webhookType)
	}

	return payload, nil
}

// Exposed method to send webhook request with payload to a url for feed processor
func SendWebhookRequestWithPayload(handlers *FeedHandlers, apiKey string, url string, payload *responses.WebhookPayload,
	webhookType string, secret string, retryCount int) error {

	if url == "" {
		logging.Error("URL is missing in the task payload")
		return nil
	}

	isActive := externalHelpers.IsWebhookUrlActive(handlers.cacheHelper, apiKey, webhookType, url)
	if !isActive {
		return nil
	}

	headers := map[string]interface{}{}

	// generate signature for the payload if secret is present
	if secret != "" {

		payloadBytes, err := json.Marshal(*payload)
		if err != nil {
			logging.Error("Error marshalling payload for the webhook: ", err)
		}

		signature, err := utils.GenerateMessageDigestOrSignature(payloadBytes, []byte(secret))
		if err != nil {
			logging.Error("Error generating signature for the payload: ", err)
		}

		headers["x-signature"] = signature
	}

	// Send POST request to the webhook url with payload
	respbytes, statusCode, err := externalHelpers.SendPostRequestToExternalService(url, headers, *payload)
	if err != nil || statusCode != 200 {

		logging.Error("Error sending webhook request to URL: ", url, " Error: ", err, "status code: ", statusCode, " Response: ", string(respbytes))

		// Retry the request, else disable the webhook if retry limit reached
		if retryCount <= enums.WebhookRetryLimit {
			return fmt.Errorf("failed to send webhook request to URL: %s, retrying", url)
		} else {
			logging.Error("Disabling Webhook as retry limit reached for URL: ", url)

			payloadBytes, _ := json.MarshalIndent(payload, "", " ")
			go externalHelpers.DisableWebhookAndSendMail(handlers.cacheHelper, apiKey, webhookType, url, string(respbytes), string(payloadBytes))
		}

	} else {
		logging.Info(fmt.Sprintf("Successfully sent (%s) webhook request to URL: %s for ApiKey: %s",
			webhookType, url, apiKey))
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
	payload, err := generatePayloadForWebhooks(handlers, postId, []string{}, enums.PostCreatedWebhookType)
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

// Exposed method to trigger post liked webhook for feed processor
func TriggerPostLikedWebhook(handlers *FeedHandlers, apiKey string, postId string, userId string) error {

	if postId == "" || userId == "" || apiKey == "" {
		logging.Error("Post ID, User ID or API Key is missing in the payload")
		return nil
	}

	// Fetch active webhook urls for type "post.liked"
	activeWebhooksUrls := externalHelpers.FetchActiveWebhookUrls(handlers.cacheHelper, apiKey, enums.PostLikedWebhookType)
	if len(activeWebhooksUrls) == 0 {
		return nil
	}

	// Create Payload for post liked webhook
	payload, err := generatePayloadForWebhooks(handlers, postId, []string{userId}, enums.PostLikedWebhookType)
	if err != nil {
		return err
	}

	// Send webhook request to all active urls
	for _, webhookUrl := range activeWebhooksUrls {
		err := handlers.taskDistributor.SendWebhookRequestWithPayload(apiKey, webhookUrl, payload, enums.PostLikedWebhookType, apiKey)
		if err != nil {
			logging.Error("error enqueuing task to send webhook request: ", err)
		}
	}

	return nil
}

// Exposed method to trigger post pinned webhook for feed processor
func TriggerPostPinnedWebhook(handlers *FeedHandlers, apiKey string, postId string, userId string) error {

	if postId == "" || userId == "" || apiKey == "" {
		logging.Error("Post ID, User ID or API Key is missing in the payload")
		return nil
	}

	// Fetch active webhook urls for type "post.pinned"
	activeWebhooksUrls := externalHelpers.FetchActiveWebhookUrls(handlers.cacheHelper, apiKey, enums.PostPinnedWebhookType)
	if len(activeWebhooksUrls) == 0 {
		return nil
	}

	// Create Payload for post pinned webhook
	payload, err := generatePayloadForWebhooks(handlers, postId, []string{userId}, enums.PostPinnedWebhookType)
	if err != nil {
		return err
	}

	// Send webhook request to all active urls
	for _, webhookUrl := range activeWebhooksUrls {
		err := handlers.taskDistributor.SendWebhookRequestWithPayload(apiKey, webhookUrl, payload, enums.PostPinnedWebhookType, apiKey)
		if err != nil {
			logging.Error("error enqueuing task to send webhook request: ", err)
		}
	}

	return nil
}

// Exposed method to trigger post tagged webhook for feed processor
func TriggerPostTaggedWebhook(handlers *FeedHandlers, apiKey string, postId string, userIds []string) error {

	if postId == "" || len(userIds) == 0 || apiKey == "" {
		logging.Error("Post ID, User IDs or API Key is missing in the payload")
		return nil
	}

	// Fetch active webhook urls for type "post.tagged"
	activeWebhooksUrls := externalHelpers.FetchActiveWebhookUrls(handlers.cacheHelper, apiKey, enums.PostTaggedWebhookType)
	if len(activeWebhooksUrls) == 0 {
		return nil
	}

	// Create Payload for post tagged webhook
	payload, err := generatePayloadForWebhooks(handlers, postId, userIds, enums.PostTaggedWebhookType)
	if err != nil {
		return err
	}

	// Send webhook request to all active urls
	for _, webhookUrl := range activeWebhooksUrls {
		err := handlers.taskDistributor.SendWebhookRequestWithPayload(apiKey, webhookUrl, payload, enums.PostTaggedWebhookType, apiKey)
		if err != nil {
			logging.Error("error enqueuing task to send webhook request: ", err)
		}
	}

	return nil
}

// Exposed method to trigger comment added webhook for feed processor
func TriggerCommentAddedWebhook(handlers *FeedHandlers, apiKey string, commentId string) error {

	if commentId == "" || apiKey == "" {
		logging.Error("Comment ID or API Key is missing in the payload")
		return nil
	}

	// Fetch active webhook urls for type "comment.added"
	activeWebhooksUrls := externalHelpers.FetchActiveWebhookUrls(handlers.cacheHelper, apiKey, enums.CommentAddedWebhook)
	if len(activeWebhooksUrls) == 0 {
		return nil
	}

	// Create Payload for comment added webhook
	payload, err := generatePayloadForWebhooks(handlers, commentId, []string{}, enums.CommentAddedWebhook)
	if err != nil {
		return err
	}

	// Send webhook request to all active urls
	for _, webhookUrl := range activeWebhooksUrls {
		err := handlers.taskDistributor.SendWebhookRequestWithPayload(apiKey, webhookUrl, payload, enums.CommentAddedWebhook, apiKey)
		if err != nil {
			logging.Error("error enqueuing task to send webhook request: ", err)
		}
	}

	return nil
}

// Exposed method to trigger comment react webhook for feed processor
func TriggerCommentReactWebhook(handlers *FeedHandlers, apiKey string, commentId string, userId string) error {

	if commentId == "" || userId == "" || apiKey == "" {
		logging.Error("Comment ID, User ID or API Key is missing in the payload")
		return nil
	}

	// Fetch active webhook urls for type "comment.react"
	activeWebhooksUrls := externalHelpers.FetchActiveWebhookUrls(handlers.cacheHelper, apiKey, enums.CommentReactWebhook)
	if len(activeWebhooksUrls) == 0 {
		return nil
	}

	// Create Payload for comment react webhook
	payload, err := generatePayloadForWebhooks(handlers, commentId, []string{userId}, enums.CommentReactWebhook)
	if err != nil {
		return err
	}

	// Send webhook request to all active urls
	for _, webhookUrl := range activeWebhooksUrls {
		err := handlers.taskDistributor.SendWebhookRequestWithPayload(apiKey, webhookUrl, payload, enums.CommentReactWebhook, apiKey)
		if err != nil {
			logging.Error("error enqueuing task to send webhook request: ", err)
		}
	}

	return nil
}

// Exposed method to trigger comment tagged webhook for feed processor
func TriggerCommentTaggedWebhook(handlers *FeedHandlers, apiKey string, commentId string, userIds []string) error {

	if commentId == "" || len(userIds) == 0 || apiKey == "" {
		logging.Error("Comment ID, User IDs or API Key is missing in the payload")
		return nil
	}

	// Fetch active webhook urls for type "comment.tagged"
	activeWebhooksUrls := externalHelpers.FetchActiveWebhookUrls(handlers.cacheHelper, apiKey, enums.CommentTaggedWebhook)
	if len(activeWebhooksUrls) == 0 {
		return nil
	}

	// Create Payload for comment tagged webhook
	payload, err := generatePayloadForWebhooks(handlers, commentId, userIds, enums.CommentTaggedWebhook)
	if err != nil {
		return err
	}

	// Send webhook request to all active urls
	for _, webhookUrl := range activeWebhooksUrls {
		err := handlers.taskDistributor.SendWebhookRequestWithPayload(apiKey, webhookUrl, payload, enums.CommentTaggedWebhook, apiKey)
		if err != nil {
			logging.Error("error enqueuing task to send webhook request: ", err)
		}
	}

	return nil
}
