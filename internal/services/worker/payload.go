package worker

// Task Names for each task type
const (
	BrokerConnectionTest              = "task:BrokerConnectionTest"
	TaskSendDeleteTopicsFromPosts     = "task:DeleteTopicsFromPosts"
	TaskSendWebhookRequestWithPayload = "task:SendWebhookRequestWithPayload"
	TaskTriggerPostCreationWebhook    = "task:TriggerPostCreationWebhook"
)

// Payload for the task to delete topics from posts
type PayloadSendDeleteTopicsFromPosts struct {
	TopicIds []string `json:"topic_ids"`
}

// Payload for the task to trigger post creation webhook
type PayloadTriggerPostCreationWebhook struct {
	PostId string `json:"post_id"`
	ApiKey string `json:"api_key"`
}

// Payload for the task to send webhook request with payload
type PayloadSendWebhookRequestWithPayload struct {
	ApiKey      string                 `json:"api_key"`
	Url         string                 `json:"url"`
	Payload     map[string]interface{} `json:"payload"`
	WebhookType string                 `json:"webhook_type"`
	Secret      string                 `json:"secret"`
}
