package main

import (
	"fmt"
	"log"

	"strconv"
	"sync"

	"github.com/google/uuid"
	gentleman "gopkg.in/h2non/gentleman.v2"
	"gopkg.in/h2non/gentleman.v2/plugins/cookies"
	"gopkg.in/h2non/gentleman.v2/plugins/url"
)

const parallelCount = 10

func main() {

	wg := &sync.WaitGroup{}
	for i := 1; i <= parallelCount; i++ {
		wg.Add(1)
		go consistentCheck(wg)
	}
	wg.Wait()

}

func consistentCheck(wg *sync.WaitGroup) {
	cli := gentleman.New()
	cli.Use(cookies.Set("uuid", uuid.New().String()))
	cli.Use(cookies.Jar())
	cli.Use(url.URL("http://localhost:8080/latency_tus"))

	for i := 1; i < 101; i++ {

		res, err := cli.Request().Send()
		if err != nil {
			log.Fatalf("Request error: %s\n", err)
			return
		}
		if !res.Ok {
			log.Fatalf("Invalid server response: %d\n", res.StatusCode)
			return
		}

		fmt.Printf("Body: %s\n", res.String())
		if res.String() != strconv.Itoa(i) {
			log.Fatalf("missing response %s : %s", res.String(), strconv.Itoa(i))
		}
	}

	cli = gentleman.New()
	cli.Use(cookies.Set("uuid", uuid.New().String()))
	cli.Use(cookies.Jar())
	cli.Use(url.URL("http://localhost:8080/latency_redis"))
	for i := 1; i < 101; i++ {

		res, err := cli.Request().Send()
		if err != nil {
			log.Fatalf("Request error: %s\n", err)
			return
		}
		if !res.Ok {
			log.Fatalf("Invalid server response: %d\n", res.StatusCode)
			return
		}

		fmt.Printf("Body: %s\n", res.String())
		if res.String() != strconv.Itoa(i) {
			log.Fatalf("missing response %s : %s", res.String(), strconv.Itoa(i))
		}
	}

	wg.Done()

}
