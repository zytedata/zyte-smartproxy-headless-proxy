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
	installHTTPClient(proxy.Tr)

	proxy.OnRequest(goproxy.ReqHostMatches(regexp.MustCompile("^.*$"))).
		HandleConnect(goproxy.AlwaysMitm)

	for _, v := range getReqHandlers(proxy, conf) {
		proxy.OnRequest().DoFunc(v)
	}
	for _, v := range getRespHandlers(proxy, conf) {
		proxy.OnResponse().DoFunc(v)
	}

	return proxy, nil
}

func getReqHandlers(proxy *goproxy.ProxyHttpServer, conf *config.Config) (handlers []handlerTypeReq) {
	handlers = append(handlers, handlerStateReq(proxy, conf))

	if conf.ConcurrentConnections > 0 {
		installRateLimiter(conf.ConcurrentConnections)
		handlers = append(handlers, handlerRateLimiterReq(proxy, conf))
	}
	if len(conf.AdblockLists) > 0 {
		handlers = append(handlers,
			newAdblockHandler(conf.AdblockLists).installRequest(proxy, conf))
	}
	if !conf.NoAutoSessions {
		handlers = append(handlers, handlerSessionReq(proxy, conf))
	}

	handlers = append(handlers,
		handlerLogReqInitial(proxy, conf),
		newHeaderHandler().installRequest(proxy, conf),
		handlerLogReqSent(proxy, conf))

	return
}

func getRespHandlers(proxy *goproxy.ProxyHttpServer, conf *config.Config) (handlers []handlerTypeResp) {
	handlers = append(handlers, handlerLogRespInitial(proxy, conf))

	if !conf.NoAutoSessions {
		handlers = append(handlers, handlerSessionResp(proxy, conf))
	}
	if conf.ConcurrentConnections > 0 {
		handlers = append(handlers, handlerRateLimiterResp(proxy, conf))
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
