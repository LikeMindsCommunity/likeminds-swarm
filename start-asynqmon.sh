
#!/bin/bash

# redis-cli -h likeminds-staging-redis-swarm-worker.privatelink.redis.cache.windows.net -p 6379 -a oWjJkbKCdZAS5XBLTNUYlTlBGeUCuJc1aAzCaLPiuBk=
# Optional: Echo to see what it's using (for debug, not in prod)
echo "Redis Address: $ASYNQ_BROKER_ADDRESS"
echo "Starting Asynqmon..."

# Start the actual service
exec asynqmon \
  --port=8080 \
  --redis-addr="$ASYNQ_BROKER_ADDRESS" \
  --redis-password="$ASYNQ_BROKER_PASSWORD" \
  --read-only=true
