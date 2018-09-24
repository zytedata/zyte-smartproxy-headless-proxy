package middleware

import (
	"net/http"

	"github.com/elazarl/goproxy"

	"github.com/scrapinghub/crawlera-headless-proxy/config"
	"github.com/scrapinghub/crawlera-headless-proxy/stats"
)

type rateLimiterMiddleware struct {
	UniqBase

	limiter            chan struct{}
	clientsServingChan chan<- bool
}

func (rl *rateLimiterMiddleware) OnRequest() ReqType {
	return rl.BaseOnRequest(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		if rl.limiter != nil {
			rl.limiter <- struct{}{}
		}
		rl.clientsServingChan <- true

		return req, nil
	})
}

func (rl *rateLimiterMiddleware) OnResponse() RespType {
	return rl.BaseOnResponse(func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		if rl.limiter != nil {
			<-rl.limiter
		}
		rl.clientsServingChan <- false

		return resp
	})
}

// NewRateLimiterMiddleware returns middleware which limits an amount of
// concurrent requests to Crawlera.
func NewRateLimiterMiddleware(conf *config.Config, proxy *goproxy.ProxyHttpServer,
	statsContainer *stats.Stats) Middleware {
	ware := &rateLimiterMiddleware{}
	ware.mtype = middlewareTypeRateLimiter

	ware.clientsServingChan = statsContainer.ClientsServingChan
	if conf.ConcurrentConnections > 0 {
		ware.limiter = make(chan struct{}, conf.ConcurrentConnections)
	}

	return ware
}
