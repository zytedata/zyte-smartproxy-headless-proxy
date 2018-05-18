package proxy

import (
	"net/http"
	"time"

	"github.com/elazarl/goproxy"
	log "github.com/sirupsen/logrus"

	"github.com/9seconds/crawlera-headless-proxy/config"
)

type logHandlerInterface interface {
	handlerReqRespInterface
	installRequestInitial(*goproxy.ProxyHttpServer, *config.Config) handlerTypeReq
	installResponseInitial(*goproxy.ProxyHttpServer, *config.Config) handlerTypeResp
}

type logHandler struct {
}

func (l *logHandler) installRequestInitial(proxy *goproxy.ProxyHttpServer, conf *config.Config) handlerTypeReq {
	return func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		state := getState(ctx)
		log.WithFields(log.Fields{
			"reqid":          state.id,
			"clientid":       state.clientID,
			"method":         req.Method,
			"url":            req.URL,
			"proto":          req.Proto,
			"content-length": req.ContentLength,
			"remote-addr":    req.RemoteAddr,
			"headers":        req.Header,
		}).Debug("Incoming HTTP request.")
		return req, nil
	}
}

func (l *logHandler) installRequest(proxy *goproxy.ProxyHttpServer, conf *config.Config) handlerTypeReq {
	return func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		state := getState(ctx)
		log.WithFields(log.Fields{
			"reqid":          state.id,
			"clientid":       state.clientID,
			"method":         req.Method,
			"url":            req.URL,
			"proto":          req.Proto,
			"content-length": req.ContentLength,
			"remote-addr":    req.RemoteAddr,
			"headers":        req.Header,
		}).Debug("HTTP request sent.")
		getState(ctx).crawleraStarted = time.Now()
		return req, nil
	}
}

func (l *logHandler) installResponseInitial(proxy *goproxy.ProxyHttpServer, conf *config.Config) handlerTypeResp {
	return func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		getState(ctx).crawleraFinished = time.Now()
		return resp
	}
}

func (l *logHandler) installResponse(proxy *goproxy.ProxyHttpServer, conf *config.Config) handlerTypeResp {
	return func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		if resp == nil {
			return resp
		}

		state := getState(ctx)
		log.WithFields(log.Fields{
			"reqid":           state.id,
			"clientid":        state.clientID,
			"method":          resp.Request.Method,
			"url":             resp.Request.URL,
			"proto":           resp.Proto,
			"content-length":  resp.ContentLength,
			"headers":         resp.Header,
			"status":          resp.Status,
			"uncompressed":    resp.Uncompressed,
			"request-headers": resp.Request.Header,
			"overall-time":    time.Since(state.requestStarted),
			"crawlera-time":   state.crawleraFinished.Sub(state.crawleraStarted),
		}).Debug("HTTP response")
		return resp
	}
}

func newLogHandler() logHandlerInterface {
	return &logHandler{}
}
