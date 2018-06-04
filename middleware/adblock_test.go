package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	gock "gopkg.in/h2non/gock.v1"
)

type AdblockTestSuite struct {
	MiddlewareTestSuite
}

func (t *AdblockTestSuite) TestWork() {
	defer gock.Off()
	gock.New("https://scrapinghub.com/testlist.txt").
		Get("/").
		Reply(200).
		BodyString("&ad_code=")

	t.cr.conf.AdblockLists = []string{"https://scrapinghub.com/testlist.txt"}

	ware := NewAdblockMiddleware(t.cr.conf, nil, t.cr.s).(*adblockMiddleware)
	time.Sleep(10 * time.Millisecond)
	t.True(ware.loaded)

	handler := ware.OnRequest()
	req := httptest.NewRequest("GET", "https://scrapinghub.com/testlist.txt", http.NoBody)
	nreq, nresp := handler(req, t.cr.Ctx())
	t.Nil(nresp)
	t.NotNil(nreq)

	req = httptest.NewRequest("GET", "https://scrapinghub.com/testlist.txt?id=1&ad_code=12", http.NoBody)
	nreq, nresp = handler(req, t.cr.Ctx())
	t.True(nresp.StatusCode >= http.StatusBadRequest)
}

func TestAdblock(t *testing.T) {
	suite.Run(t, &AdblockTestSuite{})
}
