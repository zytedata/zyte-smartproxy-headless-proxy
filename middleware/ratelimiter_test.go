package middleware

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRateLimiterLimit(t *testing.T) {
	cr := testInitNewProxyContainer()
	cr.conf.ConcurrentConnections = 1
	ware := NewRateLimiterMiddleware(cr.conf, nil, cr.s).(*rateLimiterMiddleware)

	reqHandler := ware.OnRequest()
	respHandler := ware.OnResponse()
	var counter uint32

	testFunc := func() {
		ctx := cr.Ctx()
		_, resp := reqHandler(cr.req, ctx)
		assert.Nil(t, resp)
		atomic.AddUint32(&counter, 1)
		time.Sleep(100 * time.Millisecond)
		respHandler(resp, ctx)
	}

	go testFunc()
	go testFunc()

	time.Sleep(2 * time.Millisecond)
	assert.Equal(t, atomic.LoadUint32(&counter), uint32(1))

	time.Sleep(120 * time.Millisecond)
	assert.Equal(t, atomic.LoadUint32(&counter), uint32(2))

	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, atomic.LoadUint32(&counter), uint32(2))
}

func TestRateLimiterUnlimit(t *testing.T) {
	cr := testInitNewProxyContainer()
	cr.conf.ConcurrentConnections = 0
	ware := NewRateLimiterMiddleware(cr.conf, nil, cr.s).(*rateLimiterMiddleware)

	reqHandler := ware.OnRequest()
	respHandler := ware.OnResponse()
	var counter uint32

	testFunc := func() {
		ctx := cr.Ctx()
		_, resp := reqHandler(cr.req, ctx)
		assert.Nil(t, resp)
		atomic.AddUint32(&counter, 1)
		time.Sleep(100 * time.Millisecond)
		respHandler(resp, ctx)
	}

	go testFunc()
	go testFunc()

	time.Sleep(2 * time.Millisecond)
	assert.Equal(t, atomic.LoadUint32(&counter), uint32(2))
}
