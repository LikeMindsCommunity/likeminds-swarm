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

	// add logging middleware to the mux
	mux.Use(worker.LoggingMiddleware)

	// register handlers for each task
	mux.HandleFunc(worker.BrokerConnectionTest, processor.connectionTest)
	mux.HandleFunc(worker.TaskSendDeleteTopicsFromPosts, processor.deleteTopicsFromPosts)
	mux.HandleFunc(worker.TaskSendWebhookRequestWithPayload, processor.sendWebhookRequestWithPayload)
	mux.HandleFunc(worker.TaskTriggerPostCreationWebhook, processor.triggerPostCreationWebhook)

	return processor.server.Run(mux)
}

type FeedTaskProcessor interface {
	Run() error
	connectionTest(ctx context.Context, task *asynq.Task) error
	deleteTopicsFromPosts(ctx context.Context, task *asynq.Task) error
	sendWebhookRequestWithPayload(ctx context.Context, task *asynq.Task) error
	triggerPostCreationWebhook(ctx context.Context, task *asynq.Task) error
}

type RedisTaskProcessor struct {
	server       *asynq.Server
	feedHandlers *handlers.FeedHandlers
}

func NewTaskProcessor(feedHandlers *handlers.FeedHandlers, QueueNames []string) FeedTaskProcessor {

	// get Redis client options
	redisOpt := worker.GetRedisClientOpts()

	// get AsynQ server configurations
	config := worker.GetServerConfigurations(QueueNames)

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
