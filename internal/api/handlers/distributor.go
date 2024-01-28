package handlers

import (
	"context"

	"github.com/hibiken/asynq"
)

type TaskDistributer interface {
	DistributeTaskDeleteTopicsFromPosts(ctx context.Context, payload *PayloadSendDeleteTopicsFromPostsTask, opts ...asynq.Option) error
	DistributeXYZ() error
}

type RedisTaskDistributor struct {
	client *asynq.Client
}

func NewRedisTaskDistributor(redisOpt asynq.RedisClientOpt) TaskDistributer {
	client := asynq.NewClient(redisOpt)
	return &RedisTaskDistributor{
		client: client,
	}
}
