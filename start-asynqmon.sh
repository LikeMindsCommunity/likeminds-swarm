
#!/bin/bash

set -a
. ./cmd/server/config/.env
set +a

echo "Redis Address: $ASYNQ_BROKER_ADDRESS"
echo "Starting Asynqmon..."

# Start the actual service
exec asynqmon \
  --port=8080 \
  --redis-addr="$ASYNQ_BROKER_ADDRESS" \
  --redis-password="$ASYNQ_BROKER_PASSWORD" \
  --read-only=true
