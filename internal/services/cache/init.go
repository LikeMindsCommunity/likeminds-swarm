package cache

import (
	"github.com/LikeMindsCommunity/likeminds-swarm/internal/services/environment"
	"github.com/LikeMindsCommunity/likeminds-swarm/internal/services/logging"
	"github.com/go-redis/redis/v7"
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

	client.Options().Password = password

	// disabling tls config as using private hosted DNS zone in azure
	// client.Options().TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}

	_, err := client.Ping().Result()
	if err != nil {
		panic(err)
	}

	logging.Info("Cache(Redis): Successfully connected and pinged.")

	return client
}
