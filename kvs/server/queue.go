package server

import (
	"os"

	"github.com/go-redis/redis"
)

type QueueManager struct {
	QueuePrimary   *redis.Client
	QueueSecondary *redis.Client
}
type QueueManagerI interface {
	Enqueue(u UpdateQueue) bool
	Drain() bool
}

func NewQueueManager() *QueueManager {
	return &QueueManager{
		QueuePrimary:   redis.NewClient(&redis.Options{Addr: os.Getenv("PRIMARY_REDIS_HOST")}),
		QueueSecondary: redis.NewClient(&redis.Options{Addr: os.Getenv("PRIMARY_REDIS_HOST")}),
	}
}

func (q QueueManager) Enqueue(u UpdateQueue) bool {

	return true
}
func (q QueueManager) Drain(u UpdateQueue) bool {
	return true
}

type UpdateQueue struct {
	Data     string
	UpdateAt uint64
}
