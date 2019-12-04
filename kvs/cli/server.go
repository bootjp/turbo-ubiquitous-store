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
	"syscall"
	"time"

	"github.com/bootjp/turbo-ubiquitous-store/kvs"

	"github.com/patrickmn/go-cache"
)

var emptyres = ""

type TUSCache struct {
	*cache.Cache
}

func NewTUSCache() *TUSCache {
	return &TUSCache{cache.New(100*time.Second, 100*time.Second)}
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
		return emptyres, ErrorNotfound
	}
	if v, ok := value.(string); ok {
		return v, nil
	}

	return emptyres, ErrorBindMiss
}

var BreakLine = "\r\n"
var (
	FieldsCommand = 0
	FieldsKey     = 1
	FieldsFlag    = 2
	FieldsTTL     = 3
	FieldsSize    = 4
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
				val, err := cache.TUSGet(fields[FieldsKey])
				if err != nil {
					stdlog.Println(err)
				}
				_, err = c.Write([]byte(val + BreakLine))
				if err != nil {
					stdlog.Println(err)
				}
			case "SET":
				scanner.Scan()
				value := scanner.Text()
				ttl, err := strconv.Atoi(fields[FieldsTTL])
				if err != nil {
					stdlog.Println(err)
				}

				stdlog.Println("stored", value)
				cache.TUSSet(fields[1], value, time.Duration(ttl)*time.Second)
				q := kvs.UpdateQueue{
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
	queue.Drain()
	for queue.Length() != 0 {
		// waiting queue drain
	}
	os.Remove(sockPath)
	ln.Close()
	os.Exit(0)
}

var sockPath = "/tmp/tus.sock"

func main() {
	stdlog := log.New(os.Stdout, "", log.Ltime)

	ln, err := net.Listen("unix", sockPath)
	if err != nil {
		stdlog.Fatalln(err)
	}

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)
	cache := NewTUSCache()
	queue := kvs.NewQueueManager()
	go signalHaber(ln, sigc, queue)

	for {
		fd, err := ln.Accept()
		if err != nil {
			stdlog.Fatalln("Accept error: ", err)
		}
		go server(fd, cache, queue, stdlog)
	}
}
