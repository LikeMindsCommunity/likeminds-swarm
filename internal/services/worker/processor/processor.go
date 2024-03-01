package processor

import (
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
	mux.HandleFunc(worker.TaskSendWebhookRequestWithPayload, processor.sendWebhookRequestWithPayload)
	mux.HandleFunc(worker.TaskTriggerPostCreationWebhook, processor.triggerPostCreationWebhook)
	mux.HandleFunc(worker.TaskTriggerPostLikedWebhook, processor.triggerPostLikedWebhook)
	mux.HandleFunc(worker.TaskTriggerPostPinnedWebhook, processor.triggerPostPinnedWebhook)
	mux.HandleFunc(worker.TaskTriggerPostTaggedWebhook, processor.triggerPostTaggedWebhook)
	mux.HandleFunc(worker.TaskTriggerCommentAddedWebhook, processor.triggerCommentAddedWebhook)
	mux.HandleFunc(worker.TaskTriggerCommentReactWebhook, processor.triggerCommentReactWebhook)
	mux.HandleFunc(worker.TaskTriggerCommentTaggedWebhook, processor.triggerCommentTaggedWebhook)
	mux.HandleFunc(worker.TaskTriggerCreatePost, processor.TriggerCreatePost)
	mux.HandleFunc(worker.TaskTriggerEditPost, processor.TriggerEditPost)
	mux.HandleFunc(worker.TaskTriggerDeletePost, processor.TriggerDeletePost)

	return processor.server.Run(mux)
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
