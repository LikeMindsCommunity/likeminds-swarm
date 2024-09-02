package scheduler

import (
	"github.com/hibiken/asynq"
	"github.com/nateshr/likeminds-swarm/internal/services/logging"
	"github.com/nateshr/likeminds-swarm/internal/services/worker"
)

// Task to register the crom job tasks
func (taskScheduler *RedisTaskScheduler) Run() error {
	logging.Info(`Starting the redis task scheduler`)

	// Community default feed task
	communityDefaultFeedTask := asynq.NewTask(worker.TaskAsyncCommunityDefaultFeed, []byte(""))
	taskScheduler.scheduler.Register("@every 25m", communityDefaultFeedTask)

	return taskScheduler.scheduler.Run()
}

type RedisTaskScheduler struct {
	scheduler *asynq.Scheduler
}

// NewTaskscheduler | Creates a new task scheduler for feed scheduled background tasks
func NewTaskScheduler() FeedTaskScheduler {

	// create a new redis conn opts
	redisConnOpts := worker.GetRedisConnOpts()

	// get AsynQ scheduler configurations
	config := worker.GetSchedulerConfigurations()

	// create a new asynq client
	scheduler := asynq.NewScheduler(redisConnOpts, config)

	return &RedisTaskScheduler{
		scheduler: scheduler,
	}
}
