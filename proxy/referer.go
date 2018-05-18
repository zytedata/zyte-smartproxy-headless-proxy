package proxy

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/elazarl/goproxy"
	"github.com/karlseguin/ccache"

	"github.com/9seconds/crawlera-headless-proxy/config"
)

const refererTTL = 10 * time.Second

type refererHandler struct {
	referers *ccache.LayeredCache
}

func (rh *refererHandler) installRequest(proxy *goproxy.ProxyHttpServer, conf *config.Config) handlerTypeReq {
	return func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		clientID := getState(ctx).clientID
		host := req.URL.Hostname()
		if host == "" {
			host = req.Host
		}

		referer := rh.cleanReferer(req.Referer())
		if referer == "" {
			referer = rh.get(clientID, host)
			if referer == "" {
				referer = rh.cleanReferer(req.URL.String())
			}
		}
		rh.set(clientID, host, referer)
		req.Header.Set("Referer", referer)

		return req, nil
	}
}

func (rh *refererHandler) set(primary, secondary, item string) {
	rh.referers.Set(primary, secondary, item, refererTTL)
}

func (rh *refererHandler) get(primary, secondary string) string {
	if item := rh.referers.Get(primary, secondary); item != nil {
		if !item.Expired() {
			return item.Value().(string)
		}
	}
	return ""
}

func (rh *refererHandler) cleanReferer(referer string) string {
	urlCopy, err := url.Parse(referer)
	if err != nil {
		return referer
	}

	urlCopy.Fragment = ""
	urlCopy.RawQuery = ""
	urlCopy.ForceQuery = false
	urlCopy.Host = rh.prepareHost(urlCopy)

	return urlCopy.String()
}

func (rh *refererHandler) prepareHost(data *url.URL) string {
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

func newRefererHandler() handlerReqInterface {
	return &refererHandler{
		referers: ccache.Layered(ccache.Configure()),
	}
}
