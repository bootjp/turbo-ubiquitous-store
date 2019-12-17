package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/akyoto/cache"
	"github.com/gomodule/redigo/redis"

	"github.com/bootjp/turbo-ubiquitous-store/kvs"
)

type TUSCache struct {
	*cache.Cache
	log         *log.Logger
	refreshInit bool
	refreshLock *sync.Mutex
}

func NewTUSCache() *TUSCache {
	t := &TUSCache{
		cache.New(1 * time.Hour),
		log.New(os.Stdout, "tuscache: ", log.Ltime),
		false,
		&sync.Mutex{},
	}
	t.LoadCache()
	return t
}

// load cache
func (t *TUSCache) LoadCache() {
	sconn, err := redis.Dial("tcp", os.Getenv("SLAVE_REDIS_HOST"))
	if err != nil {
		t.log.Println(sconn)
	}

	str, err := redis.Strings(sconn.Do("KEYS", "*"))
	if err != nil {
		t.log.Println(err)
	}
	for _, k := range str {
		v, err := redis.String(sconn.Do("GET", k))
		if err != nil {
			t.log.Println(err)
		}
		t.TUSSet(k, v, 1*time.Hour)
	}

}

// refrech cache with backend storage
func (t *TUSCache) RefreshCache() {
	t.refreshLock.Lock()
	defer t.refreshLock.Unlock()
	if t.refreshInit {
		return
	}

	t.refreshInit = true
	go func() {
		for {
			sconn, err := redis.Dial("tcp", os.Getenv("SLAVE_REDIS_HOST"))
			if err != nil {
				t.log.Println(sconn)
			}

			str, err := redis.Strings(sconn.Do("KEYS", "*"))
			if err != nil {
				t.log.Println(err)
			}
			for _, k := range str {
				v, err := redis.String(sconn.Do("GET", k))
				if err != nil {
					t.log.Println(err)
				}
				t.TUSSet(k, v, 1*time.Hour)
			}

			time.Sleep(10 * time.Minute)
		}
	}()
}
func (t *TUSCache) TUSSet(key string, data string, time time.Duration) bool {
	t.Set(key, data, time)
	return true
}

var ErrorNotfound = errors.New("not found")
var ErrorBindMiss = errors.New("bind miss")

func (t *TUSCache) TUSGet(key string) (string, error) {
	value, ok := t.Get(key)

	if !ok {
		return "", ErrorNotfound
	}

	return fmt.Sprintf("%v", value), nil
}

const BreakLine = "\r\n"
const Stored = "STORED" + BreakLine
const End = "END" + BreakLine
const ValueFormat = "VALUE %s 0 %d" + BreakLine + "%s" + BreakLine + End

var storedRes = []byte(Stored)

const (
	FieldsCommand = iota
	FieldsKey
	FieldsFlag
	FieldsTTL
	FieldsSize
)

func server(c net.Conn, cache *TUSCache, queue *kvs.QueueManager, stdlog *log.Logger) {
	for {
		bufReader := bufio.NewReader(c)
		scanner := bufio.NewScanner(bufReader)
		for scanner.Scan() {
			fields := strings.Fields(scanner.Text())
			if len(fields) == 0 {
				continue
			}

			switch name := strings.ToUpper(fields[FieldsCommand]); name {
			case "GET":
				if len(fields) != 2 {
					stdlog.Println("invalid command")
					continue
				}
				key := fields[FieldsKey]
				val, err := cache.TUSGet(key)
				if err != nil && err != ErrorNotfound {
					stdlog.Println(err)
				}
				_, err = c.Write([]byte(fmt.Sprintf(ValueFormat, key, len(val), val)))
				if err != nil {
					stdlog.Println(err)
				}
			case "SET":
				if len(fields) != 5 {
					stdlog.Println("invalid command", fields, len(fields))
					continue
				}
				scanner.Scan()
				value := scanner.Text()
				ttl, err := strconv.Atoi(fields[FieldsTTL])
				if err != nil {
					stdlog.Println(err)
				}

				cache.TUSSet(fields[1], value, time.Duration(ttl)*time.Second)
				_, err = c.Write(storedRes)
				if err != nil {
					stdlog.Println(err)
				}
				q := kvs.UpdateQueue{
					Key:      fields[1],
					Data:     value,
					UpdateAt: time.Now().Unix(),
				}
				queue.Enqueue(q)
			default:
				stdlog.Println(fmt.Errorf("UnSupport command %s", name))
				continue
			}
		}
	}
}

func signalHaber(ln net.Listener, c chan os.Signal, queue *kvs.QueueManager) {
	sig := <-c
	log.Printf("Caught signal %s: shutting down.", sig)
	ln.Close()
	//todo backlog: next node transfer data
	for trid := 0; queue.Length() > 0; trid++ {
		log.Println("waiting dequeue")
		time.Sleep(10 * time.Second)
		if trid > 20 {
			log.Fatal("fail deque sequence tried max")
		}
	}
	os.Remove(sockPath)
	ln.Close()
	os.Exit(0)
}

const sockPath = "/tmp/tus.sock"

func main() {
	stdlog := log.New(os.Stdout, "front_kvs: ", log.Ltime)

	ln, err := net.Listen("unix", sockPath)
	if err != nil {
		stdlog.Fatalln(err)
	}
	defer os.Remove(sockPath)

	if err := os.Chmod(sockPath, 0700); err != nil {
		log.Printf("error: %v\n", err)
		return
	}
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)
	cache := NewTUSCache()
	queue := kvs.NewQueueManager()
	go signalHaber(ln, sigc, queue)
	go queue.Forward()

	for {
		fd, err := ln.Accept()
		if err != nil {
			stdlog.Println("Accept error: ", err)
			continue
		}
		go server(fd, cache, queue, stdlog)
	}
}
