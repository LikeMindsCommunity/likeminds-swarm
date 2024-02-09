package processor

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hibiken/asynq"
	"github.com/nateshr/likeminds-swarm/internal/api/handlers"
	"github.com/nateshr/likeminds-swarm/internal/services/environment"
	"github.com/nateshr/likeminds-swarm/internal/services/logging"
	"github.com/nateshr/likeminds-swarm/internal/services/worker"
)

// Run | Method to register task handlers and run the server
func (processor *RedisTaskProcessor) Run() error {

	// create a new ServeMux to register task handlers
	mux := asynq.NewServeMux()

	// register handlers for each task
	mux.HandleFunc(worker.TaskSendDeleteTopicsFromPosts, processor.DeleteTopicsFromPosts)

	return processor.server.Run(mux)
}

type FeedTaskProcessor interface {
	Run() error
	DeleteTopicsFromPosts(ctx context.Context, task *asynq.Task) error
}

type RedisTaskProcessor struct {
	server  *asynq.Server
	handler *handlers.FeedHandlers
}

func NewTaskProcessor(handler *handlers.FeedHandlers) FeedTaskProcessor {

	// initiate redis client
	redisOpt := asynq.RedisClientOpt{
		Addr: environment.GoDotEnvVariable("ASYNQ_BROKER_ADDRESS"),
	}

	// default configurations for the Task Processor
	config := asynq.Config{
		// callback function for error while executing tasks
		ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
			// TODO: Try to find Task ID from the task payload and log it (This currently throws panic)
			logging.Fatal(fmt.Sprintf("error while executing task %s with id: %s: %s", task.Type(), task.ResultWriter().TaskID(), err.Error()))
		}),
	}

	// set concurrency from environment variable if present | default: cpu cores count
	concurrency, err := strconv.Atoi(environment.GoDotEnvVariable("ASYNQ_WORKER_CONCURRENCY"))
	if err == nil {
		config.Concurrency = concurrency
	}

	// creates a new server to process tasks
	server := asynq.NewServer(
		redisOpt,
		config,
	)

	return &RedisTaskProcessor{
		server:  server,
		handler: handler,
	}
}
