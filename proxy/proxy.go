package proxy

//go:generate ../scripts/generate_certs.sh

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/elazarl/goproxy"
	"github.com/juju/errors"
	log "github.com/sirupsen/logrus"

	"github.com/9seconds/crawlera-headless-proxy/config"
)

var headersToRemove = map[string]struct{}{
	"accept-language":           struct{}{},
	"accept":                    struct{}{},
	"user-agent":                struct{}{},
	"upgrade-insecure-requests": struct{}{},
}

// NewProxy returns a new configured instance of goproxy.
func NewProxy(conf *config.Config) (*goproxy.ProxyHttpServer, error) {
	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = conf.Debug

	crawleraURL := conf.CrawleraURL()
	crawleraURLParsed, err := url.Parse(crawleraURL)
	if err != nil {
		return nil, errors.Annotate(err, "Incorrect Crawlera URL")
	}
	proxy.Tr = &http.Transport{
		Proxy: http.ProxyURL(crawleraURLParsed),
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: !conf.DoNotVerifyCrawleraCert, // nolint: gas
		},
	}
	proxy.ConnectDial = proxy.NewConnectDialToProxy(crawleraURL)

	proxy.OnRequest(goproxy.ReqHostMatches(regexp.MustCompile("^.*$"))).
		HandleConnect(goproxy.AlwaysMitm)
	proxy.OnRequest().DoFunc(
		func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			for k, v := range conf.XHeaders {
				req.Header.Set(k, v)
			}
			profile := req.Header.Get("X-Crawlera-Profile")
			if profile == "desktop" || profile == "mobile" {
				prepareForCrawleraProfile(req.Header, req.URL.String())
			}

			log.WithFields(log.Fields{
				"method":         req.Method,
				"url":            req.URL,
				"proto":          req.Proto,
				"content-length": req.ContentLength,
				"remote-addr":    req.RemoteAddr,
				"headers":        req.Header,
			}).Debug("HTTP request")

			return req, nil
		})
	proxy.OnResponse().DoFunc(func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		log.WithFields(log.Fields{
			"method":          resp.Request.Method,
			"url":             resp.Request.URL,
			"proto":           resp.Proto,
			"content-length":  resp.ContentLength,
			"headers":         resp.Header,
			"status":          resp.Status,
			"uncompressed":    resp.Uncompressed,
			"request-headers": resp.Request.Header,
		}).Debug("HTTP response")
		return resp
	})

	return proxy, nil
}

func init() {
	ca, err := tls.X509KeyPair(certCA, certKey)
	if err != nil {
		log.Fatal("Invalid certificates")
	}
	if ca.Leaf, err = x509.ParseCertificate(ca.Certificate[0]); err != nil {
		log.Fatal("Invalid certificates")
	}

	goproxy.GoproxyCa = ca
	tlsConfig := goproxy.TLSConfigFromCA(&ca)
	goproxy.OkConnect = &goproxy.ConnectAction{
		Action:    goproxy.ConnectAccept,
		TLSConfig: tlsConfig,
	}
	goproxy.MitmConnect = &goproxy.ConnectAction{
		Action:    goproxy.ConnectMitm,
		TLSConfig: tlsConfig,
	}
	goproxy.HTTPMitmConnect = &goproxy.ConnectAction{
		Action:    goproxy.ConnectHTTPMitm,
		TLSConfig: tlsConfig,
	}
	goproxy.RejectConnect = &goproxy.ConnectAction{
		Action:    goproxy.ConnectReject,
		TLSConfig: tlsConfig,
	}
}

func prepareForCrawleraProfile(headers http.Header, requestURL string) {
	for toRemove := range headersToRemove {
		headers.Del(toRemove)
	}
	for header := range headers {
		if strings.HasPrefix(header, "X-Crawlera-") { // nolint: megacheck
			continue
		}
	}

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
