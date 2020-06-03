package middleware

import (
	. "github.com/valyala/fasthttp"
)

func ContentTypeAppJson(handler RequestHandler) RequestHandler {
	return func(ctx *RequestCtx) {
		ctx.SetContentType("application/json")
		handler(ctx)
	}
}
