package main

import (
	"log"
	"os"
	"sync"
	"time"

	"github.com/gomodule/redigo/redis"
)

type QueueNodes struct {
	QueuePrimary   redis.Conn
	QueueSecondary redis.Conn
	isRunning      bool
	mutex          sync.Mutex
	log            *log.Logger
}
type DistinctInsert interface {
	Execute()
	Running() bool
}

func (n *QueueNodes) Running() bool {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	return n.isRunning
}

func NewDistinctExecutor() *QueueNodes {
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

	return &QueueNodes{
		QueuePrimary:   pconn,
		QueueSecondary: sconn,
		isRunning:      false,
		log:            log.New(os.Stdout, "distinct_executor", log.Ltime),
	}

}

const updateQueueKey = "tus_queue"

func (n *QueueNodes) Execute() {
	if n.Running() {
		return
	}
	n.mutex.Lock()
	n.isRunning = true
	n.mutex.Unlock()
	for {
		pqlen, err := redis.Int(n.QueuePrimary.Do("LLEN", updateQueueKey))
		if err != nil {
			log.Println(err)
		}
		sqlen, err := redis.Int(n.QueueSecondary.Do("LLEN", updateQueueKey))
		if err != nil {
			log.Println(err)
		}
		var queueLength int
		if pqlen < sqlen {
			queueLength = sqlen
		} else {
			queueLength = pqlen
		}
		for queueLength == 0 {

			time.Sleep(1 * time.Second)
			continue

		}
		// this is has queue
		// todo here dequeue and insert kvs
	}
}

func main() {
	os.Setenv("PRIMARY_REDIS_HOST", "localhost:63790")
	os.Setenv("SECONDARY_REDIS_HOST", "localhost:63791")
	NewDistinctExecutor().Execute()
}
