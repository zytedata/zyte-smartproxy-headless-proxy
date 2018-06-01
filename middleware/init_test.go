package middleware

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestInitMiddlewares(t *testing.T) {
	cr := testNewProxyContainer()
	callback := InitMiddlewares(cr.s)
	req, resp := callback(cr.req, cr.ctx)

	assert.Exactly(t, req, cr.req)
	assert.Nil(t, resp)
	assert.NotNil(t, cr.ctx.UserData)

	state := cr.ctx.UserData.(*RequestState)
	assert.NotZero(t, state.ID)
	assert.NotZero(t, state.ClientID)
	assert.NotEqual(t, state.ID, state.ClientID)
	assert.Zero(t, state.CrawleraRequests)
	assert.Len(t, state.crawleraTimes, 0)
	assert.Len(t, state.seenMiddlewares, 0)
	assert.Zero(t, state.requestFinished)
	assert.WithinDuration(t, state.RequestStarted, time.Now(), time.Second)
}
