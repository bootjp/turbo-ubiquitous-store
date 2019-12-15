#!/usr/bin/env bash

env CGO_ENABLED=0 env GOOS="linux" go build -o dexe distinct_executor/cli/main.go &
env CGO_ENABLED=0 env GOOS="linux" go build -o tus kvs/cli/server.go &

