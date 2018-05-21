package middleware

import (
	"net/http"

	"github.com/elazarl/goproxy"
	"github.com/kamilsk/semaphore"
	log "github.com/sirupsen/logrus"

	"github.com/9seconds/crawlera-headless-proxy/config"
)

type rateLimiterMiddleware struct {
	UniqBase

	limiter semaphore.Semaphore
}

func (rl *rateLimiterMiddleware) OnRequest() ReqType {
	return rl.BaseOnRequest(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		if _, err := rl.limiter.Acquire(nil); err != nil {
			rstate := GetRequestState(ctx)
			log.WithFields(log.Fields{
				"request-id": rstate.ID,
				"client-id":  rstate.ClientID,
				"error":      err,
			}).Warn("Error on acquiring semaphore.")
		}
		return req, nil
	})
}

func (rl *rateLimiterMiddleware) OnResponse() RespType {
	return rl.BaseOnResponse(func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		if err := rl.limiter.Release(); err != nil {
			rstate := GetRequestState(ctx)
			log.WithFields(log.Fields{
				"request-id": rstate.ID,
				"client-id":  rstate.ClientID,
				"error":      err,
			}).Warn("Error on releasing semaphore.")
		}
		return resp
	})
}

func NewRateLimiterMiddleware(conf *config.Config, proxy *goproxy.ProxyHttpServer) *rateLimiterMiddleware {
	ware := &rateLimiterMiddleware{}
	ware.conf = conf
	ware.proxy = proxy
	ware.mtype = middlewareTypeRateLimiter

	ware.limiter = semaphore.New(conf.ConcurrentConnections)

	return ware
}
