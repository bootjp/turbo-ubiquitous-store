package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/patrickmn/go-cache"
)

var emptyres = []byte("")

type TUSCache struct {
	*cache.Cache
}

func NewTUSCache() *TUSCache {
	return &TUSCache{cache.New(100*time.Second, 100*time.Second)}
}

func (t *TUSCache) TUSSet(key string, data []byte, time time.Duration) bool {
	t.Set(key, data, time)
	return true
}

func (t *TUSCache) TUSGet(key string) ([]byte, error) {
	value, ok := t.Get(key)

	if !ok {
		return emptyres, nil
	}
	if v, ok := value.([]byte); ok {
		return v, nil
	}

	return emptyres, errors.New("not found")
}

// https://github.com/kayac/go-katsubushi/blob/master/app.go#L29-L39

var (
	respError         = []byte("ERROR\r\n")
	memdSep           = []byte("\r\n")
	memdSepLen        = len(memdSep)
	memdSpc           = []byte(" ")
	memdGets          = []byte("GETS")
	memdValue         = []byte("VALUE")
	memdEnd           = []byte("END")
	memdValHeader     = []byte("VALUE ")
	memdValFooter     = []byte("END\r\n")
	memdStatHeader    = []byte("STAT ")
	memdVersionHeader = []byte("VERSION ")
)

// MemdCmdQuit defines QUIT command.
type MemdCmdQuit int

// Execute disconnect by server.

func server(c net.Conn, cache *TUSCache) {

	for {
		bufReader := bufio.NewReader(c)
		scanner := bufio.NewScanner(bufReader)
		for scanner.Scan() {
			commandParser(scanner.Bytes())
			cache.TUSSet("aa", scanner.Bytes(), 100*time.Second)
			val, err := cache.TUSGet("aa")
			//log.Fatal(err)
			if err != nil {
				os.Exit(1)
			}
			fmt.Printf("%s", val)

		}

	}
}

//  https://github.com/kayac/go-katsubushi/blob/master/app.go#L323
func commandParser(command []byte) (cmd string, err error) {
	if len(command) == 0 {
		return
	}

	fields := strings.Fields(string(command))
	switch name := strings.ToUpper(fields[0]); name {
	case "GET":
		//atomic.AddInt64(&(app.cmdGet), 1)
		//if len(fields) < 2 {
		//	return "GET", fmt.Errorf("GET command needs key as second parameter")
		//}
		return "GET", nil
	case "SET":
		//if len(fields) < 5 && len(fields) > 3 {
		//	return "SET", fmt.Errorf("GET command needs key as second parameter")
		//}
		return "SET", nil
	default:
		return "ERROR", nil
	}
	return "ERROR", nil
}

func main() {

	ln, err := net.Listen("unix", "/tmp/tus.sock")
	if err != nil {
		log.Fatal("Listen error: ", err)
	}
	defer os.Remove("/tmp/tus.sock")

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)
	go func(ln net.Listener, c chan os.Signal) {
		sig := <-c

		log.Printf("Caught signal %s: shutting down.", sig)
		ln.Close()
		// todo enqueue here.
		os.Exit(0)
	}(ln, sigc)

	for {
		fd, err := ln.Accept()
		if err != nil {
			log.Fatal("Accept error: ", err)
		}

		c := NewTUSCache()

		go server(fd, c)
	}
}
