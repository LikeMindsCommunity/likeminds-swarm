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
