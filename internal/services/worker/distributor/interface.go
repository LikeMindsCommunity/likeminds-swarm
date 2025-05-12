package distributor

import (
	"github.com/hibiken/asynq"
	"github.com/nateshr/likeminds-swarm/internal/api/constants"
	"github.com/nateshr/likeminds-swarm/internal/api/responses"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	AsyncCreatePostTasks(postId string, opts ...asynq.Option) error
	AsyncEditPostTasks(postId string, opts ...asynq.Option) error
	AsyncDeletePostTasks(postId string, opts ...asynq.Option) error
	AsyncSendNotification(activityID primitive.ObjectID, platformCode string, versionCode string, opts ...asynq.Option) error
	AsyncCreateActivityAndSendNotification(
		communityID int, actionBy []string, actionOn string, entityType constants.EntityType, entityID primitive.ObjectID, entityOwnerID string, action constants.ActivityAction, ctaData map[string]interface{}, isRead bool, isDeleted bool, actionByEntityId primitive.ObjectID, activityText string, platformCode string, versionCode string,
		opts ...asynq.Option) error
}
