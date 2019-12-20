package main

import (
	"fmt"
	"log"
	"strconv"
	"sync"

	"github.com/google/uuid"

	"gopkg.in/h2non/gentleman.v2/plugins/cookies"
	"gopkg.in/h2non/gentleman.v2/plugins/url"

	gentleman "gopkg.in/h2non/gentleman.v2"
)

const clientCount = 100

func main() {

	wg := &sync.WaitGroup{}
	for i := 1; i <= clientCount; i++ {
		wg.Add(1)
		go consistentCheck(wg)
	}
	wg.Wait()

}

func consistentCheck(wg *sync.WaitGroup) {
	cli := gentleman.New()
	cli.Use(cookies.Set("uuid", uuid.New().String()))
	cli.Use(cookies.Jar())
	cli.Use(url.URL("http://localhost:8080/inc"))

	for i := 1; i < 101; i++ {

		res, err := cli.Request().Send()
		if err != nil {
			fmt.Printf("Request error: %s\n", err)
			return
		}
		if !res.Ok {
			fmt.Printf("Invalid server response: %d\n", res.StatusCode)
			return
		}

		fmt.Printf("Body: %s\n", res.String())
		if res.String() != strconv.Itoa(i) {
			log.Fatalf("missing response %s : %s", res.String(), strconv.Itoa(i))
		}
		log.Println("--")
	}

	wg.Done()
}
