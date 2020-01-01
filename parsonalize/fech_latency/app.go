package main

import (
	"bytes"
	"fmt"

	redis "github.com/go-redis/redis/v7"

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

	rdb := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:       []string{"192.168.0.20:7000", "192.168.0.20:7001", "192.168.0.20:7002"},
		MaxRetries:  3,
		ReadTimeout: 5 * time.Second,
	})
	_, err := rdb.Ping().Result()
	if err != nil {

		log.Println("redis ping")

		log.Fatalln(err)
	}

	m := func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/latency_tus":
			latencyTus(ctx, mc)
		case "/latency_redis":
			latencyRedis(ctx, rdb)
		default:
			ctx.Error("Unsupported path", fasthttp.StatusNotFound)
		}

	}

	fasthttp.ListenAndServe(":"+p, m)
}

func latencyTus(ctx *fasthttp.RequestCtx, mc *memcache.Client) {
	uid := handleUUID(ctx)

	start := time.Now()
	i, err := mc.Get(uid)
	result := time.Now().Sub(start)
	if err != nil {
		log.Fatalln(err)
	}
	if i == nil || i.Value == nil || bytes.Equal(i.Value, []byte("")) {
		i.Value = []byte("0")
	}

	strint := fmt.Sprintf("%s", i.Value)
	ints, err := strconv.Atoi(strint)
	if err != nil {

		log.Fatalln(err)
	}

	ints++

	start = time.Now()
	err = mc.Set(&memcache.Item{Key: uid, Value: []byte(strconv.Itoa(ints)), Expiration: 0})
	result += time.Now().Sub(start)
	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("%d\n", result)

	ctx.Response.Header.Add("NODE", os.Getenv("NAME"))
	ctx.Response.SetBodyString(strconv.Itoa(ints))
}
func latencyRedis(ctx *fasthttp.RequestCtx, rd *redis.ClusterClient) {
	uid := handleUUID(ctx)

	start := time.Now()
	res, err := rd.Get(uid).Result()
	result := time.Now().Sub(start)
	if err != nil && err != redis.Nil {
		log.Println("redis fetch")
		log.Fatalln(err)
	}

	if err == redis.Nil {
		res = "0"
	}

	ints, err := strconv.Atoi(res)
	if err != nil {
		log.Println(err)
	}

	ints++
	start = time.Now()
	_, err = rd.Set(uid, ints, 0).Result()
	result += time.Now().Sub(start)
	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("%d\n", result)

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
