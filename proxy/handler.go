package proxy

import (
	"net/http"

	"github.com/elazarl/goproxy"

	"github.com/9seconds/crawlera-headless-proxy/config"
)

type handlerTypeReq func(*http.Request, *goproxy.ProxyCtx) (*http.Request, *http.Response)
type handlerTypeResp func(*http.Response, *goproxy.ProxyCtx) *http.Response

type handlerReqInterface interface {
	installRequest(*goproxy.ProxyHttpServer, *config.Config) handlerTypeReq
}

type handlerRespInterface interface {
	installResponse(*goproxy.ProxyHttpServer, *config.Config) handlerTypeResp
}

type handlerReqRespInterface interface {
	handlerReqInterface
	handlerRespInterface
}
