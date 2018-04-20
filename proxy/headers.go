package proxy

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/elazarl/goproxy"

	"github.com/9seconds/crawlera-headless-proxy/config"
)

var headersProfileToRemove = [3]string{
	"accept",
	"user-agent",
	"upgrade-insecure-requests",
}

var headersBestToRemove = [5]string{
	"accept",
	"accept-language",
	"dnt",
	"upgrade-insecure-requests",
	"user-agent",
}

func applyBestHeaders(proxy *goproxy.ProxyHttpServer, conf *config.Config) {
	proxy.OnRequest().DoFunc(
		func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			for _, toRemove := range headersBestToRemove {
				req.Header.Del(toRemove)
			}

			setReferer(req.Header, req.URL.String())
			// this is required because golang's http client does not respect
			// header order.
			req.Header.Set("X-Crawlera-Profile", "desktop")
			req.Header.Set("X-Crawlera-Cookies", "disable")

			return req, nil
		})
}

func applyCommonHeaders(proxy *goproxy.ProxyHttpServer, conf *config.Config) {
	proxy.OnRequest().DoFunc(
		func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			for k, v := range conf.XHeaders {
				req.Header.Set(k, v)
			}

			profile := req.Header.Get("X-Crawlera-Profile")
			if profile == "desktop" || profile == "mobile" {
				for _, toRemove := range headersProfileToRemove {
					req.Header.Del(toRemove)
				}
				setReferer(req.Header, req.URL.String())
			}

			return req, nil
		})
}

func setReferer(headers http.Header, requestURL string) {
	if headers.Get("Referer") == "" {
		if urlCopy, err := url.Parse(requestURL); err == nil {
			urlCopy.Fragment = ""
			urlCopy.RawQuery = ""
			urlCopy.ForceQuery = false
			urlCopy.Host = prepareHost(urlCopy)
			headers.Set("Referer", urlCopy.String())
		}
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
