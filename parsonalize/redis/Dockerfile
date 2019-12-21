FROM golang:alpine AS build
ENV GO111MODULE=on

COPY ./ $GOPATH/src/github.com/bootjp/turbo-ubiquitous-store/parsonalize/redis
RUN \
  cd $GOPATH/src/github.com/bootjp/turbo-ubiquitous-store/parsonalize/redis && \
  GOOS=linux CGO_ENABLED=0 go build -a -o out app.go && \
  cp out /app

FROM alpine
ENV PRIMARY_REDIS_HOST=host.docker.internal:63790
ENV SECONDARY_REDIS_HOST=host.docker.internal:63791
ENV MASTER_REDIS_HOST=host.docker.internal:6379
ENV SLAVE_REDIS_HOST=host.docker.internal:6380
RUN apk add --no-cache tzdata ca-certificates
COPY --from=build /app /app

CMD ["/app"]