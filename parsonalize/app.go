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

func main() {

	mc := memcache.New("/tmp/tus.sock")

	m := func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/inc":
			incHandler(ctx, mc)
		default:
			ctx.Error("Unsupported path", fasthttp.StatusNotFound)
		}

	}

	fasthttp.ListenAndServe(":7777", m)
}

func incHandler(ctx *fasthttp.RequestCtx, mc *memcache.Client) {
	uid := handleUUID(ctx)

	i, err := mc.Get(uid)
	if err != nil {
		log.Println(err)
	}
	if i.Value == nil || bytes.Equal(i.Value, []byte("")) {
		i.Value = []byte("1")
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

	ctx.Response.SetBodyString(strconv.Itoa(ints))
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
