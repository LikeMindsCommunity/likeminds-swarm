package cache

import (
	"github.com/go-redis/redis/v7"
	"github.com/nateshr/likeminds-swarm/internal/services/environment"

	log "github.com/nateshr/likeminds-swarm/internal/services/logging"
)

// InitRedis | initialises a connection pool to redis cache
func InitRedis() *redis.Client {
	//Initializing Redis
	dsn := environment.GoDotEnvVariable("REDIS_DSN")
	if len(dsn) == 0 {
		dsn = "localhost:6379"
	}

	client := redis.NewClient(&redis.Options{
		Addr: dsn,
	})

	_, err := client.Ping().Result()
	if err != nil {
		panic(err)
	}

	log.Info("Cache(Redis): Successfully connected and pinged.")

	return client
}
