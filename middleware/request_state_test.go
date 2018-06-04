package middleware

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	gock "gopkg.in/h2non/gock.v1"
)

type RequestStateSuite struct {
	suite.Suite

	cr    *testProxyContainer
	state *RequestState
}

func (t *RequestStateSuite) SetupTest() {
	t.cr = testNewProxyContainer()
	state, _ := newRequestState(t.cr.req, t.cr.s)
	t.state = state
}

func (t *RequestStateSuite) TestFinish() {
	t.state.Finish()
	t.WithinDuration(t.state.requestFinished, time.Now(), time.Second)

	oldFinish := t.state.requestFinished
	t.state.Finish()
	t.Equal(oldFinish, t.state.requestFinished)
}

func (t *RequestStateSuite) TestElapsed() {
	t.InEpsilon(t.state.Elapsed(), time.Millisecond, float64(time.Millisecond))
	t.state.Finish()

	time.Sleep(500 * time.Millisecond)
	t.InEpsilon(t.state.Elapsed(), 500*time.Millisecond, float64(time.Millisecond))
}

func (t *RequestStateSuite) TestStartCrawleraRequest() {
	t.Nil(t.state.StartCrawleraRequest())
	t.Equal(t.state.CrawleraRequests, uint8(1))
	t.Len(t.state.crawleraTimes, 1)
	t.WithinDuration(t.state.crawleraTimes[0], time.Now(), time.Millisecond)
	t.Len(t.state.crawleraRequestsChan, 1)

	t.NotNil(t.state.StartCrawleraRequest())
	t.Len(t.state.crawleraRequestsChan, 1)
	t.Len(t.state.crawleraTimesChan, 0)
	t.Equal(t.state.CrawleraRequests, uint8(1))
	t.Len(t.state.crawleraTimes, 1)
}

func (t *RequestStateSuite) TestFinishCrawleraRequest() {
	t.NotNil(t.state.FinishCrawleraRequest())
	t.Len(t.state.crawleraRequestsChan, 0)
	t.Len(t.state.crawleraTimesChan, 0)
	t.Equal(t.state.CrawleraRequests, uint8(0))

	t.state.StartCrawleraRequest()
	t.Nil(t.state.FinishCrawleraRequest())
	t.Len(t.state.crawleraRequestsChan, 1)
	t.Len(t.state.crawleraTimesChan, 1)
	t.Equal(t.state.CrawleraRequests, uint8(1))
	t.Len(t.state.crawleraTimes, 2)

	t.WithinDuration(t.state.crawleraTimes[0], time.Now(), time.Millisecond)
	t.NotNil(t.state.FinishCrawleraRequest())
	t.Len(t.state.crawleraRequestsChan, 1)
	t.Len(t.state.crawleraTimesChan, 1)
	t.Equal(t.state.CrawleraRequests, uint8(1))
	t.Len(t.state.crawleraTimes, 2)
}

func (t *RequestStateSuite) TestCrawleraElapsed() {
	t.Zero(t.state.CrawleraElapsed())
	t.state.StartCrawleraRequest()
	t.NotZero(t.state.CrawleraElapsed())
	time.Sleep(500 * time.Millisecond)

	t.state.FinishCrawleraRequest()
	t.InEpsilon(t.state.CrawleraElapsed(), 500*time.Millisecond, float64(time.Millisecond))
}

func (t *RequestStateSuite) TestDoCrawleraRequest() {
	newClient := &http.Client{}
	defer gock.Off()
	defer gock.RestoreClient(newClient)

	gock.New("https://scrapinghub.com").
		Get("/").
		Reply(200).
		BodyString("")

	t.state.DoCrawleraRequest(newClient, t.cr.req)
	t.Len(t.state.crawleraRequestsChan, 1)
	t.Len(t.state.crawleraTimesChan, 1)
	t.Equal(t.state.CrawleraRequests, uint8(1))
	t.Len(t.state.crawleraTimes, 2)
}

func TestRequestState(t *testing.T) {
	suite.Run(t, &RequestStateSuite{})
}
