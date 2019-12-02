package main

import (
	"io"
	"log"
	"net"
	"os"
)

func reader(r io.Reader) {
	buf := make([]byte, 1024)
	for {
		var n, err = r.Read(buf[:])
		if err != nil {
			return
		}
		println("Client got:", string(buf[0:n]))
	}
}

func main() {
	c, err := net.Dial("unix", "/tmp/tus.sock")
	if err != nil {
		log.Fatal("Dial error", err)
	}
	defer c.Close()

	go reader(c)
	for {
		msg := "hi"
		_, err := c.Write([]byte(msg))
		if err != nil {
			log.Fatal("Write error:", err)
			break
		}
		println("Client sent:", msg)
		//time.Sleep(1e9)
		os.Exit(0)
	}
}
