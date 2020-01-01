package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/chasex/redis-go-cluster"

	"github.com/google/uuid"

	"github.com/valyala/fasthttp"
)

const DefaultTimeout = 10 * time.Second

func main() {
	p := os.Getenv("PORT")
	name := os.Getenv("NAME")
	if p == "" || name == "" {
		//log.Fatal("missing environment")
	}

	cluster, err := redis.NewCluster(
		&redis.Options{
			StartNodes:   []string{"127.0.0.1:7000", "127.0.0.1:7001", "127.0.0.1:7002"},
			ConnTimeout:  50 * time.Millisecond,
			ReadTimeout:  50 * time.Millisecond,
			WriteTimeout: 50 * time.Millisecond,
			KeepAlive:    16,
			AliveTime:    60 * time.Second,
		})
	if err != nil {
		log.Println(err)
	}

	m := func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/inc":
			incHandler(ctx, cluster)
		default:
			ctx.Error("Unsupported path", fasthttp.StatusNotFound)
		}

	}

	fasthttp.ListenAndServe(":"+os.Getenv("PORT"), m)
}

func incHandler(ctx *fasthttp.RequestCtx, rd *redis.Cluster) {
	uid := handleUUID(ctx)

	ints, err := redis.Int(rd.Do("GET", uid))
	if err != nil {
		log.Println(err)
	}
	fmt.Println(uid, ints)

	ints++
	fmt.Println(uid, ints)

	_, err = rd.Do("SET", uid, ints)
	fmt.Println(uid, ints)
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
