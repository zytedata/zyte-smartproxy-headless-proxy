package middleware

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	gock "gopkg.in/h2non/gock.v1"
)

func TestRequestStateFinish(t *testing.T) {
	cr := testNewProxyContainer()
	state, err := newRequestState(cr.req, cr.s)
	assert.Nil(t, err)

	state.Finish()
	assert.WithinDuration(t, state.requestFinished, time.Now(), time.Second)

	oldFinish := state.requestFinished
	state.Finish()
	assert.Equal(t, oldFinish, state.requestFinished)
}

func TestRequestStateElapsed(t *testing.T) {
	cr := testNewProxyContainer()
	state, err := newRequestState(cr.req, cr.s)
	assert.Nil(t, err)

	assert.InEpsilon(t, state.Elapsed(), time.Millisecond, float64(time.Millisecond))
	state.Finish()
	time.Sleep(500 * time.Millisecond)
	assert.InEpsilon(t, state.Elapsed(), 500*time.Millisecond, float64(time.Millisecond))
}

func TestRequestStateStartCrawleraRequest(t *testing.T) {
	cr := testNewProxyContainer()
	state, err := newRequestState(cr.req, cr.s)
	assert.Nil(t, err)

	assert.Nil(t, state.StartCrawleraRequest())
	assert.Equal(t, state.CrawleraRequests, uint8(1))
	assert.Len(t, state.crawleraTimes, 1)
	assert.WithinDuration(t, state.crawleraTimes[0], time.Now(), time.Millisecond)
	assert.Len(t, state.crawleraRequestsChan, 1)

	assert.NotNil(t, state.StartCrawleraRequest())
	assert.Len(t, state.crawleraRequestsChan, 1)
	assert.Len(t, state.crawleraTimesChan, 0)
	assert.Equal(t, state.CrawleraRequests, uint8(1))
	assert.Len(t, state.crawleraTimes, 1)
}

func TestRequestStateFinishCrawleraRequest(t *testing.T) {
	cr := testNewProxyContainer()
	state, err := newRequestState(cr.req, cr.s)
	assert.Nil(t, err)

	assert.NotNil(t, state.FinishCrawleraRequest())
	assert.Len(t, state.crawleraRequestsChan, 0)
	assert.Len(t, state.crawleraTimesChan, 0)
	assert.Equal(t, state.CrawleraRequests, uint8(0))

	state.StartCrawleraRequest()
	assert.Nil(t, state.FinishCrawleraRequest())
	assert.Len(t, state.crawleraRequestsChan, 1)
	assert.Len(t, state.crawleraTimesChan, 1)
	assert.Equal(t, state.CrawleraRequests, uint8(1))
	assert.Len(t, state.crawleraTimes, 2)

	assert.WithinDuration(t, state.crawleraTimes[0], time.Now(), time.Millisecond)
	assert.NotNil(t, state.FinishCrawleraRequest())
	assert.Len(t, state.crawleraRequestsChan, 1)
	assert.Len(t, state.crawleraTimesChan, 1)
	assert.Equal(t, state.CrawleraRequests, uint8(1))
	assert.Len(t, state.crawleraTimes, 2)
}

func TestRequestStateCrawleraElapsed(t *testing.T) {
	cr := testNewProxyContainer()
	state, err := newRequestState(cr.req, cr.s)
	assert.Nil(t, err)

	assert.Zero(t, state.CrawleraElapsed())
	state.StartCrawleraRequest()
	assert.NotZero(t, state.CrawleraElapsed())
	time.Sleep(500 * time.Millisecond)

	state.FinishCrawleraRequest()
	assert.InEpsilon(t, state.CrawleraElapsed(), 500*time.Millisecond, float64(time.Millisecond))
}

func TestRequestStateDoCrawleraRequest(t *testing.T) {
	cr := testNewProxyContainer()
	state, err := newRequestState(cr.req, cr.s)
	assert.Nil(t, err)

	newClient := &http.Client{}
	defer gock.Off()
	defer gock.RestoreClient(newClient)

	gock.New("https://scrapinghub.com").
		Get("/").
		Reply(200).
		BodyString("")

	state.DoCrawleraRequest(newClient, cr.req)
	assert.Len(t, state.crawleraRequestsChan, 1)
	assert.Len(t, state.crawleraTimesChan, 1)
	assert.Equal(t, state.CrawleraRequests, uint8(1))
	assert.Len(t, state.crawleraTimes, 2)
}
