package handlers

import (
	"context"
	"fmt"

	"github.com/hibiken/asynq"
)

type TaskProcessor interface {
	Start() error
	ProcessTaskDeleteTopicsFromPosts(
		ctx context.Context,
		task *asynq.Task,
	) error
}

type RedisTaskProcessor struct {
	server  *asynq.Server
	handler *FeedHandlers
}

func NewRedisTaskProcessor(redisOpt asynq.RedisClientOpt, handler *FeedHandlers) TaskProcessor {
	server := asynq.NewServer(
		redisOpt,
		asynq.Config{},
	)

	return &RedisTaskProcessor{
		server:  server,
		handler: handler,
	}
}

func (processor *RedisTaskProcessor) Start() error {
	fmt.Println("Server Mux Starting server")
	mux := asynq.NewServeMux()

	mux.HandleFunc(TaskSendDeleteTopicsFromPosts, processor.ProcessTaskDeleteTopicsFromPosts)
	return processor.server.Start(mux)
}
