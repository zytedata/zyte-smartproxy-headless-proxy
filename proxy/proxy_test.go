package proxy

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrepareForCrawleraProfile(t *testing.T) {
	headers := http.Header{}
	headers.Add("aCCEPt", "*/*")
	headers.Add("accept-Language", "en")
	headers.Add("user-agent", "hello!")
	headers.Add("Cookie", "k=v")

	prepareForCrawleraProfile(headers, "https://scrapinghub.com/crawlera?hello=1")

	assert.Equal(t, headers.Get("Accept"), "")
	assert.Equal(t, headers.Get("Accept-Language"), "en")
	assert.Equal(t, headers.Get("user-agent"), "")
	assert.Equal(t, headers.Get("Cookie"), "k=v")
	assert.Equal(t, headers.Get("Referer"), "https://scrapinghub.com/crawlera")
}

func TestPrepareHostKeepPort(t *testing.T) {
	headers := http.Header{}
	headers.Add("Referer", "https://scrapinghub.com:9999/crawlera")

	prepareForCrawleraProfile(headers, "https://scrapinghub.com/crawlera?hello=1")

	assert.Equal(t, headers.Get("Referer"), "https://scrapinghub.com:9999/crawlera")
}
