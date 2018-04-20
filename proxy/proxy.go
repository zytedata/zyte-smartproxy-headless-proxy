package proxy

//go:generate ../scripts/generate_certs.sh

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"net/url"
	"regexp"

	"github.com/elazarl/goproxy"
	"github.com/juju/errors"
	log "github.com/sirupsen/logrus"

	"github.com/9seconds/crawlera-headless-proxy/config"
)

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
			log.WithFields(log.Fields{
				"method":         req.Method,
				"url":            req.URL,
				"proto":          req.Proto,
				"content-length": req.ContentLength,
				"remote-addr":    req.RemoteAddr,
				"headers":        req.Header,
			}).Debug("Incoming HTTP request.")
			return req, nil
		})

	if conf.BestPractices {
		applyBestHeaders(proxy, conf)
	} else {
		applyCommonHeaders(proxy, conf)
	}

	proxy.OnRequest().DoFunc(
		func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			log.WithFields(log.Fields{
				"method":         req.Method,
				"url":            req.URL,
				"proto":          req.Proto,
				"content-length": req.ContentLength,
				"remote-addr":    req.RemoteAddr,
				"headers":        req.Header,
			}).Debug("HTTP request to send.")

			return req, nil
		})

	proxy.OnResponse().DoFunc(
		func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
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

// InitCertificates sets certificates for goproxy
func InitCertificates(certCA, certKey []byte) error {
	ca, err := tls.X509KeyPair(certCA, certKey)
	if err != nil {
		return errors.Annotate(err, "Invalid certificates")
	}
	if ca.Leaf, err = x509.ParseCertificate(ca.Certificate[0]); err != nil {
		return errors.Annotate(err, "Invalid certificates")
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

	return nil
}
