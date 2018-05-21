package middleware

import (
	"net/http"

	"github.com/elazarl/goproxy"

	"github.com/9seconds/crawlera-headless-proxy/config"
)

type headersMiddleware struct {
	UniqBase

	xheaders map[string]string
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
		for k, v := range h.xheaders {
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

// NewHeadersMiddleware returns a middleware which mangles request
// headers. It injects x-headers for example and cleans up browser
// profiles related headers.
func NewHeadersMiddleware(conf *config.Config, proxy *goproxy.ProxyHttpServer) Middleware {
	ware := &headersMiddleware{}
	ware.xheaders = conf.XHeaders
	ware.mtype = middlewareTypeHeaders

	return ware
}
