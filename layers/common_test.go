package layers

import (
	"time"

	"github.com/9seconds/httransform"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"github.com/valyala/fasthttp"

	"github.com/scrapinghub/crawlera-headless-proxy/stats"
)

type CommonLayerTestSuite struct {
	suite.Suite

	state *httransform.LayerState
}

func (suite *CommonLayerTestSuite) SetupTest() {
	suite.state = &httransform.LayerState{
		Request:         &fasthttp.Request{},
		Response:        &fasthttp.Response{},
		RequestHeaders:  &httransform.HeaderSet{},
		ResponseHeaders: &httransform.HeaderSet{},
	}
	suite.state.Set(logLayerContextType, &log.Entry{})
	suite.state.Set(metricsLayerContextType, stats.NewStats())
	suite.state.Set(startTimeLayerContextType, time.Time{})
	suite.state.Set(clientIDLayerContextType, "id")
}
