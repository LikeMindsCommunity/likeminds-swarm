package distributor

import (
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/nateshr/likeminds-swarm/internal/api/enums"
	"github.com/nateshr/likeminds-swarm/internal/services/worker"
)

// Task Distributor for "task:DeleteTopicsFromPosts"
func (distributor *RedisTaskDistributor) DeleteTopicsFromPosts(topicIds []string, opts ...asynq.Option) error {

	// create task payload
	payload := worker.PayloadSendDeleteTopicsFromPosts{
		TopicIds: topicIds,
	}

	// marshal the payload
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}

	// enqueue task with payload and options
	_, err = worker.EnqueueTaskToQueue(distributor.client, worker.TaskSendDeleteTopicsFromPosts, jsonPayload, opts...)
	if err != nil {
		return fmt.Errorf("failed to enqueue task %v", err)
	}

	return nil
}

// Task Distributor for "task:SendWebhookRequestWithPayload"
func (distributor *RedisTaskDistributor) SendWebhookRequestWithPayload(apiKey string, url string, payload map[string]interface{},
	webhookType string, secret string, opts ...asynq.Option) error {

	// create task payload
	taskPayload := worker.PayloadSendWebhookRequestWithPayload{
		ApiKey:      apiKey,
		Url:         url,
		Payload:     payload,
		WebhookType: webhookType,
		Secret:      secret,
	}

	// Set max retry limit for the task
	opts = append(opts, asynq.MaxRetry(enums.WebhookRetryLimit))

	// marshal the payload
	jsonPayload, err := json.Marshal(taskPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}

	// enqueue task with payload and options
	_, err = worker.EnqueueTaskToQueue(distributor.client, worker.TaskSendWebhookRequestWithPayload, jsonPayload, opts...)
	if err != nil {
		return fmt.Errorf("failed to enqueue task %v", err)
	}

	return nil
}

// Task Distributor for "task:TriggerPostCreationWebhook"
func (distributor *RedisTaskDistributor) TriggerPostCreationWebhook(postId string, apiKey string,
	opts ...asynq.Option) error {

	// create task payload
	payload := worker.PayloadTriggerPostCreationWebhook{
		PostId: postId,
		ApiKey: apiKey,
	}

	// marshal the payload
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}

	// enqueue task with payload and options
	_, err = worker.EnqueueTaskToQueue(distributor.client, worker.TaskTriggerPostCreationWebhook, jsonPayload, opts...)
	if err != nil {
		return fmt.Errorf("failed to enqueue task %v", err)
	}

	return nil
}
