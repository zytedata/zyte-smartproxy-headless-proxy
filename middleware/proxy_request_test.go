package middleware

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProxyRequestWrapper(t *testing.T) {
	cr := testInitNewProxyContainer()
	ware := NewProxyRequestMiddleware(cr.conf, nil, cr.s)

	reqHandler := ware.OnRequest()
	respHandler := ware.OnResponse()

	req, resp := reqHandler(cr.req, cr.ctx)
	assert.NotNil(t, req)
	assert.Nil(t, resp)

	assert.Len(t, GetRequestState(cr.ctx).crawleraTimes, 1)
	assert.Equal(t, GetRequestState(cr.ctx).CrawleraRequests, uint8(1))

	respHandler(nil, cr.ctx)
	assert.Len(t, GetRequestState(cr.ctx).crawleraTimes, 2)
	assert.Equal(t, GetRequestState(cr.ctx).CrawleraRequests, uint8(1))
}
