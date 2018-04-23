package proxy

import (
	"net/http"

	"github.com/elazarl/goproxy"
	"github.com/kamilsk/semaphore"
	log "github.com/sirupsen/logrus"

	"github.com/9seconds/crawlera-headless-proxy/config"
)

var rateLimiter semaphore.Semaphore

func installRateLimiter(number int) {
	rateLimiter = semaphore.New(number)
}

func handlerRateLimiterReq(proxy *goproxy.ProxyHttpServer, conf *config.Config) handlerTypeReq {
	return func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		if _, err := rateLimiter.Acquire(nil); err != nil {
			log.WithFields(log.Fields{
				"reqid": getState(ctx).id,
				"error": err,
			}).Warn("Error on acquiring semaphore.")
		}
		return req, nil
	}
}

func handlerRateLimiterResp(proxy *goproxy.ProxyHttpServer, conf *config.Config) handlerTypeResp {
	return func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		if err := rateLimiter.Release(); err != nil {
			log.WithFields(log.Fields{
				"reqid": getState(ctx).id,
				"error": err,
			}).Warn("Error on releasing semaphore.")
		}
		return resp
	}
}
