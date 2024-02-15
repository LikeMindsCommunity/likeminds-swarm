package worker

import "github.com/nateshr/likeminds-swarm/internal/api/responses"

// Task Names for each task type
const (
	BrokerConnectionTest              = "task:BrokerConnectionTest"
	TaskSendWebhookRequestWithPayload = "task:SendWebhookRequestWithPayload"
	TaskTriggerPostCreationWebhook    = "task:TriggerPostCreationWebhook"
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
