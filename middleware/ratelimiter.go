package middleware

import (
	"net/http"

	"github.com/elazarl/goproxy"

	"github.com/9seconds/crawlera-headless-proxy/config"
)

type rateLimiterMiddleware struct {
	UniqBase

	limiter chan struct{}
}

func (rl *rateLimiterMiddleware) OnRequest() ReqType {
	return rl.BaseOnRequest(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		rl.limiter <- struct{}{}
		return req, nil
	})
}

func (rl *rateLimiterMiddleware) OnResponse() RespType {
	return rl.BaseOnResponse(func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		<-rl.limiter
		return resp
	})
}

// NewRateLimiterMiddleware returns middleware which limits an amount of
// concurrent requests to Crawlera.
func NewRateLimiterMiddleware(conf *config.Config, proxy *goproxy.ProxyHttpServer) Middleware {
	ware := &rateLimiterMiddleware{}
	ware.mtype = middlewareTypeRateLimiter

	ware.limiter = make(chan struct{}, conf.ConcurrentConnections)

	return ware
}
