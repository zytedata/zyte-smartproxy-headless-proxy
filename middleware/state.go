package middleware

import (
	"net/http"
	"time"

	"github.com/elazarl/goproxy"

	"github.com/9seconds/crawlera-headless-proxy/config"
	"github.com/9seconds/crawlera-headless-proxy/stats"
)

type stateMiddleware struct {
	UniqBase

	requestsNumberChan   chan<- struct{}
	clientsConnectedChan chan<- bool
	overallTimesChan     chan<- time.Duration
	crawleraTimesChan    chan<- time.Duration
}

func (s *stateMiddleware) OnRequest() ReqType {
	return s.BaseOnRequest(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		s.requestsNumberChan <- struct{}{}
		s.clientsConnectedChan <- true

		return req, nil
	})
}

func (s *stateMiddleware) OnResponse() RespType {
	return s.BaseOnResponse(func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		rstate := GetRequestState(ctx)
		rstate.Finish()

		s.overallTimesChan <- rstate.Elapsed()
		s.clientsConnectedChan <- false
		for _, value := range rstate.CrawleraTimes() {
			s.crawleraTimesChan <- value
		}

		return resp
	})
}

// NewStateMiddleware returns middleware which does things that
// InitiMiddleware does not do.
func NewStateMiddleware(conf *config.Config, proxy *goproxy.ProxyHttpServer, statsContainer *stats.Stats) Middleware {
	ware := &stateMiddleware{}
	ware.mtype = middlewareTypeState

	ware.requestsNumberChan = statsContainer.RequestsNumberChan
	ware.clientsConnectedChan = statsContainer.ClientsConnectedChan
	ware.overallTimesChan = statsContainer.OverallTimesChan
	ware.crawleraTimesChan = statsContainer.CrawleraTimesChan

	return ware
}
