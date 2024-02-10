package distributor

import (
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/nateshr/likeminds-swarm/internal/services/worker"
)

// Task Distributor for "task:DeleteTopicsFromPosts"
func (distributor *RedisTaskDistributor) DeleteTopicsFromPosts(topicIds []string, opts ...asynq.Option) error {

	// create task payload
	payload := worker.PayloadSendDeleteTopicsFromPostsTask{
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
