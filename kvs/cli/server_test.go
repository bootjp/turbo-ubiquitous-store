package main

import (
	"net"
	"os"
	"sync"
	"testing"
	"time"
)

var clients = 2

func TestCommand(t *testing.T) {
	go main()
	time.Sleep(1 * time.Second)
	wg := &sync.WaitGroup{}
	wg.Add(clients)

	addr, err := net.ResolveUnixAddr("unix", sockPath)
	if err != nil {
		t.Errorf("%v", err)
	}

	for i := 0; i < clients; i++ {
		conn, err := net.DialUnix("unix", nil, addr)
		if err != nil {
			t.Error(err)
		}
		go func() {
			_, err := conn.Write([]byte("SET xxxx 50 50 50\r\n"))
			if err != nil {
				t.Fatal(err)
			}
			_, err = conn.Write([]byte("example\r\n"))
			if err != nil {
				t.Fatal(err)
			}
			time.Sleep(1 * time.Microsecond)
			_, err = conn.Write([]byte("GET xxxx\r\n"))
			if err != nil {
				t.Fatal(err)
			}
			var response = make([]byte, len("example"))
			_, err = conn.Read(response)
			if string(response) != "example" {
				t.Fatal(string(response))
			}
			wg.Done()
		}()
	}

	wg.Wait()
	os.Remove(sockPath)
}
