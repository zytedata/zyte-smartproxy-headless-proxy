package proxy

//go:generate scripts/generate_certs.sh

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"net/url"

	"github.com/elazarl/goproxy"
	log "github.com/sirupsen/logrus"

	"github.com/9seconds/crawlera-headless-proxy/config"
)

func NewProxy(conf *config.Config) *goproxy.ProxyHttpServer {
	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = conf.Debug

	crawleraURL := conf.CrawleraURL()
	crawleraURLParsed, _ := url.Parse(crawleraURL) // nolint: gas
	proxy.Tr = &http.Transport{
		Proxy:           http.ProxyURL(crawleraURLParsed),
		TLSClientConfig: &tls.Config{InsecureSkipVerify: !conf.VerifyCrawleraCert}, // nolint: gas
	}
	proxy.ConnectDial = proxy.NewConnectDialToProxy(crawleraURL)

	proxy.OnRequest().DoFunc(
		func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			for k, v := range conf.XHeaders {
				req.Header.Set(k, v)
			}
			return req, nil
		})

	return proxy
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
	goproxy.OkConnect = &goproxy.ConnectAction{
		Action:    goproxy.ConnectAccept,
		TLSConfig: goproxy.TLSConfigFromCA(&ca),
	}
	goproxy.MitmConnect = &goproxy.ConnectAction{
		Action:    goproxy.ConnectMitm,
		TLSConfig: goproxy.TLSConfigFromCA(&ca),
	}
	goproxy.HTTPMitmConnect = &goproxy.ConnectAction{
		Action:    goproxy.ConnectHTTPMitm,
		TLSConfig: goproxy.TLSConfigFromCA(&ca),
	}
	goproxy.RejectConnect = &goproxy.ConnectAction{
		Action:    goproxy.ConnectReject,
		TLSConfig: goproxy.TLSConfigFromCA(&ca),
	}
}
