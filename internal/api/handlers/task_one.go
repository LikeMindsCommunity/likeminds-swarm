package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"github.com/nateshr/likeminds-swarm/internal/api/constants"
	log "github.com/nateshr/likeminds-swarm/internal/services/logging"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const TaskSendDeleteTopicsFromPosts = "task:send_delete_topics_from_posts"

type PayloadSendDeleteTopicsFromPostsTask struct {
	TopicIds []string `json:"topic_ids"`
}

func (distributor *RedisTaskDistributor) DistributeTaskDeleteTopicsFromPosts(
	ctx context.Context,
	payload *PayloadSendDeleteTopicsFromPostsTask,
	opts ...asynq.Option,
) error {
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

func (distributor *RedisTaskDistributor) DistributeXYZ() error {
	return fmt.Errorf("xyz")
}

func (processor *RedisTaskProcessor) ProcessTaskDeleteTopicsFromPosts(
	ctx context.Context,
	task *asynq.Task,
) error {
	var payload PayloadSendDeleteTopicsFromPostsTask
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w")
	}

	// convert topic_ids to object ids
	objectIDs := ConvertIdsToObjectIds(payload.TopicIds)

	return DeleteTopicsFromPostsAndUpdatePost(processor.handler, objectIDs)
}

// Helper Method to convert Multiple Ids to Object Ids without throwing error
func ConvertIdsToObjectIds(listIds []string) []primitive.ObjectID {

	hexIds := make([]primitive.ObjectID, len(listIds))

	for i, id := range listIds {
		hexIds[i], _ = primitive.ObjectIDFromHex(id)
	}

	return hexIds
}

// deletes topics from posts and update the post in ES index
func DeleteTopicsFromPostsAndUpdatePost(handlers *FeedHandlers, topicIDs []primitive.ObjectID) error {
	fmt.Println("DeleteTopicsFromPostsAndUpdatePost task started: ", time.Now())
	// Create a filter to find posts to be updated
	filter := bson.M{
		"topic_ids": bson.M{
			"$in": topicIDs,
		},
	}

	// find posts based on the filter
	postResults, err := handlers.postHelper.FindPostHelper(filter, gin.H{})

	// extract postIds from the postResults
	postIDs, postIDsString := []primitive.ObjectID{}, []string{}
	for _, post := range postResults {
		postIDs = append(postIDs, post.ID)
		postIDsString = append(postIDsString, post.ID.String())
	}

	// update the filter to update posts by postIds
	filter = bson.M{
		"_id": bson.M{
			"$in": postIDs,
		},
	}

	// Create an update to pull the specified topic IDs from the array
	update := bson.M{
		"$pull": bson.M{
			"topic_ids": bson.M{
				"$in": topicIDs,
			},
		},
	}

	// deletes the topics from posts based on the passed filter and update query
	err = handlers.postHelper.UpdateManyPostsHelper(filter, update, true)
	if err != nil {
		return err
	}

	// fetch the updated posts and update in ES
	updatedPosts, err := handlers.postHelper.FindPostHelper(filter, gin.H{})

	for _, postData := range updatedPosts {
		// update post data in elastic search
		err = handlers.esHelper.IndexDocument(ParsePostIndexData(&postData), postData.ID.Hex(), constants.PostIndexName)
		if err != nil {
			log.Error(err.Error())
		}
	}

	fmt.Println("DeleteTopicsFromPostsAndUpdatePost task ended: ", time.Now())
	return nil
}
