package processor

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/nateshr/likeminds-swarm/internal/api/handlers"
	"github.com/nateshr/likeminds-swarm/internal/services/logging"
	"github.com/nateshr/likeminds-swarm/internal/services/worker"
)

// Task to test the connection with the broker
func (processor *RedisTaskProcessor) connectionTest(ctx context.Context, task *asynq.Task) error {
	logging.Info(`Successfully received and completed task:BrokerConnectionTest`)
	return nil
}

// Task to send webhook request with payload
func (processor *RedisTaskProcessor) sendWebhookRequestWithPayload(ctx context.Context, task *asynq.Task) error {

	payload := worker.PayloadSendWebhookRequestWithPayload{}
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	// Get task retry count
	retryCount, _ := asynq.GetRetryCount(ctx)

	return handlers.SendWebhookRequestWithPayload(processor.feedHandlers, payload.ApiKey, payload.Url, &payload.Payload,
		payload.WebhookType, payload.Secret, retryCount)
}

// Task to trigger post creation webhook
func (processor *RedisTaskProcessor) triggerPostCreationWebhook(ctx context.Context, task *asynq.Task) error {

	payload := worker.PayloadTriggerPostCreationWebhook{}
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	return handlers.TriggerPostCreationWebhook(processor.feedHandlers, payload.ApiKey, payload.PostId)
}

// Task to trigger post liked webhook
func (processor *RedisTaskProcessor) triggerPostLikedWebhook(ctx context.Context, task *asynq.Task) error {

	payload := worker.PayloadTriggerPostLikedWebhook{}
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	return handlers.TriggerPostLikedWebhook(processor.feedHandlers, payload.ApiKey, payload.PostId, payload.UserId)
}

// Task to trigger post pinned webhook
func (processor *RedisTaskProcessor) triggerPostPinnedWebhook(ctx context.Context, task *asynq.Task) error {

	payload := worker.PayloadTriggerPostPinnedWebhook{}
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	return handlers.TriggerPostPinnedWebhook(processor.feedHandlers, payload.ApiKey, payload.PostId, payload.UserId)
}

// Task to trigger post tagged webhook
func (processor *RedisTaskProcessor) triggerPostTaggedWebhook(ctx context.Context, task *asynq.Task) error {

	payload := worker.PayloadTriggerPostTaggedWebhook{}
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	return handlers.TriggerPostTaggedWebhook(processor.feedHandlers, payload.ApiKey, payload.PostId, payload.UserIds)
}

// Task to trigger comment added webhook
func (processor *RedisTaskProcessor) triggerCommentAddedWebhook(ctx context.Context, task *asynq.Task) error {

	payload := worker.PayloadTriggerCommentAddedWebhook{}
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	return handlers.TriggerCommentAddedWebhook(processor.feedHandlers, payload.ApiKey, payload.CommentId)
}

// Task to trigger comment react webhook
func (processor *RedisTaskProcessor) triggerCommentReactWebhook(ctx context.Context, task *asynq.Task) error {

	payload := worker.PayloadTriggerCommentReactWebhook{}
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	return handlers.TriggerCommentReactWebhook(processor.feedHandlers, payload.ApiKey, payload.CommentId, payload.UserId)
}

// Task to trigger comment tagged webhook
func (processor *RedisTaskProcessor) triggerCommentTaggedWebhook(ctx context.Context, task *asynq.Task) error {

	payload := worker.PayloadTriggerCommentTaggedWebhook{}
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	return handlers.TriggerCommentTaggedWebhook(processor.feedHandlers, payload.ApiKey, payload.CommentId, payload.UserIds)
}

// Task to trigger create post
func (processor *RedisTaskProcessor) createPostBackgroundTasks(ctx context.Context, task *asynq.Task) error {

	payload := worker.PayloadPost{}
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	return handlers.CreatePostAsyncTasks(processor.feedHandlers, payload.PostID)
}

// Task to trigger edit post
func (processor *RedisTaskProcessor) editPostBackgroundTasks(ctx context.Context, task *asynq.Task) error {

	payload := worker.PayloadPost{}
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	return handlers.EditPostAsyncTasks(processor.feedHandlers, payload.PostID)
}

// Task to trigger delete post
func (processor *RedisTaskProcessor) deletePostBackgroundTasks(ctx context.Context, task *asynq.Task) error {

	payload := worker.PayloadPost{}
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	return handlers.DeletePostTopics(processor.feedHandlers, payload.PostID)
}

// Task to send notification
func (processor *RedisTaskProcessor) sendNotification(ctx context.Context, task *asynq.Task) error {

	payload := worker.PayloadSendNotification{}
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	return handlers.SendNotification(*processor.feedHandlers, payload.ActivityID, payload.PlatformCode, payload.VersionCode)
}

// Task to compute community default feed
func (processor *RedisTaskProcessor) computeCommunityDefaultFeed(ctx context.Context, task *asynq.Task) error {
	return handlers.AsyncComputeCommunityDefaultFeed(processor.feedHandlers)
}

// Task to createActivity and send notification
func (processor *RedisTaskProcessor) createActivityAndSendNotificationBackgroundTask(ctx context.Context, task *asynq.Task) error {

	payload := worker.PayloadCreateActivityAndSendNotification{}
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	// return handlers.createActivityAndSendNotification(*processor.feedHandlers, payload.ActivityID, payload.PlatformCode, payload.VersionCode)
	return processor.feedHandlers.CreateActivityAndSendNotification(
		payload.CommunityID,
		payload.ActionBy,
		payload.ActionOn,
		payload.EntityType,
		payload.EntityID,
		payload.EntityOwnerID,
		payload.Action,
		payload.CtaData,
		payload.IsRead,
		payload.IsDeleted,
		payload.ActionByEntityId,
		payload.ActivityText,
		payload.PlatformCode,
		payload.VersionCode)
}
