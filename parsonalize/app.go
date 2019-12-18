package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/valyala/fasthttp"
)

func main() {

	addr, err := net.ResolveUnixAddr("unix", "/tmp/tus.sock")
	if err != nil {
		log.Fatal(err)
	}

	// todo fix blocking and refuse.
	m := func(ctx *fasthttp.RequestCtx) {
		uid := handleUUID(ctx)

		conn, err := net.DialUnix("unix", nil, addr)
		defer conn.Close()
		if err != nil {
			fmt.Println(err)
		}

		cmd := fmt.Sprintf("GET %s", uid)
		fmt.Println(cmd)
		_, err = conn.Write([]byte(cmd + "\r\n"))
		if err != nil {
			fmt.Println(err)
		}
		var meta = make([]byte, len(fmt.Sprintf("VALUE %s 0 x", uid)))
		_, err = conn.Read(meta)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Printf("meta: %s\n", meta)

		var response = make([]byte, 1)
		_, err = conn.Read(response)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Printf("response: %s\n", response)

		if string(response) == "\r" {
			response = []byte("0")
		}

		val, err := strconv.Atoi(string(response))
		if err != nil {
			fmt.Println(err)
		}
		val++
		fmt.Println("num", val)

		fmt.Println(uid)

		_, err = conn.Write([]byte("SET " + uid + " 1676598712 1676598712 1676598712\r\n" + strconv.Itoa(val) + "\r\n"))
		if err != nil {
			fmt.Println(err)
		}

		if err != nil {
			fmt.Println(err)
		}

		ctx.Response.SetBody(response)
	}

	fasthttp.ListenAndServe(":7779", m)

}

func handleUUID(ctx *fasthttp.RequestCtx) string {
	ctx.Response.SetBodyString(os.Getenv("PORT"))
	uid := ctx.Request.Header.Cookie("uuid")
	if uid != nil {
		c := &fasthttp.Cookie{}
		c.SetKey("uuid")
		c.SetValueBytes(uid)
		c.SetExpire(time.Now().Add(86400 * time.Second * 365))
		ctx.Response.Header.SetCookie(c)
		return string(uid)
	}
	c := &fasthttp.Cookie{}
	c.SetKey("uuid")
	struid := uuid.New().String()
	c.SetValue(struid)
	c.SetExpire(time.Now().Add(86400 * time.Second * 365))
	ctx.Response.Header.SetCookie(c)

	return struid
}
