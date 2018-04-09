package proxy

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrepareForCrawleraProfile(t *testing.T) {
	headers := http.Header{}
	headers.Add("aCCEPt", "*/*")
	headers.Add("accept-Language", "en")
	headers.Add("user-agent", "hello!")
	headers.Add("Cookie", "k=v")
	fakeURL, _ := url.Parse("https://scrapinghub.com/crawlera?hello=1")

	prepareForCrawleraProfile(headers, fakeURL)
	assert.Equal(t, headers.Get("Accept"), "")
	assert.Equal(t, headers.Get("Accept-Language"), "")
	assert.Equal(t, headers.Get("Accept-Language"), "")
	assert.Equal(t, headers.Get("user-agent"), "")
	assert.Equal(t, headers.Get("Cookie"), "k=v")
	assert.Equal(t, headers.Get("Referer"), "https://scrapinghub.com/crawlera")
}
