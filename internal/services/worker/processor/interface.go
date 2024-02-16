package processor

import (
	"context"

	"github.com/hibiken/asynq"
)

// FeedTaskProcessor | Interface for feed background tasks
type FeedTaskProcessor interface {
	Run() error
	ConnectionTest(ctx context.Context, task *asynq.Task) error
	DeleteTopicsFromPosts(ctx context.Context, task *asynq.Task) error
}
