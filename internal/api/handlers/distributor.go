package handlers

import (
	"github.com/hibiken/asynq"
)

type FeedTaskDistributor interface {
	DistributeTaskDeleteTopicsFromPosts(payload *PayloadSendDeleteTopicsFromPostsTask, opts ...asynq.Option) error
}

type RedisTaskDistributor struct {
	client *asynq.Client
}

func NewRedisTaskDistributor(redisOpt asynq.RedisClientOpt) FeedTaskDistributor {
	client := asynq.NewClient(redisOpt)
	return &RedisTaskDistributor{
		client: client,
	}
}
