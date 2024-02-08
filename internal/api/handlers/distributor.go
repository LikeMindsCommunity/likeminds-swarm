package handlers

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	"github.com/nateshr/likeminds-swarm/internal/services/environment"
	"github.com/nateshr/likeminds-swarm/internal/services/logging"
)

// Task Names for each task type
const (
	TaskSendDeleteTopicsFromPosts = "task:DeleteTopicsFromPosts"
)

// PayloadSendDeleteTopicsFromPostsTask | Payload for the task to delete topics from posts
type PayloadSendDeleteTopicsFromPostsTask struct {
	TopicIds []string `json:"topic_ids"`
}
type FeedTaskDistributor interface {
	DistributeTaskDeleteTopicsFromPosts(topicIds []string, opts ...asynq.Option) error
}

type RedisTaskDistributor struct {
	client *asynq.Client
}

func NewTaskDistributor() FeedTaskDistributor {

	client := asynq.NewClient(asynq.RedisClientOpt{
		Addr: environment.GoDotEnvVariable("ASYNQ_BROKER_ADDRESS"),
	})

	return &RedisTaskDistributor{
		client: client,
	}
}

func (distributor *RedisTaskDistributor) DistributeTaskDeleteTopicsFromPosts(topicIds []string, opts ...asynq.Option) error {

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
	_, err = EnqueueBackgroundTask(distributor.client, TaskSendDeleteTopicsFromPosts, jsonPayload, opts...)
	if err != nil {
		return fmt.Errorf("failed to enqueue task %v", err)
	}

	return nil
}

// EnqueueBackgroundTask | Enqueues a task to the background queue with the provided payload and options
func EnqueueBackgroundTask(client *asynq.Client, taskName string, taskPayload []byte, opts ...asynq.Option) (*asynq.TaskInfo, error) {

	// Default options for tasks
	defaultOpts := []asynq.Option{
		asynq.MaxRetry(1),                 // max number of times the task can be retried
		asynq.Retention(10 * time.Minute), // how long to keep the task result in completed state
	}

	// creates a new task with the provided payload and default options
	task := asynq.NewTask(taskName, taskPayload, defaultOpts...)

	// enqueues the task to the queue with task specific options
	taskInfo, err := client.Enqueue(task, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to enqueue task %w", err)
	}

	logging.Info(fmt.Sprintf(` Task Enqueued | 
		taskId: %s;
		taskType: %s;
		taskPayload: %s; 
		taskQueue: %s; 
		taskState: %s; 
		taskResult: %s;`,
		taskInfo.ID, taskInfo.Type, taskInfo.Payload, taskInfo.Queue, taskInfo.State.String(), taskInfo.Result))

	return taskInfo, nil
}
