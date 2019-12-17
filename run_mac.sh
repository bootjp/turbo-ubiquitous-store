#!/usr/bin/env bash
export SLAVE_REDIS_HOST=localhost:6380
export PRIMARY_REDIS_HOST=localhost:63790
export SECONDARY_REDIS_HOST=localhost:63791
export MASTER_REDIS_HOST=localhost:6379
export SLAVE_REDIS_HOST=localhost:6380

#env PORT=8888 go run parsonalize/app.go &
#p1=$!
#env PORT=8800 go run parsonalize/app.go &
#p2=$!

go run distinct_executor/cli/main.go &
PID=$!
go run kvs/cli/server.go
kill $PID $p1 $2
