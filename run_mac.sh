#!/usr/bin/env bash

env PORT=8888 go run parsonalize/app.go &
p1=$!
env PORT=8800 go run parsonalize/app.go &
p2=$!

env PRIMARY_REDIS_HOST=localhost:63790 env SECONDARY_REDIS_HOST=localhost:63791 env MASTER_REDIS_HOST=localhost:6379 go run distinct_executor/cli/main.go &
PID=$!
env PRIMARY_REDIS_HOST=localhost:63790 env SECONDARY_REDIS_HOST=localhost:63791 env MASTER_REDIS_HOST=localhost:6379 go run kvs/cli/server.go
kill $PID $p1 $2
