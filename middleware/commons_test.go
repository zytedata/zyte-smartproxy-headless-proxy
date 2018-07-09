package middleware

import (
	"net/http"
	"net/http/httptest"

	"github.com/elazarl/goproxy"
	"github.com/stretchr/testify/suite"

	"bitbucket.org/scrapinghub/crawlera-headless-proxy/config"
	"bitbucket.org/scrapinghub/crawlera-headless-proxy/stats"
)

type MiddlewareTestSuite struct {
	suite.Suite

	cr *testProxyContainer
}

func (t *MiddlewareTestSuite) SetupTest() {
	t.cr = testInitNewProxyContainer()
}

type testProxyContainer struct {
	req  *http.Request
	resp *http.Response
	s    *stats.Stats
	ctx  *goproxy.ProxyCtx
	conf *config.Config
}

func (t *testProxyContainer) Ctx() *goproxy.ProxyCtx {
	callback := InitMiddlewares(t.s)
	ctx := &goproxy.ProxyCtx{Req: t.req, Resp: t.resp}
	callback(t.req, ctx)
	return ctx
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
