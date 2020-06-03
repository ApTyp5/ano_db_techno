package logs

import (
	. "github.com/valyala/fasthttp"
)

func AccessLog(handler RequestHandler) RequestHandler {
	return func(ctx *RequestCtx) {
		Info("access\n", "request:", ctx.String())
		handler(ctx)
	}
}
