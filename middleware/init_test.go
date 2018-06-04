package middleware

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type InitMiddlewaresSuite struct {
	MiddlewareTestSuite
}

func (t *InitMiddlewaresSuite) TestWork() {
	callback := InitMiddlewares(t.cr.s)
	req, resp := callback(t.cr.req, t.cr.ctx)

	t.Exactly(req, t.cr.req)
	t.Nil(resp)
	t.NotNil(t.cr.ctx.UserData)

	state := GetRequestState(t.cr.ctx)
	t.NotZero(state.ID)
	t.NotZero(state.ClientID)
	t.NotEqual(state.ID, state.ClientID)
	t.Zero(state.CrawleraRequests)
	t.Len(state.crawleraTimes, 0)
	t.Len(state.seenMiddlewares, 0)
	t.Zero(state.requestFinished)
	t.WithinDuration(state.RequestStarted, time.Now(), time.Second)
}

func TestInitMiddlewares(t *testing.T) {
	suite.Run(t, &InitMiddlewaresSuite{})
}
