package kvs

import (
	"os"
	"sync"

	"github.com/go-redis/redis"
)

type QueueManager struct {
	QueuePrimary   *redis.Client
	QueueSecondary *redis.Client
	Queue          []UpdateQueue
	mutex          *sync.Mutex
	draining       bool
}
type QueueManagerI interface {
	Enqueue(u UpdateQueue) bool
	Drain() bool
	Length() int
}

const defaultQueueLength = 5000

func NewQueueManager() *QueueManager {
	return &QueueManager{
		QueuePrimary:   redis.NewClient(&redis.Options{Addr: os.Getenv("PRIMARY_REDIS_HOST")}),
		QueueSecondary: redis.NewClient(&redis.Options{Addr: os.Getenv("SECONDARY_REDIS_HOST")}),
		Queue:          make([]UpdateQueue, defaultQueueLength),
		mutex:          &sync.Mutex{},
	}
}

func (q QueueManager) Enqueue(u UpdateQueue) bool {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	q.Queue = append(q.Queue, u)
	return true
}

func (q QueueManager) Drain() bool {
	if !q.draining {
		return false
	}
	q.mutex.Lock()
	defer q.mutex.Unlock()
	q.draining = true
	return true
}
func (q QueueManager) Length() int {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	return len(q.Queue)
}

type UpdateQueue struct {
	Data     string
	UpdateAt int64
}
