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

func (distributor *RedisTaskDistributor) DistributeTaskDeleteTopicsFromPosts(payload *PayloadSendDeleteTopicsFromPostsTask, opts ...asynq.Option) error {
	// wrap into payload
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}

	// enqueue the task
	taskInfo, err := EnqueueBackgroundTask(distributor.client, TaskSendDeleteTopicsFromPosts, jsonPayload, opts...)
	if err != nil {
		return fmt.Errorf("failed to enqueue task %w", err)
	}
	fmt.Print("task id", taskInfo.ID)

	return nil
}

func (processor *RedisTaskProcessor) ProcessTaskDeleteTopicsFromPosts(ctx context.Context, task *asynq.Task) error {
	var payload PayloadSendDeleteTopicsFromPostsTask
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
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
