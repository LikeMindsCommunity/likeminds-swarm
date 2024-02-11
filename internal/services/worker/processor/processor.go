package processor

import (
	"context"

	"github.com/hibiken/asynq"
	"github.com/nateshr/likeminds-swarm/internal/api/handlers"
	"github.com/nateshr/likeminds-swarm/internal/services/worker"
)

// Run | Method to register task handlers and run the server
func (processor *RedisTaskProcessor) Run() error {

	// create a new ServeMux to register task handlers
	mux := asynq.NewServeMux()

	// register handlers for each task
	mux.HandleFunc(worker.BrokerConnectionTest, processor.ConnectionTest)
	mux.HandleFunc(worker.TaskSendDeleteTopicsFromPosts, processor.DeleteTopicsFromPosts)

	return processor.server.Run(mux)
}

type FeedTaskProcessor interface {
	Run() error
	ConnectionTest(ctx context.Context, task *asynq.Task) error
	DeleteTopicsFromPosts(ctx context.Context, task *asynq.Task) error
}

type RedisTaskProcessor struct {
	server       *asynq.Server
	feedHandlers *handlers.FeedHandlers
}

func NewTaskProcessor(feedHandlers *handlers.FeedHandlers) FeedTaskProcessor {

	// get Redis client options
	redisOpt := worker.GetRedisClientOpts()

	// get AsynQ server configurations
	config := worker.GetServerConfigurations()

	// creates a new server to process tasks
	server := asynq.NewServer(
		redisOpt,
		config,
	)

	return &RedisTaskProcessor{
		server:       server,
		feedHandlers: feedHandlers,
	}
}
