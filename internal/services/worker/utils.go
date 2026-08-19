package worker

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/LikeMindsCommunity/likeminds-swarm/internal/services/environment"
	"github.com/LikeMindsCommunity/likeminds-swarm/internal/services/logging"
	"github.com/hibiken/asynq"
)

// Queue Names
const (
	DefaultQueue string = "default" // Priority -> 10
)

// getAllQueues | Returns a map of all queue names with their priority
func getAllQueues() map[string]int {

	// Queues priority map
	queuePriortiyMap := map[string]int{
		DefaultQueue: 10,
	}

	return queuePriortiyMap
}

// getValidQueuesMap | Return valid queue priority map from qNames else return all queues
func getValidQueuesMap(qNames []string) map[string]int {

	userQueueMap := map[string]int{}

	defaultQueues := getAllQueues()

	// If qNames are passed, return valid queue priority map
	if len(qNames) > 0 {
		for _, qname := range qNames {
			if _, exists := defaultQueues[qname]; exists {
				userQueueMap[qname] = defaultQueues[qname]
			} else {
				logging.Error("Invalid queue name: ", qname)
			}
		}
	}

	if len(userQueueMap) > 0 {
		logging.Info("Listening to user queues: ", userQueueMap)
		return userQueueMap
	}

	logging.Info("Listening to default queues: ", defaultQueues)
	return defaultQueues
}

// GetRedisClientOpts | Returns the redis client options for the Asynq client
func GetRedisClientOpts() *asynq.RedisClientOpt {

	brokerAddress := environment.GoDotEnvVariable("ASYNQ_BROKER_ADDRESS")
	if brokerAddress == "" {
		brokerAddress = "localhost:6379"
	}

	redisClientOpt := asynq.RedisClientOpt{
		Addr: brokerAddress,
	}

	redisClientOpt.Password = environment.GoDotEnvVariable("ASYNQ_BROKER_PASSWORD")

	// disabling tls config as using private hosted DNS zone in azure
	// redisClientOpt.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}

	return &redisClientOpt
}

// EnqueueTaskToQueue | Enqueues a task to the queue with the provided payload and options
func EnqueueTaskToQueue(client *asynq.Client, taskName string, taskPayload []byte, opts ...asynq.Option) (*asynq.TaskInfo, error) {

	// Default options for tasks
	defaultOpts := []asynq.Option{
		asynq.MaxRetry(1),                 // max number of times the task can be retried
		asynq.Retention(10 * time.Minute), // how long to keep the task result in completed state
		asynq.Queue(DefaultQueue),         // which queue to enqueue the task to
	}

	// creates a new task with the provided payload and DEFAULT OPTIONS
	task := asynq.NewTask(taskName, taskPayload, defaultOpts...)

	// enqueues the task to the queue with task specific options (Overrides default options)
	taskInfo, err := client.Enqueue(task, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to enqueue task %w", err)
	}

	logging.Info(fmt.Sprintf(`Task Enqueued | taskId: %s ; taskName: %s ; taskPayload: %s ; taskQueue: %s ;`,
		taskInfo.ID, taskInfo.Type, string(taskInfo.Payload), taskInfo.Queue))

	return taskInfo, nil
}

// loggingMiddleware | Processor middleware to log task lifescycle (start, end & error)
func LoggingMiddleware(h asynq.Handler) asynq.Handler {
	return asynq.HandlerFunc(func(ctx context.Context, task *asynq.Task) error {
		start := time.Now()

		// Get task metadata
		taskId, _ := asynq.GetTaskID(ctx)
		taskPayload := task.Payload()
		taskName := task.Type()
		taskQueue, _ := asynq.GetQueueName(ctx)

		// Log task received with task metadata
		logging.Info(fmt.Sprintf(`Task Received | taskId: %s ; taskName: %s ; taskPayload: %s ; taskQueue: %s ;`,
			taskId, taskName, string(taskPayload), taskQueue))

		// Process task
		err := h.ProcessTask(ctx, task)
		if err != nil {

			// Log the error with task metadata
			logging.Error(fmt.Sprintf(`Task Failed | taskId: %s ; taskType: %s ; taskPayload: %s ; error: %s ;`,
				taskId, taskName, taskPayload, err.Error()))

			return err
		}

		// Log task finished with task metadata
		logging.Info(fmt.Sprintf(`Task Finished | taskId: %s ; taskName: %s ; taskPayload: %s ; taskQueue: %s ; Finished in %v ;`,
			taskId, taskName, string(taskPayload), taskQueue, time.Since(start)))

		return nil
	})
}

// GetServerConfigurations | Returns configurations for the Asynq server
func GetServerConfigurations(QueueNames []string) asynq.Config {

	// default configurations for the Task Processor
	config := asynq.Config{

		// Using custom logger for logging
		Logger: logging.NewCustomLogger(),

		// Queue names and their priority
		Queues: getValidQueuesMap(QueueNames),

		// How to handle retries for tasks
		RetryDelayFunc: retryDelayFunctionForWebhookTasks,
	}

	// set concurrency from environment variable if present | default: cpu cores count
	concurrency, err := strconv.Atoi(environment.GoDotEnvVariable("ASYNQ_WORKER_CONCURRENCY"))
	if err == nil {
		config.Concurrency = concurrency
	}

	return config
}

func GetSchedulerConfigurations() *asynq.SchedulerOpts {

	// default configurations for the Task Processor
	config := asynq.SchedulerOpts{

		// Using custom logger for logging
		Logger: logging.NewCustomLogger(),
	}

	return &config
}

func retryDelayFunctionForWebhookTasks(n int, e error, t *asynq.Task) time.Duration {

	// If task is webhookRequest
	if t.Type() == TaskSendWebhookRequestWithPayload {

		// Calculate retry delay for webhook tasks (1 -> 60 -> 3600 seconds)
		delayinSeconds := math.Pow(60, float64(n))

		return time.Duration(delayinSeconds) * time.Second
	}

	// Retry delay for all other tasks
	return asynq.DefaultRetryDelayFunc(n, e, t)
}
