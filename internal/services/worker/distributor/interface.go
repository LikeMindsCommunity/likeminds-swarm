package distributor

import "github.com/hibiken/asynq"

// FeedTaskDistributor | Interface for feed background task distributor
type FeedTaskDistributor interface {
	DeleteTopicsFromPosts(topicIds []string, opts ...asynq.Option) error
}
