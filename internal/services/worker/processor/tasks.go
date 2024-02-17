package processor

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/nateshr/likeminds-swarm/internal/api/handlers"
	"github.com/nateshr/likeminds-swarm/internal/helpers"
	"github.com/nateshr/likeminds-swarm/internal/services/logging"
	"github.com/nateshr/likeminds-swarm/internal/services/worker"
)

// ConnectionTest | Task to test the connection with the broker
func (processor *RedisTaskProcessor) ConnectionTest(ctx context.Context, task *asynq.Task) error {
	logging.Info(`Successfully received and completed task:BrokerConnectionTest`)
	return nil
}

// DeleteTopicsFromPosts | Task to delete topics from posts
func (processor *RedisTaskProcessor) DeleteTopicsFromPosts(ctx context.Context, task *asynq.Task) error {

	var payload worker.PayloadSendDeleteTopicsFromPostsTask
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	// convert topic_ids to object ids
	objectIDs := helpers.ConvertIdsToObjectIds(payload.TopicIds)

	return handlers.DeleteTopicsFromPostsAndUpdatePost(processor.feedHandlers, objectIDs)
}
