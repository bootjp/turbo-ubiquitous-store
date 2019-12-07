package kvs

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"github.com/gomodule/redigo/redis"
)

type QueueManager struct {
	QueuePrimary   redis.Conn
	QueueSecondary redis.Conn
	Queue          []UpdateQueue
	NextNode       *NextNode // Next node when replacing kvs node
	mutex          *sync.Mutex
	draining       bool
	executeDequeue bool
	log            *log.Logger
}
type NextNode struct {
	Host net.TCPAddr
}

type QueueManagerI interface {
	Enqueue(u UpdateQueue)
	Forward()
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
		NextNode:       nil,
		mutex:          &sync.Mutex{},
		log:            log.New(os.Stdout, "queue_manager: ", log.Ltime),
	}
}

func (q *QueueManager) Length() int {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	return len(q.Queue)
}

const updateQueueKey = "tus_queue"

func (q *QueueManager) Forward() {
	for {
		q.log.Println("queue current", len(q.Queue))
		if q.Length() == 0 {
			_, err := q.QueuePrimary.Do("PING")
			if err != nil {
				log.Println(err)
			}
			_, err = q.QueueSecondary.Do("PING")
			if err != nil {
				log.Println(err)
			}
			time.Sleep(1 * time.Second)
			q.log.Println("continue wait queue")
			continue
		}
		q.mutex.Lock()

		jsonBytes, err := json.Marshal(q.Queue[0])
		if err != nil {
			fmt.Println("JSON Marshal error:", err)
		}
		data := string(jsonBytes)
		_, err = q.QueuePrimary.Do("LPUSH", updateQueueKey, data)
		if err != nil {
			log.Println(err)
		}
		_, err = q.QueueSecondary.Do("LPUSH", updateQueueKey, data)
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
}

type UpdateQueue struct {
	Data     string `json:"data"`
	Key      string `json:"key"`
	UpdateAt int64  `json:"update_at"`
}
