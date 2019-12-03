package kvs

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
	Length() int
}

func NewQueueManager() *QueueManager {
	return &QueueManager{
		QueuePrimary:   redis.NewClient(&redis.Options{Addr: os.Getenv("PRIMARY_REDIS_HOST")}),
		QueueSecondary: redis.NewClient(&redis.Options{Addr: os.Getenv("SECONDARY_REDIS_HOST")}),
	}
}

func (q QueueManager) Enqueue(u UpdateQueue) bool {

	return true
}

func (q QueueManager) Drain() bool {
	return true
}
func (q QueueManager) Length() int {
	return 0
}

type UpdateQueue struct {
	Data     string
	UpdateAt uint64
}
