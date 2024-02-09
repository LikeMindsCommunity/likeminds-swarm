package distributor

import (
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
)

// Task Distributor for "task:DeleteTopicsFromPosts"
func (distributor *RedisTaskDistributor) DeleteTopicsFromPosts(topicIds []string, opts ...asynq.Option) error {

	// create task payload
	payload := PayloadSendDeleteTopicsFromPostsTask{
		TopicIds: topicIds,
	}

	// marshal the payload
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}

	// enqueue task with payload and options
	_, err = enqueueBackgroundTask(distributor.client, TaskSendDeleteTopicsFromPosts, jsonPayload, opts...)
	if err != nil {
		return fmt.Errorf("failed to enqueue task %v", err)
	}

	return nil
}

// FeedTaskDistributor | Interface for the feed task distributor
type FeedTaskDistributor interface {
	DeleteTopicsFromPosts(topicIds []string, opts ...asynq.Option) error
}

type RedisTaskDistributor struct {
	client *asynq.Client
}

// NewTaskDistributor | Creates a new task distributor for feed background tasks
func NewTaskDistributor() FeedTaskDistributor {

	// create a new redis client
	redisClientOpts := InitRedisBrokerClient()

	// create a new asynq client
	client := asynq.NewClient(redisClientOpts)

	return &RedisTaskDistributor{
		client: client,
	}
}
