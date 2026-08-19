package worker

import (
	"github.com/LikeMindsCommunity/likeminds-swarm/internal/api/constants"
	"github.com/LikeMindsCommunity/likeminds-swarm/internal/api/responses"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Task Names for each task type
const (
	BrokerConnectionTest                       = "task:BrokerConnectionTest"
	TaskSendWebhookRequestWithPayload          = "task:SendWebhookRequestWithPayload"
	TaskTriggerPostCreationWebhook             = "task:TriggerPostCreationWebhook"
	TaskTriggerPostLikedWebhook                = "task:TriggerPostLikedWebhook"
	TaskTriggerPostPinnedWebhook               = "task:TriggerPostPinnedWebhook"
	TaskTriggerPostTaggedWebhook               = "task:TriggerPostTaggedWebhook"
	TaskTriggerCommentAddedWebhook             = "task:TriggerCommentAddedWebhook"
	TaskTriggerCommentReactWebhook             = "task:TriggerCommentReactWebhook"
	TaskTriggerCommentTaggedWebhook            = "task:TriggerCommentTaggedWebhook"
	TaskAsyncCreatePostTasks                   = "task:AsyncCreatePostTasks"
	TaskAsyncEditPostTasks                     = "task:AsyncEditPostTasks"
	TaskAsyncDeletePostTasks                   = "task:AsnycDeletePostTasks"
	TaskAsyncSendNotification                  = "task:AsyncSendNotification"
	TaskAsyncCommunityDefaultFeed              = "task:AsyncCommunityDefaultFeed"
	TaskAsyncCreateActivityAndSendNotification = "task:AsyncCreateActivityAndSendNotification"
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

// Payload for send notification task
type PayloadSendNotification struct {
	ActivityID   primitive.ObjectID `json:"activity_id"`
	PlatformCode string             `json:"platform_code"`
	VersionCode  string             `json:"version_code"`
}

// Payload for create activiy and send notification
type PayloadCreateActivityAndSendNotification struct {
	CommunityID      int                      `json:"community_id"`
	ActionBy         []string                 `json:"action_by"`
	ActionOn         string                   `json:"action_on"`
	EntityType       constants.EntityType     `json:"entity_type"`
	EntityID         primitive.ObjectID       `json:"entity_id"`
	EntityOwnerID    string                   `json:"entitiy_owner_id"`
	Action           constants.ActivityAction `json:"action"`
	CtaData          map[string]interface{}   `json:"cta_data"`
	IsRead           bool                     `json:"is_read"`
	IsDeleted        bool                     `json:"is_deleted"`
	ActionByEntityId primitive.ObjectID       `json:"action_entity_id"`
	ActivityText     string                   `json:"activity_text"`
	PlatformCode     string                   `json:"platform_code"`
	VersionCode      string                   `json:"version_code"`
}
