package main

import (
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"

	"github.com/bootjp/turbo-ubiquitous-store/kvs"

	"github.com/gomodule/redigo/redis"
)

const (
	NODE_PRIMARY = iota
	NODE_SECONDARY
)

type Node struct {
	Conn redis.Conn
	Type int
}

type QueueNodes struct {
	ActiveNode     *Node
	queuePrimary   *Node
	queueSecondary *Node
	isRunning      bool
	mutex          sync.Mutex
	log            *log.Logger
}

type MasterNode = Node

type DistinctInsert interface {
	Execute()
	Running() bool
	GetActiveNode() redis.Conn
}

func (n *QueueNodes) Running() bool {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	return n.isRunning
}

func (n *QueueNodes) GetActiveNode() *Node {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	return n.ActiveNode
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
		ActiveNode:     &Node{Conn: pconn, Type: NODE_PRIMARY},
		queuePrimary:   &Node{Conn: pconn, Type: NODE_PRIMARY},
		queueSecondary: &Node{Conn: sconn, Type: NODE_SECONDARY},
		isRunning:      false,
		log:            log.New(os.Stdout, "distinct_executor: ", log.Ltime),
	}

}

const updateQueueKey = "tus_queue"

func (n *QueueNodes) Execute(m *MasterNode) {
	if n.Running() {
		return
	}
	n.mutex.Lock()
	n.isRunning = true
	n.mutex.Unlock()

	for {
		qcount, err := redis.Int(n.ActiveNode.Conn.Do("LLEN", updateQueueKey))
		if err != nil {
			log.Println(err)
			n.mutex.Lock()
			switch n.ActiveNode.Type {
			case NODE_PRIMARY:
				n.ActiveNode = n.queueSecondary
			case NODE_SECONDARY:
				n.ActiveNode = n.queuePrimary
			}
			_, err := n.ActiveNode.Conn.Do("PING")
			if err != nil {
				n.mutex.Unlock()
				log.Fatal("no available node, please recheck queue node health check")
			}
			n.mutex.Unlock()
		}

		if qcount == 0 {
			time.Sleep(1 * time.Second)
			continue
		}
		byte, err := redis.Bytes(n.ActiveNode.Conn.Do("RPOP", updateQueueKey))
		if err != nil {
			log.Println(err)
		}
		ptr := &kvs.UpdateQueue{}
		if err := json.Unmarshal(byte, ptr); err != nil {
			log.Println(err)
		}
		_, err = m.Conn.Do("SET", ptr.Key, ptr.Data)
		if err != nil {
			log.Println(err)
		}

		log.Println("dequeue and store", ptr)
	}
}

func main() {
	os.Setenv("PRIMARY_REDIS_HOST", "localhost:63790")
	os.Setenv("SECONDARY_REDIS_HOST", "localhost:63791")
	os.Setenv("MASTER_REDIS_HOST", "localhost:6379")

	mconn, err := redis.Dial("tcp", os.Getenv("MASTER_REDIS_HOST"))
	if err != nil {
		log.Println(err)
		log.Fatal("failed to connect master storage.")
	}

	NewDistinctExecutor().Execute(&Node{Conn: mconn, Type: NODE_PRIMARY})
}
