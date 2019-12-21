package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/google/uuid"

	"github.com/valyala/fasthttp"
)

const DefaultTimeout = 10 * time.Second

func main() {
	p := os.Getenv("PORT")
	name := os.Getenv("NAME")
	sock := os.Getenv("SOCK_PATH")
	if p == "" || name == "" {
		log.Fatal("missing environment")
	}

	mc := memcache.New(sock)
	mc.Timeout = DefaultTimeout

	m := func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/inc":
			incHandler(ctx, mc)
		default:
			ctx.Error("Unsupported path", fasthttp.StatusNotFound)
		}

	}

	fasthttp.ListenAndServe(":"+p, m)
}

func incHandler(ctx *fasthttp.RequestCtx, mc *memcache.Client) {
	uid := handleUUID(ctx)

	i, err := mc.Get(uid)
	if err != nil {
		log.Println(err)
	}
	if i == nil || i.Value == nil || bytes.Equal(i.Value, []byte("")) {
		i.Value = []byte("0")
	}

	strint := fmt.Sprintf("%s", i.Value)
	ints, err := strconv.Atoi(strint)
	if err != nil {
		log.Println(err)
	}

	ints++
	mc.Set(&memcache.Item{Key: uid, Value: []byte(strconv.Itoa(ints)), Expiration: 0})

	if err != nil {
		log.Println(err)
	}

	ctx.Response.Header.Add("NODE", os.Getenv("NAME"))
	ctx.Response.SetBodyString(strconv.Itoa(ints))
}

func handleUUID(ctx *fasthttp.RequestCtx) string {
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
