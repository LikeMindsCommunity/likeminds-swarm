package distributor

import (
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/nateshr/likeminds-swarm/internal/api/enums"
	"github.com/nateshr/likeminds-swarm/internal/api/responses"
	"github.com/nateshr/likeminds-swarm/internal/services/worker"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Task Distributor for "task:SendWebhookRequestWithPayload"
func (distributor *RedisTaskDistributor) SendWebhookRequestWithPayload(apiKey string, url string, payload *responses.WebhookPayload,
	webhookType string, secret string, opts ...asynq.Option) error {

	// create task payload
	taskPayload := worker.PayloadSendWebhookRequestWithPayload{
		ApiKey:      apiKey,
		Url:         url,
		Payload:     *payload,
		WebhookType: webhookType,
		Secret:      secret,
	}

	// Set max retry limit for the task
	opts = append(opts, asynq.MaxRetry(enums.WebhookRetryLimit))

	jsonPayload, err := json.Marshal(taskPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}

	_, err = worker.EnqueueTaskToQueue(distributor.client, worker.TaskSendWebhookRequestWithPayload, jsonPayload, opts...)
	if err != nil {
		return fmt.Errorf("failed to enqueue task %v", err)
	}

	return nil
}

// Task Distributor for "task:TriggerPostCreationWebhook"
func (distributor *RedisTaskDistributor) TriggerPostCreationWebhook(postId string, apiKey string, opts ...asynq.Option) error {

	if postId == "" || apiKey == "" {
		return fmt.Errorf("postId or apiKey is missing")
	}

	payload := worker.PayloadTriggerPostCreationWebhook{
		PostId: postId,
		ApiKey: apiKey,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}

	_, err = worker.EnqueueTaskToQueue(distributor.client, worker.TaskTriggerPostCreationWebhook, jsonPayload, opts...)
	if err != nil {
		return fmt.Errorf("failed to enqueue task %v", err)
	}

	return nil
}

// Task Distributor for "task:TriggerPostLikedWebhook"
func (distributor *RedisTaskDistributor) TriggerPostLikedWebhook(postId string, userId string, apiKey string, opts ...asynq.Option) error {

	if postId == "" || userId == "" || apiKey == "" {
		return fmt.Errorf("postId, userId or apiKey is missing")
	}

	payload := worker.PayloadTriggerPostLikedWebhook{
		PostId: postId,
		UserId: userId,
		ApiKey: apiKey,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}

	_, err = worker.EnqueueTaskToQueue(distributor.client, worker.TaskTriggerPostLikedWebhook, jsonPayload, opts...)
	if err != nil {
		return fmt.Errorf("failed to enqueue task %v", err)
	}

	return nil
}

// Task Distributor for "task:TriggerPostPinnedWebhook"
func (distributor *RedisTaskDistributor) TriggerPostPinnedWebhook(postId string, userId string, apiKey string, opts ...asynq.Option) error {

	if postId == "" || userId == "" || apiKey == "" {
		return fmt.Errorf("postId, userId or apiKey is missing")
	}

	payload := worker.PayloadTriggerPostPinnedWebhook{
		PostId: postId,
		UserId: userId,
		ApiKey: apiKey,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}

	_, err = worker.EnqueueTaskToQueue(distributor.client, worker.TaskTriggerPostPinnedWebhook, jsonPayload, opts...)
	if err != nil {
		return fmt.Errorf("failed to enqueue task %v", err)
	}

	return nil
}

// Task Distributor for "task:TriggerPostTaggedWebhook"
func (distributor *RedisTaskDistributor) TriggerPostTaggedWebhook(postId string, userIds []string, apiKey string, opts ...asynq.Option) error {

	if postId == "" || len(userIds) == 0 || apiKey == "" {
		return fmt.Errorf("postId, userIds or apiKey is missing")
	}

	payload := worker.PayloadTriggerPostTaggedWebhook{
		PostId:  postId,
		UserIds: userIds,
		ApiKey:  apiKey,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}

	_, err = worker.EnqueueTaskToQueue(distributor.client, worker.TaskTriggerPostTaggedWebhook, jsonPayload, opts...)
	if err != nil {
		return fmt.Errorf("failed to enqueue task %v", err)
	}

	return nil
}

// Task Distributor for "task:TriggerCommentAddedWebhook"
func (distributor *RedisTaskDistributor) TriggerCommentAddedWebhook(commentId string, apiKey string, opts ...asynq.Option) error {

	if commentId == "" || apiKey == "" {
		return fmt.Errorf("commentId or apiKey is missing")
	}

	payload := worker.PayloadTriggerCommentAddedWebhook{
		CommentId: commentId,
		ApiKey:    apiKey,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}

	_, err = worker.EnqueueTaskToQueue(distributor.client, worker.TaskTriggerCommentAddedWebhook, jsonPayload, opts...)

	if err != nil {
		return fmt.Errorf("failed to enqueue task %v", err)
	}

	return nil
}

// Task Distributor for "task:TriggerCommentReactWebhook"
func (distributor *RedisTaskDistributor) TriggerCommentReactWebhook(commentId string, userId string, apiKey string, opts ...asynq.Option) error {

	if commentId == "" || userId == "" || apiKey == "" {
		return fmt.Errorf("commentId, userId or apiKey is missing")
	}

	payload := worker.PayloadTriggerCommentReactWebhook{
		CommentId: commentId,
		UserId:    userId,
		ApiKey:    apiKey,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}

	_, err = worker.EnqueueTaskToQueue(distributor.client, worker.TaskTriggerCommentReactWebhook, jsonPayload, opts...)
	if err != nil {
		return fmt.Errorf("failed to enqueue task %v", err)
	}

	return nil
}

// Task Distributor for "task:TriggerCommentTaggedWebhook"
func (distributor *RedisTaskDistributor) TriggerCommentTaggedWebhook(commentId string, userIds []string, apiKey string, opts ...asynq.Option) error {

	if commentId == "" || len(userIds) == 0 || apiKey == "" {
		return fmt.Errorf("commentId, userIds or apiKey is missing")
	}

	payload := worker.PayloadTriggerCommentTaggedWebhook{
		CommentId: commentId,
		UserIds:   userIds,
		ApiKey:    apiKey,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}

	_, err = worker.EnqueueTaskToQueue(distributor.client, worker.TaskTriggerCommentTaggedWebhook, jsonPayload, opts...)
	if err != nil {
		return fmt.Errorf("failed to enqueue task %v", err)
	}

	return nil
}

// Task Distributor for "task:TaskTriggerCreatePost"
func (distributor *RedisTaskDistributor) EnqueueCreatePostBackgroundTasks(postId string, opts ...asynq.Option) error {

	if postId == "" {
		return fmt.Errorf("missing Post ID")
	}

	payload := worker.PayloadPost{
		PostID: postId,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}

	_, err = worker.EnqueueTaskToQueue(distributor.client, worker.TaskCreatePostBackgroundTasks, jsonPayload, opts...)
	if err != nil {
		return fmt.Errorf("failed to enqueue task %v", err)
	}

	return nil
}

// Task Distributor for "task:TaskTriggerEditPost"
func (distributor *RedisTaskDistributor) EnqueueEditPostBackgroundTasks(postId string, opts ...asynq.Option) error {

	if postId == "" {
		return fmt.Errorf("missing Post ID")
	}

	payload := worker.PayloadPost{
		PostID: postId,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}

	_, err = worker.EnqueueTaskToQueue(distributor.client, worker.TaskEditPostBackgroundTasks, jsonPayload, opts...)
	if err != nil {
		return fmt.Errorf("failed to enqueue task %v", err)
	}

	return nil
}

// Task Distributor for "task:TaskTriggerDeletePost"
func (distributor *RedisTaskDistributor) EnqueueDeletePostBackgroundTasks(postId string, opts ...asynq.Option) error {

	if postId == "" {
		return fmt.Errorf("missing Post ID")
	}

	payload := worker.PayloadPost{
		PostID: postId,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}

	_, err = worker.EnqueueTaskToQueue(distributor.client, worker.TaskDeletePostBackgroundTasks, jsonPayload, opts...)
	if err != nil {
		return fmt.Errorf("failed to enqueue task %v", err)
	}

	return nil
}

// Task Distributor for "task:TaskSendNotification"
func (distributor *RedisTaskDistributor) EnqueueSendNotification(activityID primitive.ObjectID, platformCode string, versionCode string, opts ...asynq.Option) error {

	if activityID == primitive.NilObjectID || platformCode == "" || versionCode == "" {
		return fmt.Errorf("missing activity ID, platform code or version code")
	}

	payload := worker.PayloadSendNotification{
		ActivityID:   activityID,
		PlatformCode: platformCode,
		VersionCode:  versionCode,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}

	_, err = worker.EnqueueTaskToQueue(distributor.client, worker.TaskSendNotification, jsonPayload, opts...)
	if err != nil {
		return fmt.Errorf("failed to enqueue task %v", err)
	}

	return nil
}
