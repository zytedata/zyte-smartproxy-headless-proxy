package proxy

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"

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
	handler
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
			if req.Header.Get("Referer") == "" {
				setReferer(req.Header, req.URL.String())
			}
		}

		return req, nil
	}
}

func setReferer(headers http.Header, requestURL string) {
	if urlCopy, err := url.Parse(requestURL); err == nil {
		urlCopy.Fragment = ""
		urlCopy.RawQuery = ""
		urlCopy.ForceQuery = false
		urlCopy.Host = prepareHost(urlCopy)
		headers.Set("Referer", urlCopy.String())
	}
}

func prepareHost(data *url.URL) string {
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

func newHeaderHandler() handlerInterface {
	return &headerHandler{}
}
