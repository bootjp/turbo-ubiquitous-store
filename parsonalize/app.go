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

		var response = make([]byte, len("VALUE 1"))
		_, err = conn.Read(response)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("response", response)
		if string(response) == "" {
			response = []byte("0")
		}

		val, err := strconv.Atoi(string(response[6:]))
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("num", val)
		val++
		fmt.Println(uid)
		fmt.Println("SET " + uid + " 1676598712 1676598712 1676598712\r\n" + string(val) + "\r\n")

		fmt.Println(cmd)

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
