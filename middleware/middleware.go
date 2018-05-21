package middleware

import (
	"net/http"

	"github.com/elazarl/goproxy"
	log "github.com/sirupsen/logrus"

	"github.com/9seconds/crawlera-headless-proxy/config"
)

type ReqType func(*http.Request, *goproxy.ProxyCtx) (*http.Request, *http.Response)
type RespType func(*http.Response, *goproxy.ProxyCtx) *http.Response
type middlewareType uint8

type Middleware interface {
	OnRequest() ReqType
	OnResponse() RespType
}

type Base struct {
	conf  *config.Config
	proxy *goproxy.ProxyHttpServer
	mtype middlewareType
}

func (b *Base) BaseOnRequest(callback ReqType) ReqType {
	return func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		GetRequestState(ctx).seenMiddlewares[b.mtype] = struct{}{}
		return callback(req, ctx)
	}
}

func (b *Base) BaseOnResponse(callback RespType) RespType {
	return func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		if _, ok := GetRequestState(ctx).seenMiddlewares[b.mtype]; ok {
			return callback(resp, ctx)
		}
		return resp
	}
}

type UniqBase struct {
	Base
}

func (u *UniqBase) BaseOnRequest(callback ReqType) ReqType {
	baseFunc := u.BaseOnRequest(callback)
	return func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		if _, ok := GetRequestState(ctx).seenMiddlewares[u.mtype]; ok {
			log.Fatalf("%v middleware already set", u.mtype)
		}
		return baseFunc(req, ctx)
	}
}

func (u *UniqBase) BaseOnResponse(callback RespType) RespType {
	baseFunc := u.Base.BaseOnResponse(callback)
	return func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		delete(GetRequestState(ctx).seenMiddlewares, u.mtype)
		return baseFunc(resp, ctx)
	}
}
