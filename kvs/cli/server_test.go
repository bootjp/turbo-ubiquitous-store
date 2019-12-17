package main

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"sync"
	"testing"
	"time"
)

var clients = 10

func TestCommand(t *testing.T) {
	os.Setenv("PRIMARY_REDIS_HOST", "localhost:63790")
	os.Setenv("SECONDARY_REDIS_HOST", "localhost:63791")
	os.Setenv("MASTER_REDIS_HOST", "localhost:6379")
	go main()

	time.Sleep(1 * time.Second)
	wg := &sync.WaitGroup{}

	addr, err := net.ResolveUnixAddr("unix", sockPath)
	if err != nil {
		t.Errorf("%v", err)
	}
	defer os.Remove(sockPath)

	for i := 0; i < clients; i++ {
		t.Log(i)
		wg.Add(1)
		go func(wg *sync.WaitGroup) {
			conn, err := net.DialUnix("unix", nil, addr)
			if err != nil {
				t.Error(err)
			}

			cmd := fmt.Sprintf("SET xxxx%d 50 50 50\r\nexample\r\n", i)
			_, err = conn.Write([]byte(cmd))
			t.Log("aaa")
			if err != nil {
				t.Fatal(err)
			}

			cmd = fmt.Sprintf("GET xxxx%d\r\n", i)
			_, err = conn.Write([]byte(cmd))
			t.Log("ccc")
			if err != nil {
				t.Fatal(err)
			}
			var response = make([]byte, len("example"))
			_, err = conn.Read(response)
			t.Log("ddd")
			if err != nil {
				t.Error(err)
			}
			if bytes.EqualFold(response, []byte("example")) {
				t.Fatal(string(response))
			}
			t.Log("response equal")
			conn.Close()
			wg.Done()
		}(wg)
	}

	wg.Wait()

}
