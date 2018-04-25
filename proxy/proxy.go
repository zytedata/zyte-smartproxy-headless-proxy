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

type handlerTypeReq func(*http.Request, *goproxy.ProxyCtx) (*http.Request, *http.Response)
type handlerTypeResp func(*http.Response, *goproxy.ProxyCtx) *http.Response

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

	handlersReq := []handlerTypeReq{
		handlerStateReq(proxy, conf),
	}
	handlersResp := []handlerTypeResp{}

	installHTTPClient(proxy.Tr)
	if conf.ConcurrentConnections > 0 {
		installRateLimiter(conf.ConcurrentConnections)
		handlersReq = append(handlersReq, handlerRateLimiterReq(proxy, conf))
	}

	if !conf.NoAutoSessions {
		handlersReq = append(handlersReq, handlerSessionReq(proxy, conf))
		handlersResp = append(handlersResp, handlerSessionResp(proxy, conf))
	}

	handlersReq = append(handlersReq,
		handlerLogReqInitial(proxy, conf),
		handlerHeadersReq(proxy, conf),
		handlerLogReqSent(proxy, conf))

	handlersResp = append(handlersResp, handlerLogRespInitial(proxy, conf))
	if conf.ConcurrentConnections > 0 {
		handlersResp = append(handlersResp, handlerRateLimiterResp(proxy, conf))
	}

	for _, v := range handlersReq {
		proxy.OnRequest().DoFunc(v)
	}
	for _, v := range handlersResp {
		proxy.OnResponse().DoFunc(v)
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
