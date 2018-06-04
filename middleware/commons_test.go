package middleware

import (
	"net/http"
	"net/http/httptest"

	"github.com/9seconds/crawlera-headless-proxy/config"
	"github.com/9seconds/crawlera-headless-proxy/stats"
	"github.com/elazarl/goproxy"
)

type testProxyContainer struct {
	req  *http.Request
	resp *http.Response
	s    *stats.Stats
	ctx  *goproxy.ProxyCtx
	conf *config.Config
}

func testNewProxyContainer() *testProxyContainer {
	req := httptest.NewRequest("GET", "https://scrapinghub.com", http.NoBody)
	resp := &http.Response{}

	return &testProxyContainer{
		req:  req,
		resp: resp,
		s:    stats.NewStats(),
		ctx:  &goproxy.ProxyCtx{Req: req, Resp: resp},
		conf: config.NewConfig(),
	}
}

func testInitNewProxyContainer() *testProxyContainer {
	cr := testNewProxyContainer()
	callback := InitMiddlewares(cr.s)
	req, _ := callback(cr.req, cr.ctx)
	cr.req = req

	return cr
}
