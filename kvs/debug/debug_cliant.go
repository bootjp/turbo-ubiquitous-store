package main

import (
	"bufio"
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
		var sc = bufio.NewScanner(os.Stdin)
		if sc.Scan() {
			t := sc.Text()
			_, err := c.Write([]byte(t + "\r\n"))
			if err != nil {
				log.Fatal("Write error:", err)
				break
			}
			println("Client sent:", t)
		}

	}
}
