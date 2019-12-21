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
	NodePrimary = iota
	NodeSecondary
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
	Load(node *MasterNode)
	Running() bool
	GetActiveNode() *Node
	Distinct(*[]kvs.UpdateQueue) *[]kvs.UpdateQueue
}

func (n *QueueNodes) DetachNode(node *Node) {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	switch node.Type {
	case NodePrimary:
		n.ActiveNode = n.queueSecondary
	case NodeSecondary:
		n.ActiveNode = n.queuePrimary
	}

	_, err := n.ActiveNode.Conn.Do("PING")
	if err != nil {
		log.Fatal("no available node, please recheck queue node health check")
	}
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
		ActiveNode:     &Node{Conn: pconn, Type: NodePrimary},
		queuePrimary:   &Node{Conn: pconn, Type: NodePrimary},
		queueSecondary: &Node{Conn: sconn, Type: NodeSecondary},
		isRunning:      false,
		log:            log.New(os.Stdout, "distinct_executor: ", log.Ltime),
	}

}

const updateQueueKey = "tus_queue"
const windowSize = 1 * time.Minute

func (n *QueueNodes) Distinct(buffer *[]kvs.UpdateQueue) map[string]kvs.UpdateQueue {

	m := map[string]kvs.UpdateQueue{}

	for _, queue := range *buffer {

		if _, ok := m[queue.Key]; !ok {
			m[queue.Key] = queue
		} else {
			// データが新しければ新しいものに書き換える
			if m[queue.Key].UpdateAt < queue.UpdateAt {
				m[queue.Key] = queue
			}
		}
	}

	return m
}

func (n *QueueNodes) Load(m *MasterNode) {
	if n.Running() {
		return
	}
	n.mutex.Lock()
	n.isRunning = true
	n.mutex.Unlock()

	var buffer []kvs.UpdateQueue

	for range time.Tick(windowSize) {
		n.log.Println("tick start")
		qcount, err := redis.Int(n.ActiveNode.Conn.Do("LLEN", updateQueueKey))
		if err != nil {
			n.log.Println(err)
			n.DetachNode(n.ActiveNode)
		}
		for qcount > 0 {
			memory := &kvs.UpdateQueue{}
			byte, err := redis.Bytes(n.ActiveNode.Conn.Do("RPOP", updateQueueKey))
			if err := json.Unmarshal(byte, memory); err != nil {
				n.log.Println(err)
			}
			buffer = append(buffer, *memory)
			qcount, err = redis.Int(n.ActiveNode.Conn.Do("LLEN", updateQueueKey))
			if err != nil {
				n.log.Println(err)
				n.DetachNode(n.ActiveNode)
			}
		}

		if len(buffer) == 0 {
			n.log.Println("distinct executor waiting queue")
			continue
		}
		data := n.Distinct(&buffer)
		for _, v := range data {
			_, err = m.Conn.Do("SET", v.Key, v.Data)
			if err != nil {
				n.log.Println(err, v)
			}
		}
	}
}

func main() {

	mconn, err := redis.Dial("tcp", os.Getenv("MASTER_REDIS_HOST"))
	if err != nil {
		log.Println(err)
		log.Fatal("failed to connect master storage.")
	}

	NewDistinctExecutor().Load(&Node{Conn: mconn, Type: NodePrimary})
}
