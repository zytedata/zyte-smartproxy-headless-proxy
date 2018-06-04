package middleware

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type HeadersTestSuite struct {
	MiddlewareTestSuite
}

func (t *HeadersTestSuite) W() ReqType {
	return NewHeadersMiddleware(t.cr.conf, nil, t.cr.s).OnRequest()
}

func (t *HeadersTestSuite) TestSetXheaders() {
	t.cr.conf.SetXHeader("cookies", "disable")
	t.cr.conf.SetXHeader("timeout", "180")

	req, resp := t.W()(t.cr.req, t.cr.Ctx())
	t.Nil(resp)
	t.Equal(req.Header.Get("X-Crawlera-Cookies"), "disable")
	t.Equal(req.Header.Get("X-Crawlera-Timeout"), "180")
	t.NotContains(req.Header, "X-Crawlera-Profile")
}

func (t *HeadersTestSuite) TestSetProfile() {
	t.cr.conf.SetXHeader("profile", "desktop")

	req := t.cr.req
	req.Header.Set("Accept", "text/plain")
	req.Header.Set("Accept-Language", "ru_RU")

	req, resp := t.W()(t.cr.req, t.cr.Ctx())
	t.Nil(resp)

	t.Equal(req.Header.Get("X-Crawlera-Profile"), "desktop")
	t.NotContains(req.Header, "Accept")
	t.NotContains(req.Header, "Accept-Language")
}

func TestHeaders(t *testing.T) {
	suite.Run(t, &HeadersTestSuite{})
}
