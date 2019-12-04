package kvs

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/gomodule/redigo/redis"
)

type QueueManager struct {
	QueuePrimary   redis.Conn
	QueueSecondary redis.Conn
	Queue          []UpdateQueue
	mutex          *sync.Mutex
	draining       bool
	executeDequeue bool
}
type QueueManagerI interface {
	Enqueue(u UpdateQueue)
	Drain() bool
	Dequeue()
}

func NewQueueManager() *QueueManager {
	pconn, err := redis.Dial("tcp", os.Getenv("PRIMARY_REDIS_HOST"))
	if err != nil {
		log.Println(err)
		log.Fatal("failed to connect primary storage.")
	}

	sconn, err := redis.Dial("tcp", os.Getenv("SECONDARY_REDIS_HOST"))
	if err != nil {
		log.Println(err)
		log.Fatal("failed to connect secondary queue .")
	}

	return &QueueManager{
		QueuePrimary:   pconn,
		QueueSecondary: sconn,
		Queue:          make([]UpdateQueue, 0),
		mutex:          &sync.Mutex{},
	}
}

func (q *QueueManager) Dequeue() {
	fmt.Println("welcome deque")
	for {
		fmt.Println("queue current", len(q.Queue))
		q.mutex.Lock()
		if len(q.Queue) == 0 {
			q.mutex.Unlock()
			_, err := q.QueuePrimary.Do("PING")
			if err != nil {
				log.Println(err)
			}
			_, err = q.QueueSecondary.Do("PING")
			if err != nil {
				log.Println(err)
			}
			time.Sleep(1 * time.Second)
			fmt.Println("continue wait queue")
			continue
		}
		fmt.Println("dequeue")

		data := q.Queue[0]

		jsonBytes, err := json.Marshal(data)
		if err != nil {
			fmt.Println("JSON Marshal error:", err)
		}
		_, err = q.QueuePrimary.Do("LPUSH", "tus_queue", string(jsonBytes))
		if err != nil {
			log.Println(err)
		}
		_, err = q.QueueSecondary.Do("LPUSH", "tus_queue", string(jsonBytes))
		if err != nil {
			log.Println(err)
		}

		q.Queue = q.Queue[1:]
		q.mutex.Unlock()
	}
}

func (q *QueueManager) Enqueue(u UpdateQueue) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	q.Queue = append(q.Queue, u)
	fmt.Println(q.Queue)
}

func (q *QueueManager) Drain() bool {
	if !q.draining {
		return false
	}
	q.mutex.Lock()
	defer q.mutex.Unlock()
	q.draining = true
	return true
}

type UpdateQueue struct {
	Data     string `json:"data"`
	Key      string `json:"key"`
	UpdateAt int64  `json:"update_at"`
}
