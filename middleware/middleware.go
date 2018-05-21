package middleware

import (
	"net/http"

	"github.com/elazarl/goproxy"
	log "github.com/sirupsen/logrus"
)

// ReqType is a base type for goproxy OnRequest() callback.
type ReqType func(*http.Request, *goproxy.ProxyCtx) (*http.Request, *http.Response)

// RespType is a base type for goproxy OnRequest() callback.
type RespType func(*http.Response, *goproxy.ProxyCtx) *http.Response

// middlewareType is an unique identifier of the middleware
type middlewareType uint8

// Middleware is an interface with 2 methods for goproxy callback
// initialization Please be noticed that response callbacks should be
// added in reverse order
type Middleware interface {
	OnRequest() ReqType
	OnResponse() RespType
}

// Base is a base type for any middleware. It provides 2 basic methods,
// BaseOnRequest which is doing some basic machinery for every
// middleware and BaseOnResponse - the same but for response. Usually,
// OnRequest and OnResponse implementation should invoke these 2 methods
// passing callback which contains the function specific for this
// certain middleware.
type Base struct {
	mtype middlewareType
}

// BaseOnRequest does basic middleware routines. You need to pass your
// callback in OnRequest method middleware to this method.
func (b *Base) BaseOnRequest(callback ReqType) ReqType {
	return func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		GetRequestState(ctx).seenMiddlewares[b.mtype] = struct{}{}
		return callback(req, ctx)
	}
}

// BaseOnResponse does basic middleware routines. You need to pass your
// callback in OnRequest method middleware to this method.
func (b *Base) BaseOnResponse(callback RespType) RespType {
	return func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		if _, ok := GetRequestState(ctx).seenMiddlewares[b.mtype]; ok {
			return callback(resp, ctx)
		}
		return resp
	}
}

// UniqBase is a version of the middleware which checks that it was
// executed only once in all directions (request and response). On error
// goroutine panics.
type UniqBase struct {
	Base
}

// BaseOnRequest does basic middleware routines. You need to pass your
// callback in OnRequest method middleware to this method.
func (u *UniqBase) BaseOnRequest(callback ReqType) ReqType {
	baseFunc := u.Base.BaseOnRequest(callback)
	return func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		if _, ok := GetRequestState(ctx).seenMiddlewares[u.mtype]; ok {
			log.Fatalf("%v middleware already set", u.mtype)
		}
		return baseFunc(req, ctx)
	}
}

// BaseOnResponse does basic middleware routines. You need to pass your
// callback in OnRequest method middleware to this method.
func (u *UniqBase) BaseOnResponse(callback RespType) RespType {
	baseFunc := u.Base.BaseOnResponse(callback)
	return func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		defer delete(GetRequestState(ctx).seenMiddlewares, u.mtype)
		return baseFunc(resp, ctx)
	}
}
