package cache

import (
	"crypto/tls"

	"github.com/go-redis/redis/v7"
	"github.com/nateshr/likeminds-swarm/internal/services/environment"
	"github.com/nateshr/likeminds-swarm/internal/services/logging"
)

// InitRedis | initialises a connection pool to redis cache
func InitRedis() *redis.Client {
	//Initializing Redis
	dsn := environment.GoDotEnvVariable("REDIS_DSN")
	password := environment.GoDotEnvVariable("REDIS_PASSWORD")
	if len(dsn) == 0 {
		dsn = "localhost:6379"
	}

	client := redis.NewClient(&redis.Options{
		Addr: dsn,
	})

	serverEnviornment := environment.GoDotEnvVariable("SERVER_ENVIRONMENT")
	if serverEnviornment == "load" {
		client.Options().Password = password
		client.Options().TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	}

	_, err := client.Ping().Result()
	if err != nil {
		panic(err)
	}

	logging.Info("Cache(Redis): Successfully connected and pinged.")

	return client
}
