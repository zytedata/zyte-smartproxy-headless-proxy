package proxy

import (
	"net/http"

	"github.com/elazarl/goproxy"
	"github.com/kamilsk/semaphore"
	log "github.com/sirupsen/logrus"

	"github.com/9seconds/crawlera-headless-proxy/config"
)

func applyRateLimiter(proxy *goproxy.ProxyHttpServer, conf *config.Config) {
	limiter := semaphore.New(conf.ConcurrentConnections)

	proxy.OnRequest().DoFunc(
		func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			if _, err := limiter.Acquire(nil); err != nil {
				log.WithFields(log.Fields{
					"error": err,
				}).Warn("Error on acquiring semaphore.")
			}
			return req, nil
		})

	proxy.OnResponse().DoFunc(
		func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
			if err := limiter.Release(); err != nil {
				log.WithFields(log.Fields{
					"error": err,
				}).Warn("Error on releasing semaphore.")
			}
			return resp
		})

}
