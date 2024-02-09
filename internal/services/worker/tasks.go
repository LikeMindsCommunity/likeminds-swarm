package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/nateshr/likeminds-swarm/internal/api/handlers"
	"github.com/nateshr/likeminds-swarm/internal/helpers"
	"github.com/nateshr/likeminds-swarm/internal/services/worker/distributor"
)

// DeleteTopicsFromPosts | Task to delete topics from posts
func (processor *RedisTaskProcessor) DeleteTopicsFromPosts(ctx context.Context, task *asynq.Task) error {

	var payload distributor.PayloadSendDeleteTopicsFromPostsTask
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	// convert topic_ids to object ids
	objectIDs := helpers.ConvertIdsToObjectIds(payload.TopicIds)

	return handlers.DeleteTopicsFromPostsAndUpdatePost(processor.handler, objectIDs)
}
