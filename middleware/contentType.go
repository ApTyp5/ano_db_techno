package middleware

import (
	. "github.com/valyala/fasthttp"
)

func ContentTypeAppJson(handler HandlerFunc) HandlerFunc {
	return func(ctx *Context) {
		ctx.SetContentType("application/json")
		handler(ctx)
	}
}
