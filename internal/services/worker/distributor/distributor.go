package distributor

import (
	"github.com/hibiken/asynq"
	"github.com/nateshr/likeminds-swarm/internal/services/worker"
)

// FeedTaskDistributor | Interface for feed background task distributor
type FeedTaskDistributor interface {
	DeleteTopicsFromPosts(topicIds []string, opts ...asynq.Option) error
}

type RedisTaskDistributor struct {
	client *asynq.Client
}

// NewTaskDistributor | Creates a new task distributor for feed background tasks
func NewTaskDistributor() FeedTaskDistributor {

	// create a new redis client
	redisClientOpts := worker.InitRedisBrokerClient()

	// create a new asynq client
	client := asynq.NewClient(redisClientOpts)

	return &RedisTaskDistributor{
		client: client,
	}
}
