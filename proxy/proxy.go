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

	"bitbucket.org/scrapinghub/crawlera-headless-proxy/config"
	"bitbucket.org/scrapinghub/crawlera-headless-proxy/middleware"
	"bitbucket.org/scrapinghub/crawlera-headless-proxy/stats"
)

// NewProxy returns a new configured instance of goproxy.
func NewProxy(conf *config.Config, statsContainer *stats.Stats) (*goproxy.ProxyHttpServer, error) {
	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = false

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

	proxy.OnRequest().DoFunc(middleware.InitMiddlewares(statsContainer))
	middlewares := []middleware.Middleware{
		middleware.NewIncomingLogMiddleware(conf, proxy, statsContainer),
		middleware.NewStateMiddleware(conf, proxy, statsContainer),
	}
	if len(conf.AdblockLists) > 0 {
		middlewares = append(middlewares, middleware.NewAdblockMiddleware(conf, proxy, statsContainer))
	}
	middlewares = append(middlewares,
		middleware.NewRateLimiterMiddleware(conf, proxy, statsContainer),
		middleware.NewHeadersMiddleware(conf, proxy, statsContainer),
		middleware.NewRefererMiddleware(conf, proxy, statsContainer),
	)
	if !conf.NoAutoSessions {
		middlewares = append(middlewares, middleware.NewSessionsMiddleware(conf, proxy, statsContainer))
	}
	middlewares = append(middlewares, middleware.NewProxyRequestMiddleware(conf, proxy, statsContainer))

	for i := 0; i < len(middlewares); i++ {
		proxy.OnRequest().DoFunc(middlewares[i].OnRequest())
		proxy.OnResponse().DoFunc(middlewares[len(middlewares)-i-1].OnResponse())
	}

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
