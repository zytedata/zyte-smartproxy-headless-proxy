package layers

import (
	"net/http"
	"testing"
	"time"

	"github.com/juju/errors"
	"github.com/stretchr/testify/suite"
	gock "gopkg.in/h2non/gock.v1"
)

type AdblockLayerTestSuite struct {
	CommonLayerTestSuite

	lists []string
	layer *AdblockLayer
}

func (suite *AdblockLayerTestSuite) SetupTest() {
	suite.CommonLayerTestSuite.SetupTest()
	gock.New("https://scrapinghub.com/testlist.txt").
		Get("/").
		Reply(200).
		BodyString("ad_code=")

	suite.lists = []string{"https://scrapinghub.com/testlist.txt"}
	suite.layer = NewAdblockLayer(suite.lists).(*AdblockLayer)
}

func (suite *AdblockLayerTestSuite) TearDownTest() {
	gock.Off()
}

func (suite *AdblockLayerTestSuite) TestPass() {
	time.Sleep(10 * time.Millisecond)
	suite.True(suite.layer.loaded)

	suite.state.RequestHeaders.SetString("host", "scrapinghub.com")
	suite.state.Request.SetRequestURI("https://scrapinghub.com/testlist.txt")
	suite.Nil(suite.layer.OnRequest(suite.state))
}

func (suite *AdblockLayerTestSuite) TestPassOnResponse() {
	time.Sleep(10 * time.Millisecond)
	suite.True(suite.layer.loaded)

	suite.layer.OnResponse(suite.state, errors.New("Unexpected"))
	suite.Equal(suite.state.Response.Header.StatusCode(), http.StatusOK)
}

func (suite *AdblockLayerTestSuite) TestDontPassOnResponse() {
	time.Sleep(10 * time.Millisecond)
	suite.True(suite.layer.loaded)

	suite.layer.OnResponse(suite.state, errAdblockedRequest)
	suite.NotEqual(suite.state.Response.Header.StatusCode(), http.StatusOK)
}

func (suite *AdblockLayerTestSuite) TestDontPass() {
	time.Sleep(10 * time.Millisecond)
	suite.True(suite.layer.loaded)

	suite.state.RequestHeaders.SetString("host", "scrapinghub.com")
	suite.state.Request.SetRequestURI("https://scrapinghub.com/testlist.txt/?ad_code=111")
	suite.Equal(suite.layer.OnRequest(suite.state), errAdblockedRequest)
}

func TestAdblockLayer(t *testing.T) {
	suite.Run(t, &AdblockLayerTestSuite{})
}
