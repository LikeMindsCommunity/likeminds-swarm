package worker

import (
	"github.com/nateshr/likeminds-swarm/internal/api/responses"
)

// Task Names for each task type
const (
	BrokerConnectionTest              = "task:BrokerConnectionTest"
	TaskSendWebhookRequestWithPayload = "task:SendWebhookRequestWithPayload"
	TaskTriggerPostCreationWebhook    = "task:TriggerPostCreationWebhook"
	TaskTriggerPostLikedWebhook       = "task:TriggerPostLikedWebhook"
	TaskTriggerPostPinnedWebhook      = "task:TriggerPostPinnedWebhook"
	TaskTriggerPostTaggedWebhook      = "task:TriggerPostTaggedWebhook"
	TaskTriggerCommentAddedWebhook    = "task:TriggerCommentAddedWebhook"
	TaskTriggerCommentReactWebhook    = "task:TriggerCommentReactWebhook"
	TaskTriggerCommentTaggedWebhook   = "task:TriggerCommentTaggedWebhook"
	TaskTriggerCreatePost             = "task:TaskTriggerCreatePost"
	TaskTriggerEditPost               = "task:TaskTriggerEditPost"
	TaskTriggerDeletePost             = "task:TaskTriggerDeletePost"
)

// Payload for the task to trigger post creation webhook
type PayloadTriggerPostCreationWebhook struct {
	PostId string `json:"post_id"`
	ApiKey string `json:"api_key"`
}

// Payload for the task to send webhook request with payload
type PayloadSendWebhookRequestWithPayload struct {
	ApiKey      string                   `json:"api_key"`
	Url         string                   `json:"url"`
	WebhookType string                   `json:"webhook_type"`
	Secret      string                   `json:"secret"`
	Payload     responses.WebhookPayload `json:"payload"`
}

// Payload for the task to trigger post liked webhook
type PayloadTriggerPostLikedWebhook struct {
	PostId string `json:"post_id"`
	UserId string `json:"user_id"`
	ApiKey string `json:"api_key"`
}

// Payload for the task to trigger post pinned webhook
type PayloadTriggerPostPinnedWebhook struct {
	PostId string `json:"post_id"`
	UserId string `json:"user_id"`
	ApiKey string `json:"api_key"`
}

// Payload for the task to trigger post tagged webhook
type PayloadTriggerPostTaggedWebhook struct {
	PostId  string   `json:"post_id"`
	UserIds []string `json:"user_ids"`
	ApiKey  string   `json:"api_key"`
}

// Payload for the task to trigger comment added webhook
type PayloadTriggerCommentAddedWebhook struct {
	CommentId string `json:"comment_id"`
	ApiKey    string `json:"api_key"`
}

// Payload for the task to trigger comment react webhook
type PayloadTriggerCommentReactWebhook struct {
	CommentId string `json:"comment_id"`
	UserId    string `json:"user_id"`
	ApiKey    string `json:"api_key"`
}

// Payload for the task to trigger comment tagged webhook
type PayloadTriggerCommentTaggedWebhook struct {
	CommentId string   `json:"comment_id"`
	UserIds   []string `json:"user_ids"`
	ApiKey    string   `json:"api_key"`
}

type PayloadPost struct {
	PostID string `json:"post_id"`
}
