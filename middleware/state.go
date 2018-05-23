package middleware

import (
	"io"
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
	trafficChan          chan<- uint64
}

// trafficCounter is a data structure which is wrapper for
// http.Response.Body. Its intent is to collect statistics on body size
// and send to collector.
type trafficCounter struct {
	body        io.ReadCloser
	counter     uint64
	trafficChan chan<- uint64
}

// Read supports io.ReadCloser interface.
func (tcb *trafficCounter) Read(p []byte) (n int, err error) {
	n, err = tcb.body.Read(p)
	tcb.counter += uint64(n)
	return
}

// Close supports io.ReadCloser interface.
func (tcb *trafficCounter) Close() error {
	tcb.trafficChan <- tcb.counter
	return tcb.body.Close()
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

		// This is to calculate statistics on traffic. Unfortunately, we
		// cannot trust Content-Length because of chunked encoding.
		if resp != nil {
			resp.Body = newTrafficCounter(resp, s.trafficChan)
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
	ware.trafficChan = statsContainer.TrafficChan

	return ware
}

func newTrafficCounter(resp *http.Response, trafficChan chan<- uint64) *trafficCounter {
	var trafficHeader int
	for k, vs := range resp.Header {
		trafficHeader += len(k)
		for _, v := range vs {
			trafficHeader += len(v)
		}
	}

	return &trafficCounter{
		body:        resp.Body,
		trafficChan: trafficChan,
		counter:     uint64(trafficHeader),
	}
}
