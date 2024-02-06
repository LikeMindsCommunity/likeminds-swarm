package handlers

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/hibiken/asynq"
	"github.com/nateshr/likeminds-swarm/internal/services/environment"
)

type FeedTaskProcessor interface {
	Run() error
	ProcessTaskDeleteTopicsFromPosts(ctx context.Context, task *asynq.Task) error
}

type RedisTaskProcessor struct {
	server  *asynq.Server
	handler *FeedHandlers
}

func NewRedisTaskProcessor(redisOpt asynq.RedisClientOpt, handler *FeedHandlers) FeedTaskProcessor {
	concurrency, _ := strconv.Atoi(environment.GoDotEnvVariable("ASYNQ_WORKER_CONCURRENCY"))
	// creates a new server to process tasks
	server := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Concurrency:  concurrency,
			ErrorHandler: asynq.ErrorHandlerFunc(reportError),
		},
	)

	return &RedisTaskProcessor{
		server:  server,
		handler: handler,
	}
}

func (processor *RedisTaskProcessor) Run() error {
	mux := asynq.NewServeMux()

	mux.HandleFunc(TaskSendDeleteTopicsFromPosts, processor.ProcessTaskDeleteTopicsFromPosts)
	return processor.server.Run(mux)
}

// callback function for error while executing tasks
func reportError(ctx context.Context, task *asynq.Task, err error) {
	log.Fatal(fmt.Sprintf("error while executing task %s with id: %s: %s", task.Type(), task.ResultWriter().TaskID(), err.Error()))
}
