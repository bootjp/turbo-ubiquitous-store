#!/usr/bin/env bash

env PORT=8888 go run parsonalize/app.go &
p1=$!
env PORT=8800 go run parsonalize/app.go &
p2=$!

env PRIMARY_REDIS_HOST=172.17.0.1:63790 env SECONDARY_REDIS_HOST=172.17.0.1:63791 env MASTER_REDIS_HOST=172.17.0.1:6379 ./dexe &
k1=$!
env PRIMARY_REDIS_HOST=172.17.0.1:63790 env SECONDARY_REDIS_HOST=172.17.0.1:63791 env MASTER_REDIS_HOST=172.17.0.1:6379 ./tus &
k2=$!

kill p1 p2 k1 k2
