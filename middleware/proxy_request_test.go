package middleware

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type ProxyRequestSuite struct {
	MiddlewareTestSuite
}

func (t *ProxyRequestSuite) TestWrapper() {
	ware := NewProxyRequestMiddleware(t.cr.conf, nil, t.cr.s)

	reqHandler := ware.OnRequest()
	respHandler := ware.OnResponse()

	req, resp := reqHandler(t.cr.req, t.cr.ctx)
	t.NotNil(req)
	t.Nil(resp)
	t.Len(GetRequestState(t.cr.ctx).crawleraTimes, 1)
	t.Equal(GetRequestState(t.cr.ctx).CrawleraRequests, uint8(1))

	respHandler(nil, t.cr.ctx)
	t.Len(GetRequestState(t.cr.ctx).crawleraTimes, 2)
	t.Equal(GetRequestState(t.cr.ctx).CrawleraRequests, uint8(1))
}

func TestProxyRequest(t *testing.T) {
	suite.Run(t, &ProxyRequestSuite{})
}
