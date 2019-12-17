package main

import (
	"os"
	"time"

	"github.com/google/uuid"

	"github.com/valyala/fasthttp"
)

func main() {
	m := func(ctx *fasthttp.RequestCtx) {
		ctx.Response.SetBodyString(os.Getenv("PORT"))
		uid := ctx.Request.Header.Cookie("uuid")
		if uid != nil {
			c := &fasthttp.Cookie{}
			c.SetKey("uuid")
			c.SetValueBytes(uid)
			c.SetExpire(time.Now().Add(86400 * time.Second * 365))
			ctx.Response.Header.SetCookie(c)
			return
		}
		c := &fasthttp.Cookie{}
		c.SetKey("uuid")
		c.SetValue(uuid.New().String())
		c.SetExpire(time.Now().Add(86400 * time.Second * 365))
		ctx.Response.Header.SetCookie(c)

		return
	}

	fasthttp.ListenAndServe(":"+os.Getenv("PORT"), m)
}
