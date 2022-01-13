package layers

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type RateLimiterLayerTestSuite struct {
	suite.Suite
}

func (suite *RateLimiterLayerTestSuite) TestRateLimiter() {
	layer := NewRateLimiterLayer(2).(*RateLimiterLayer)

	suite.Nil(layer.OnRequest(nil))
	suite.Len(layer.limiter, 1)
	layer.OnResponse(nil, nil) // nolint:errcheck
	suite.Len(layer.limiter, 0)
}

func TestRateLimiter(t *testing.T) {
	suite.Run(t, &RateLimiterLayerTestSuite{})
}
