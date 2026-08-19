package distributor

import (
	"github.com/LikeMindsCommunity/likeminds-swarm/internal/services/logging"
	"github.com/LikeMindsCommunity/likeminds-swarm/internal/services/worker"
	"github.com/hibiken/asynq"
)

type RedisTaskDistributor struct {
	client *asynq.Client
}

// NewTaskDistributor | Creates a new task distributor for feed background tasks
func NewTaskDistributor() FeedTaskDistributor {

	// create a new redis client
	redisClientOpts := worker.GetRedisClientOpts()

	// create a new asynq client
	client := asynq.NewClient(redisClientOpts)

	// enqueue a test task to check the connection with the broker
	_, err := client.Enqueue(asynq.NewTask(worker.BrokerConnectionTest, []byte("test payload")))
	if err != nil {
		logging.Fatal("Cannot enqueue task, error with redis broker: ", err)
	}

	return &RedisTaskDistributor{
		client: client,
	}
}
