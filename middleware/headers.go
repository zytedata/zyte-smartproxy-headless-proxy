package middleware

import (
	"net/http"

	"github.com/elazarl/goproxy"

	"github.com/9seconds/crawlera-headless-proxy/config"
)

type headersMiddleware struct {
	UniqBase
}

var headersProfileToRemove = [6]string{
	"accept",
	"accept-encoding",
	"accept-language",
	"dnt",
	"upgrade-insecure-requests",
	"user-agent",
}

func (h *headersMiddleware) OnRequest() ReqType {
	return h.BaseOnRequest(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		for k, v := range h.conf.XHeaders {
			req.Header.Set(k, v)
		}

		profile := req.Header.Get("X-Crawlera-Profile")
		if profile == "desktop" || profile == "mobile" {
			for _, v := range headersProfileToRemove {
				req.Header.Del(v)
			}
		}

		return req, nil
	})
}

func (h *headersMiddleware) OnResponse() RespType {
	return h.BaseOnResponse(func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		return resp
	})
}

func NewHeadersMiddleware(conf *config.Config, proxy *goproxy.ProxyHttpServer) *headersMiddleware {
	ware := &headersMiddleware{}
	ware.conf = conf
	ware.proxy = proxy
	ware.mtype = middlewareTypeHeaders

	return ware
}
