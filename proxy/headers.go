package proxy

import (
	"net/http"

	"github.com/elazarl/goproxy"

	"github.com/9seconds/crawlera-headless-proxy/config"
)

var headersProfileToRemove = [6]string{
	"accept",
	"accept-encoding",
	"accept-language",
	"dnt",
	"upgrade-insecure-requests",
	"user-agent",
}

type headerHandler struct {
}

func (hh *headerHandler) installRequest(proxy *goproxy.ProxyHttpServer, conf *config.Config) handlerTypeReq {
	return func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		for k, v := range conf.XHeaders {
			req.Header.Set(k, v)
		}

		profile := req.Header.Get("X-Crawlera-Profile")
		if profile == "desktop" || profile == "mobile" {
			for _, v := range headersProfileToRemove {
				req.Header.Del(v)
			}
		}

		return req, nil
	}
}

func newHeaderHandler() handlerReqInterface {
	return &headerHandler{}
}
