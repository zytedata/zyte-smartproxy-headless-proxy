package middleware

import (
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type RefererTestSuite struct {
	MiddlewareTestSuite

	handler ReqType
}

func (t *RefererTestSuite) SetupTest() {
	t.MiddlewareTestSuite.SetupTest()

	ware := NewRefererMiddleware(t.cr.conf, nil, t.cr.s).(*refererMiddleware)
	t.handler = ware.OnRequest()
}

func (t *RefererTestSuite) TestEmpty() {
	req, resp := t.handler(t.cr.req, t.cr.Ctx())
	t.Nil(resp)
	t.Equal(req.Header.Get("Referer"), "https://scrapinghub.com")
}

func (t *RefererTestSuite) TestDuplicate() {
	req, _ := t.handler(t.cr.req, t.cr.Ctx())
	req.URL = &url.URL{
		Scheme: "https",
		Host:   "scrapinghub.com",
		Path:   "path",
	}

	req, resp := t.handler(req, t.cr.Ctx())
	t.Nil(resp)
	t.Equal(req.Header.Get("Referer"), "https://scrapinghub.com")
}

func (t *RefererTestSuite) TestRespectOwnReferer() {
	req, _ := t.handler(t.cr.req, t.cr.Ctx())
	req.URL = &url.URL{
		Scheme: "https",
		Host:   "scrapinghub.com",
		Path:   "path",
	}
	req.Header.Set("Referer", "https://www.google.com")

	req, resp := t.handler(req, t.cr.Ctx())
	t.Nil(resp)
	t.Equal(req.Header.Get("Referer"), "https://www.google.com")

	req.Header.Del("Referer")
	req, resp = t.handler(req, t.cr.Ctx())
	t.Nil(resp)
	t.Equal(req.Header.Get("Referer"), "https://www.google.com")
}

func (t *RefererTestSuite) TestExpireReferer() {
	req, _ := t.handler(t.cr.req, t.cr.Ctx())
	req.Header.Set("Referer", "https://www.google.com")

	t.handler(req, t.cr.Ctx())
	time.Sleep(refererTTL + 500*time.Millisecond)
	req.Header.Del("Referer")

	req, resp := t.handler(req, t.cr.Ctx())
	t.Nil(resp)
	t.Equal(req.Header.Get("Referer"), "https://scrapinghub.com")
}

func TestReferer(t *testing.T) {
	suite.Run(t, &RefererTestSuite{})
}
