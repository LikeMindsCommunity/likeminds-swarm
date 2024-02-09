package worker

import (
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	"github.com/nateshr/likeminds-swarm/internal/services/logging"
)

// EnqueueBackgroundTask | Enqueues a task to the background queue with the provided payload and options
func enqueueBackgroundTask(client *asynq.Client, taskName string, taskPayload []byte, opts ...asynq.Option) (*asynq.TaskInfo, error) {

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
