#!/bin/bash

# Wait for NATS to be ready
echo "Checking NATS connection..."
until nc -z localhost 4222; do 
    echo "Waiting for NATS to be ready..."
    sleep 1
done

echo "NATS is ready. Creating KV buckets..."

# Create KV buckets for auth service
# Note: auth-session bucket removed as we're using JWT tokens instead of sessions
nats kv add reset-token --server nats://localhost:4222 || echo "reset-token bucket already exists"
nats kv add verify-token --server nats://localhost:4222 || echo "verify-token bucket already exists"

# Create kloudlite-events stream if not exists
if ! nats --server nats://localhost:4222 stream info kloudlite-events &>/dev/null; then
    echo "Creating kloudlite-events stream..."
    nats --server nats://localhost:4222 \
        stream add kloudlite-events \
        --subjects "kloudlite.events.>" \
        --retention limits \
        --storage file \
        --max-msgs=-1 \
        --max-bytes=-1 \
        --max-age=0s \
        --compression=s2 \
        --discard=old \
        --defaults || {
            echo "Failed to create kloudlite-events stream"
            exit 1
        }
    echo "Stream created successfully"
else
    echo "Stream kloudlite-events already exists"
fi

# Create notifications stream if not exists
if ! nats --server nats://localhost:4222 stream info notifications &>/dev/null; then
    echo "Creating notifications stream..."
    nats --server nats://localhost:4222 \
        stream add notifications \
        --subjects "notifications.>" \
        --retention limits \
        --storage file \
        --max-msgs=-1 \
        --max-bytes=-1 \
        --max-age=24h \
        --compression=s2 \
        --discard=old \
        --defaults || {
            echo "Failed to create notifications stream"
            exit 1
        }
    echo "Notifications stream created successfully"
else
    echo "Stream notifications already exists"
fi

echo "NATS initialization complete!"