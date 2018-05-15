package proxy

import (
	"net/http"

	"github.com/elazarl/goproxy"
	"github.com/kamilsk/semaphore"
	log "github.com/sirupsen/logrus"

	"github.com/9seconds/crawlera-headless-proxy/config"
)

type rateLimitHandler struct {
	limiter semaphore.Semaphore
}

func (rl *rateLimitHandler) installRequest(proxy *goproxy.ProxyHttpServer, conf *config.Config) handlerTypeReq {
	return func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		if _, err := rl.limiter.Acquire(nil); err != nil {
			log.WithFields(log.Fields{
				"reqid": getState(ctx).id,
				"error": err,
			}).Warn("Error on acquiring semaphore.")
		}
		return req, nil
	}
}

func (rl *rateLimitHandler) installResponse(proxy *goproxy.ProxyHttpServer, conf *config.Config) handlerTypeResp {
	return func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		if err := rl.limiter.Release(); err != nil {
			log.WithFields(log.Fields{
				"reqid": getState(ctx).id,
				"error": err,
			}).Warn("Error on releasing semaphore.")
		}
		return resp
	}
}

func newRateLimiter(number int) handlerReqRespInterface {
	return &rateLimitHandler{limiter: semaphore.New(number)}
}
