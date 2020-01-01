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
		Addrs: []string{"172.17.0.1:7000", "172.17.0.1:7001", "172.17.0.1:7002"},
	})

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
	fmt.Println(result)
	if err != nil {
		log.Fatalln(err)
	}

	ctx.Response.Header.Add("NODE", os.Getenv("NAME"))
	ctx.Response.SetBodyString(strconv.Itoa(ints))
}
func latencyRedis(ctx *fasthttp.RequestCtx, rd *redis.ClusterClient) {
	uid := handleUUID(ctx)

	start := time.Now()
	res, err := rd.Get(uid).Result()
	result := time.Now().Sub(start)

	if res == "" {
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

	fmt.Println(result)
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
