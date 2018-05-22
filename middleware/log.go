package middleware

import (
	"net/http"

	"github.com/elazarl/goproxy"
	log "github.com/sirupsen/logrus"

	"github.com/9seconds/crawlera-headless-proxy/config"
	"github.com/9seconds/crawlera-headless-proxy/stats"
)

type incomingLogMiddleware struct {
	UniqBase
}

func (i *incomingLogMiddleware) OnRequest() ReqType {
	return i.BaseOnRequest(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		rstate := GetRequestState(ctx)

		log.WithFields(log.Fields{
			"request-id":     rstate.ID,
			"client-id":      rstate.ClientID,
			"method":         req.Method,
			"url":            req.URL,
			"proto":          req.Proto,
			"content-length": req.ContentLength,
			"remote-addr":    req.RemoteAddr,
			"headers":        req.Header,
		}).Debug("Incoming HTTP request")

		return req, nil
	})
}

func (i *incomingLogMiddleware) OnResponse() RespType {
	return i.BaseOnResponse(func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		if resp == nil {
			return nil
		}
		rstate := GetRequestState(ctx)

		log.WithFields(log.Fields{
			"request-id":        rstate.ID,
			"client-id":         rstate.ClientID,
			"method":            resp.Request.Method,
			"url":               resp.Request.URL,
			"proto":             resp.Proto,
			"content-length":    resp.ContentLength,
			"headers":           resp.Header,
			"status":            resp.Status,
			"uncompressed":      resp.Uncompressed,
			"request-headers":   resp.Request.Header,
			"elapsed-time":      rstate.Elapsed(),
			"crawlera-elapsed":  rstate.CrawleraElapsed(),
			"crawlera-requests": rstate.CrawleraRequests,
		}).Debug("HTTP response")

		return resp
	})
}

// NewIncomingLogMiddleware returns a middleware which logs response/requests.
func NewIncomingLogMiddleware(conf *config.Config, proxy *goproxy.ProxyHttpServer, statsContainer *stats.Stats) Middleware {
	ware := &incomingLogMiddleware{}
	ware.mtype = middlewareTypeIncomingLog

	return ware
}
