package proxy

import (
	"net/http"

	"github.com/elazarl/goproxy"
	log "github.com/sirupsen/logrus"

	"github.com/9seconds/crawlera-headless-proxy/config"
)

type handlerTypeReq func(*http.Request, *goproxy.ProxyCtx) (*http.Request, *http.Response)
type handlerTypeResp func(*http.Response, *goproxy.ProxyCtx) *http.Response

type handlerInterface interface {
	installRequest(*goproxy.ProxyHttpServer, *config.Config) handlerTypeReq
	installResponse(*goproxy.ProxyHttpServer, *config.Config) handlerTypeResp
}

type handler struct {
}

func (h *handler) installRequest(proxy *goproxy.ProxyHttpServer, conf *config.Config) handlerTypeReq {
	return func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		log.Fatal("installRequest is not implemented")
		return req, nil
	}
}

func (h *handler) installResponse(proxy *goproxy.ProxyHttpServer, conf *config.Config) handlerTypeResp {
	return func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		log.Fatal("installResponse is not implemented")
		return resp
	}
}
