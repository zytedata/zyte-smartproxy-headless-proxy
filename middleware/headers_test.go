package middleware

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHeadersSetXheaders(t *testing.T) {
	cr := testInitNewProxyContainer()
	cr.conf.SetXHeader("cookies", "disable")
	cr.conf.SetXHeader("timeout", "180")
	ware := NewHeadersMiddleware(cr.conf, nil, cr.s).OnRequest()

	req, resp := ware(cr.req, cr.ctx)
	assert.Nil(t, resp)
	assert.Equal(t, req.Header.Get("X-Crawlera-Cookies"), "disable")
	assert.Equal(t, req.Header.Get("X-Crawlera-Timeout"), "180")
	assert.NotContains(t, req.Header, "X-Crawlera-Profile")
}

func TestHeadersSetProfile(t *testing.T) {
	cr := testInitNewProxyContainer()
	cr.conf.SetXHeader("profile", "desktop")

	req := cr.req
	req.Header.Set("Accept", "text/plain")
	req.Header.Set("Accept-Language", "ru_RU")

	ware := NewHeadersMiddleware(cr.conf, nil, cr.s).OnRequest()
	req, resp := ware(cr.req, cr.ctx)
	assert.Nil(t, resp)

	assert.Equal(t, req.Header.Get("X-Crawlera-Profile"), "desktop")
	assert.NotContains(t, req.Header, "Accept")
	assert.NotContains(t, req.Header, "Accept-Language")
}
