package distributor

import (
	"github.com/hibiken/asynq"
	"github.com/nateshr/likeminds-swarm/internal/api/responses"
)

// FeedTaskDistributor | Interface for feed background task distributor
type FeedTaskDistributor interface {
	SendWebhookRequestWithPayload(apiKey string, url string, payload *responses.WebhookPayload, webhookType string, secret string, opts ...asynq.Option) error
	TriggerPostCreationWebhook(postId string, apiKey string, opts ...asynq.Option) error
	TriggerPostLikedWebhook(postId string, userId string, apiKey string, opts ...asynq.Option) error
	TriggerPostPinnedWebhook(postId string, userId string, apiKey string, opts ...asynq.Option) error
	TriggerPostTaggedWebhook(postId string, userIds []string, apiKey string, opts ...asynq.Option) error
	TriggerCommentAddedWebhook(commentId string, apiKey string, opts ...asynq.Option) error
	TriggerCommentReactWebhook(commentId string, userId string, apiKey string, opts ...asynq.Option) error
	TriggerCommentTaggedWebhook(commentId string, userIds []string, apiKey string, opts ...asynq.Option) error
	TriggerCreatePostBackgroundTasks(postId string, opts ...asynq.Option) error
	TriggerEditPostBackgroundTasks(postId string, opts ...asynq.Option) error
	TriggerDeletePostBackgroundTasks(postId string, opts ...asynq.Option) error
}
