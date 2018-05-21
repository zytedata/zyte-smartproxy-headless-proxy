package middleware

import (
	"net/http"

	"github.com/elazarl/goproxy"

	"github.com/9seconds/crawlera-headless-proxy/config"
)

type stateMiddleware struct {
	UniqBase
}

func (s *stateMiddleware) OnRequest() ReqType {
	return s.BaseOnRequest(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		return req, nil
	})
}

func (s *stateMiddleware) OnResponse() RespType {
	return s.BaseOnResponse(func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		GetRequestState(ctx).Finish()
		return resp
	})
}

func NewStateMiddleware(conf *config.Config, proxy *goproxy.ProxyHttpServer) *stateMiddleware {
	ware := &stateMiddleware{}
	ware.conf = conf
	ware.proxy = proxy
	ware.mtype = middlewareTypeState

	return ware
}
