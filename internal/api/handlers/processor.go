package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hibiken/asynq"
	"github.com/nateshr/likeminds-swarm/internal/helpers"
	"github.com/nateshr/likeminds-swarm/internal/services/environment"
	"github.com/nateshr/likeminds-swarm/internal/services/logging"
)

type FeedTaskProcessor interface {
	Run() error
	DeleteTopicsFromPosts(ctx context.Context, task *asynq.Task) error
}

type RedisTaskProcessor struct {
	server  *asynq.Server
	handler *FeedHandlers
}

func NewTaskProcessor(redisOpt asynq.RedisClientOpt, handler *FeedHandlers) FeedTaskProcessor {

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

// Run | Method to register task handlers and run the server
func (processor *RedisTaskProcessor) Run() error {

	// create a new ServeMux to register task handlers
	mux := asynq.NewServeMux()

	// register handlers for each task
	mux.HandleFunc(TaskSendDeleteTopicsFromPosts, processor.DeleteTopicsFromPosts)

	return processor.server.Run(mux)
}

// DeleteTopicsFromPosts | Task to delete topics from posts
func (processor *RedisTaskProcessor) DeleteTopicsFromPosts(ctx context.Context, task *asynq.Task) error {

	var payload PayloadSendDeleteTopicsFromPostsTask
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	// convert topic_ids to object ids
	objectIDs := helpers.ConvertIdsToObjectIds(payload.TopicIds)

	return DeleteTopicsFromPostsAndUpdatePost(processor.handler, objectIDs)
}
