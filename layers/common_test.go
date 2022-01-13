package layers

import (
	"context"
	"net"
	"time"

	"github.com/9seconds/httransform/v2/events"
	"github.com/9seconds/httransform/v2/layers"
	"github.com/scrapinghub/crawlera-headless-proxy/stats"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/valyala/fasthttp"
)

type CommonLayerTestSuite struct {
	suite.Suite
	ctx           *layers.Context
	eventsChannel *EventChannelMock
}

type EventChannelMock struct {
	mock.Mock
}

func (e *EventChannelMock) Send(ctx context.Context, eventType events.EventType, value interface{}, shardKey string) {
	e.Called(ctx, eventType, value, shardKey)
}

func (suite *CommonLayerTestSuite) SetupTest() {
	fhttpCtx := &fasthttp.RequestCtx{}

	remoteAddr := &net.TCPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 65342,
	}

	fhttpCtx.Init(&fasthttp.Request{}, remoteAddr, nil)

	suite.ctx = layers.AcquireContext()
	suite.eventsChannel = &EventChannelMock{}

	// nolint:errcheck
	suite.ctx.Init(
		fhttpCtx,
		"127.0.0.1:8000",
		suite.eventsChannel,
		"user",
		events.RequestTypeTLS)

	logger := log.WithFields(log.Fields{})
	suite.ctx.Set(logLayerContextType, logger)
	suite.ctx.Set(metricsLayerContextType, stats.NewStats())
	suite.ctx.Set(startTimeLayerContextType, time.Time{})
	suite.ctx.Set(clientIDLayerContextType, "id")
}
