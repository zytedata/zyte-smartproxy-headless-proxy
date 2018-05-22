package middleware

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/elazarl/goproxy"
	"github.com/karlseguin/ccache"

	"github.com/9seconds/crawlera-headless-proxy/config"
	"github.com/9seconds/crawlera-headless-proxy/stats"
)

const refererTTL = 5 * time.Second

type refererMiddleware struct {
	UniqBase

	referers *ccache.LayeredCache
}

func (r *refererMiddleware) OnRequest() ReqType {
	return r.BaseOnRequest(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		clientID := GetRequestState(ctx).ClientID
		host := req.URL.Hostname()
		if host == "" {
			host = req.Host
		}

		referer := r.cleanReferer(req.Referer())
		if referer == "" {
			referer = r.get(clientID, host)
			if referer == "" {
				referer = r.cleanReferer(req.URL.String())
			}
		}
		r.set(clientID, host, referer)
		req.Header.Set("Referer", referer)

		return req, nil
	})
}

func (r *refererMiddleware) OnResponse() RespType {
	return r.BaseOnResponse(func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		return resp
	})
}

func (r *refererMiddleware) set(primary, secondary, item string) {
	r.referers.Set(primary, secondary, item, refererTTL)
}

func (r *refererMiddleware) get(primary, secondary string) string {
	if item := r.referers.Get(primary, secondary); item != nil && !item.Expired() {
		return item.Value().(string)
	}
	return ""
}

func (r *refererMiddleware) cleanReferer(referer string) string {
	urlCopy, err := url.Parse(referer)
	if err != nil {
		return referer
	}

	urlCopy.Fragment = ""
	urlCopy.RawQuery = ""
	urlCopy.ForceQuery = false
	urlCopy.Host = r.prepareHost(urlCopy)

	return urlCopy.String()
}

func (r *refererMiddleware) prepareHost(data *url.URL) string {
	splitted := strings.Split(data.Host, ":")
	if len(splitted) == 1 {
		return data.Host
	}

	if port, err := strconv.Atoi(splitted[len(splitted)-1]); err == nil {
		if (data.Scheme == "http" && port == 80) || (data.Scheme == "https" && port == 443) {
			return strings.Join(splitted[:len(splitted)-1], ":")
		}
	}

	return data.Host
}

// NewRefererMiddleware returns middleware which manages Referer header.
// If Referer is not set it tries to find out correct value for that.
func NewRefererMiddleware(conf *config.Config, proxy *goproxy.ProxyHttpServer, statsContainer *stats.Stats) Middleware {
	ware := &refererMiddleware{}
	ware.mtype = middlewareTypeReferer

	ware.referers = ccache.Layered(ccache.Configure())

	return ware
}
