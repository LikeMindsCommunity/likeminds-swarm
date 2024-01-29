package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const TaskSendDeleteTopicsFromPosts = "task:send_delete_topics_from_posts"

type PayloadSendDeleteTopicsFromPostsTask struct {
	TopicIds []string `json:"topic_ids"`
}

// arguments -> (topicsIds)

func (distributor *RedisTaskDistributor) DistributeTaskDeleteTopicsFromPosts(ctx context.Context, payload *PayloadSendDeleteTopicsFromPostsTask, opts ...asynq.Option,
) error {
	// wrap into payload
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}
	task := asynq.NewTask(TaskSendDeleteTopicsFromPosts, jsonPayload, opts...)
	info, err := distributor.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to enqueue task", err)
	}
	fmt.Print("queue", info.Queue)

	return nil
}

func (processor *RedisTaskProcessor) ProcessTaskDeleteTopicsFromPosts(ctx context.Context, task *asynq.Task,
) error {
	var payload PayloadSendDeleteTopicsFromPostsTask
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w")
	}

	// convert topic_ids to object ids
	objectIDs := ConvertIdsToObjectIds(payload.TopicIds)

	return DeleteTopicsFromPostsAndUpdatePost(processor.handler, objectIDs, &gin.Context{})
}

// Helper Method to convert Multiple Ids to Object Ids without throwing error
func ConvertIdsToObjectIds(listIds []string) []primitive.ObjectID {

	hexIds := make([]primitive.ObjectID, len(listIds))

	for i, id := range listIds {
		hexIds[i], _ = primitive.ObjectIDFromHex(id)
	}

	return hexIds
}
