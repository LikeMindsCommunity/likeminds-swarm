package processor

import (
	"context"

	"github.com/hibiken/asynq"
)

// FeedTaskProcessor | Interface for feed background tasks
type FeedTaskProcessor interface {
	Run() error
	connectionTest(ctx context.Context, task *asynq.Task) error
	sendWebhookRequestWithPayload(ctx context.Context, task *asynq.Task) error
	triggerPostCreationWebhook(ctx context.Context, task *asynq.Task) error
	triggerPostLikedWebhook(ctx context.Context, task *asynq.Task) error
	triggerPostPinnedWebhook(ctx context.Context, task *asynq.Task) error
	triggerPostTaggedWebhook(ctx context.Context, task *asynq.Task) error
	triggerCommentAddedWebhook(ctx context.Context, task *asynq.Task) error
	triggerCommentReactWebhook(ctx context.Context, task *asynq.Task) error
	triggerCommentTaggedWebhook(ctx context.Context, task *asynq.Task) error
	createPostBackgroundTasks(ctx context.Context, task *asynq.Task) error
	editPostBackgroundTasks(ctx context.Context, task *asynq.Task) error
	deletePostBackgroundTasks(ctx context.Context, task *asynq.Task) error
	sendNotification(ctx context.Context, task *asynq.Task) error
}
