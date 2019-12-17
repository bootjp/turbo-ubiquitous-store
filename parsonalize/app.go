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

		cmd := fmt.Sprintf("GET %s\r￿\n", uid)
		_, err = conn.Write([]byte(cmd))
		if err != nil {
			fmt.Println(err)
		}

		var response = make([]byte, 100)
		_, err = conn.Read(response)
		if err != nil {
			fmt.Println(err)
		}

		val, err := strconv.Atoi(string(response))
		if err != nil {
			fmt.Println(err)
		}
		val++
		cmd = fmt.Sprintf("SET %s 50 10000 50\r\n%d\r\n", uid, val)
		_, err = conn.Write([]byte(cmd))
		if err != nil {
			fmt.Println(err)
		}

		ctx.Response.SetBodyString(string(val))
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
