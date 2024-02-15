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

	return handlers.TriggerPostCreationWebhook(processor.feedHandlers, payload.PostId, payload.ApiKey)
}
