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

	"github.com/9seconds/crawlera-headless-proxy/config"
)

// NewProxy returns a new configured instance of goproxy.
func NewProxy(conf *config.Config) (*goproxy.ProxyHttpServer, error) {
	proxy := goproxy.NewProxyHttpServer()
	// proxy.Verbose = conf.Debug

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

	adblock := newAdblockHandler(conf.AdblockLists)
	headers := newHeaderHandler()
	logs := newLogHandler()
	state := newStateHandler()
	limiter := newRateLimiter(conf.ConcurrentConnections)
	sessions := newSessionHandler(proxy.Tr)

	for _, v := range getReqHandlers(proxy, conf, state, limiter, adblock, headers, sessions, logs) {
		proxy.OnRequest().DoFunc(v)
	}
	for _, v := range getRespHandlers(proxy, conf, limiter, sessions, logs) {
		proxy.OnResponse().DoFunc(v)
	}

	return proxy, nil
}

func getReqHandlers(proxy *goproxy.ProxyHttpServer, conf *config.Config,
	state, limiter, adblock, headers, sessions handlerInterface,
	logs logHandlerInterface) (handlers []handlerTypeReq) {
	handlers = append(handlers, state.installRequest(proxy, conf))

	if conf.ConcurrentConnections > 0 {
		handlers = append(handlers, limiter.installRequest(proxy, conf))
	}
	if len(conf.AdblockLists) > 0 {
		handlers = append(handlers, adblock.installRequest(proxy, conf))
	}
	if !conf.NoAutoSessions {
		handlers = append(handlers, sessions.installRequest(proxy, conf))
	}

	handlers = append(handlers,
		logs.installRequestInitial(proxy, conf),
		headers.installRequest(proxy, conf),
		logs.installRequest(proxy, conf))

	return
}

func getRespHandlers(proxy *goproxy.ProxyHttpServer, conf *config.Config,
	limiter, sessions handlerInterface,
	logs logHandlerInterface) (handlers []handlerTypeResp) {
	handlers = append(handlers, logs.installResponse(proxy, conf))

	if !conf.NoAutoSessions {
		handlers = append(handlers, sessions.installResponse(proxy, conf))
	}
	if conf.ConcurrentConnections > 0 {
		handlers = append(handlers, limiter.installResponse(proxy, conf))
	}

	return
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
