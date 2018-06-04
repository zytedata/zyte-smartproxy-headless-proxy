package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	gock "gopkg.in/h2non/gock.v1"
)

func TestAdblock(t *testing.T) {
	defer gock.Off()
	gock.New("https://scrapinghub.com/testlist.txt").
		Get("/").
		Reply(200).
		BodyString("&ad_code=")

	cr := testInitNewProxyContainer()
	cr.conf.AdblockLists = []string{"https://scrapinghub.com/testlist.txt"}

	ware := NewAdblockMiddleware(cr.conf, nil, cr.s).(*adblockMiddleware)
	time.Sleep(10 * time.Millisecond)
	assert.True(t, ware.loaded)

	handler := ware.OnRequest()
	req := httptest.NewRequest("GET", "https://scrapinghub.com/testlist.txt", http.NoBody)
	nreq, nresp := handler(req, cr.Ctx())
	assert.Nil(t, nresp)
	assert.NotNil(t, nreq)

	req = httptest.NewRequest("GET", "https://scrapinghub.com/testlist.txt?id=1&ad_code=12", http.NoBody)
	nreq, nresp = handler(req, cr.Ctx())
	assert.True(t, nresp.StatusCode >= http.StatusBadRequest)
}
