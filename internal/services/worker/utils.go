package worker

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/hibiken/asynq"
	"github.com/nateshr/likeminds-swarm/internal/services/environment"
	"github.com/nateshr/likeminds-swarm/internal/services/logging"
)

// GetRedisClientOpts | Returns the redis client options for the Asynq client
func GetRedisClientOpts() *asynq.RedisClientOpt {

	brokerAddress := environment.GoDotEnvVariable("ASYNQ_BROKER_ADDRESS")
	if brokerAddress == "" {
		brokerAddress = "localhost:6379"
	}

	redisClientOpt := asynq.RedisClientOpt{
		Addr: brokerAddress,
	}

	return &redisClientOpt
}

// EnqueueTaskToQueue | Enqueues a task to the queue with the provided payload and options
func EnqueueTaskToQueue(client *asynq.Client, taskName string, taskPayload []byte, opts ...asynq.Option) (*asynq.TaskInfo, error) {

	// Default options for tasks
	defaultOpts := []asynq.Option{
		asynq.MaxRetry(1),                 // max number of times the task can be retried
		asynq.Retention(10 * time.Minute), // how long to keep the task result in completed state
	}

	// creates a new task with the provided payload and DEFAULT OPTIONS
	task := asynq.NewTask(taskName, taskPayload, defaultOpts...)

	// enqueues the task to the queue with task specific options (Overrides default options)
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

// GetServerConfigurations | Returns configurations for the Asynq server
func GetServerConfigurations() asynq.Config {

	// default configurations for the Task Processor
	config := asynq.Config{
		// callback function for error while executing tasks
		ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
			// TODO: Try to find Task ID from the task payload and log it (This currently throws panic)
			logging.Fatal(fmt.Sprintf("error while executing task %s", err.Error()))
		}),
	}

	// set concurrency from environment variable if present | default: cpu cores count
	concurrency, err := strconv.Atoi(environment.GoDotEnvVariable("ASYNQ_WORKER_CONCURRENCY"))
	if err == nil {
		config.Concurrency = concurrency
	}

	return config

}
