#!/usr/bin/env bash

env PRIMARY_REDIS_HOST=localhost:63790 env SECONDARY_REDIS_HOST=localhost:63791 env MASTER_REDIS_HOST=localhost:6379 go run distinct_executor/cli/main.go &
env PRIMARY_REDIS_HOST=localhost:63790 env SECONDARY_REDIS_HOST=localhost:63791 env MASTER_REDIS_HOST=localhost:6379 go run kvs/cli/server.go
kill $!
