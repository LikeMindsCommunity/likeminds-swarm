package handlers

import (
	"github.com/hibiken/asynq"
)

type TaskDistributor interface {
	DistributeTaskDeleteTopicsFromPosts(payload *PayloadSendDeleteTopicsFromPostsTask, opts ...asynq.Option) error
}

type RedisTaskDistributor struct {
	client *asynq.Client
}

func NewRedisTaskDistributor(redisOpt asynq.RedisClientOpt) TaskDistributor {
	client := asynq.NewClient(redisOpt)
	return &RedisTaskDistributor{
		client: client,
	}
}
