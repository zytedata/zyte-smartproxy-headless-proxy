package config

import (
	"bytes"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigDefault(t *testing.T) {
	conf := NewConfig()

	assert.False(t, conf.Debug)
	assert.False(t, conf.DoNotVerifyCrawleraCert)
	assert.Equal(t, conf.BindPort, 3128)
	assert.Equal(t, conf.CrawleraPort, 8010)
	assert.Equal(t, conf.BindIP, "127.0.0.1")
	assert.Equal(t, conf.APIKey, "")
	assert.Equal(t, conf.CrawleraHost, "proxy.crawlera.com")
	assert.Len(t, conf.XHeaders, 0)
}

func TestConfigMaybeSetDebug(t *testing.T) {
	conf := NewConfig()

	conf.MaybeSetDebug(false)
	assert.False(t, conf.Debug)

	conf.MaybeSetDebug(true)
	assert.True(t, conf.Debug)

	conf.MaybeSetDebug(false)
	assert.True(t, conf.Debug)
}

func TestConfigMaybeDoNotVerifyCrawleraCert(t *testing.T) {
	conf := NewConfig()

	conf.MaybeDoNotVerifyCrawleraCert(false)
	assert.False(t, conf.DoNotVerifyCrawleraCert)

	conf.MaybeDoNotVerifyCrawleraCert(true)
	assert.True(t, conf.DoNotVerifyCrawleraCert)

	conf.MaybeDoNotVerifyCrawleraCert(false)
	assert.True(t, conf.DoNotVerifyCrawleraCert)
}

func TestConfigMaybeSetBindIP(t *testing.T) {
	conf := NewConfig()

	assert.Equal(t, conf.Bind(), "127.0.0.1:3128")

	conf.MaybeSetBindIP(net.ParseIP("10.0.0.10"))
	assert.Equal(t, conf.Bind(), "10.0.0.10:3128")

	conf.MaybeSetBindIP(nil)
	assert.Equal(t, conf.Bind(), "10.0.0.10:3128")
}

func TestConfigMaybeSetBindPort(t *testing.T) {
	conf := NewConfig()

	assert.Equal(t, conf.Bind(), "127.0.0.1:3128")

	conf.MaybeSetBindPort(3000)
	assert.Equal(t, conf.Bind(), "127.0.0.1:3000")

	conf.MaybeSetBindPort(0)
	assert.Equal(t, conf.Bind(), "127.0.0.1:3000")
}

func TestConfigMaybeSetAPIKey(t *testing.T) {
	conf := NewConfig()

	assert.Equal(t, conf.APIKey, "")

	conf.MaybeSetAPIKey("111")
	assert.Equal(t, conf.APIKey, "111")

	conf.MaybeSetAPIKey("")
	assert.Equal(t, conf.APIKey, "111")
}

func TestMaybeSetCrawleraHost(t *testing.T) {
	conf := NewConfig()

	assert.Equal(t, conf.CrawleraURL(), "http://:@proxy.crawlera.com:8010")

	conf.MaybeSetAPIKey("111")
	assert.Equal(t, conf.CrawleraURL(), "http://111:@proxy.crawlera.com:8010")

	conf.MaybeSetCrawleraPort(0)
	conf.MaybeSetCrawleraHost("")
	assert.Equal(t, conf.CrawleraURL(), "http://111:@proxy.crawlera.com:8010")

	conf.MaybeSetCrawleraHost("localhost")
	assert.Equal(t, conf.CrawleraURL(), "http://111:@localhost:8010")

	conf.MaybeSetCrawleraPort(3333)
	assert.Equal(t, conf.CrawleraURL(), "http://111:@localhost:3333")
}

func TestSetXHeader(t *testing.T) {
	conf := NewConfig()

	conf.SetXHeader("X-Crawlera-Profile", "desktop")
	assert.Equal(t, conf.XHeaders["X-Crawlera-Profile"], "desktop")

	conf.SetXHeader("x-Crawlera-PROFILE", "mobile")
	assert.Equal(t, conf.XHeaders["X-Crawlera-Profile"], "mobile")

	conf.SetXHeader("PROFILE", "desktop")
	assert.Equal(t, conf.XHeaders["X-Crawlera-Profile"], "desktop")
}

func TestParse(t *testing.T) {
	text := bytes.NewBufferString(`debug = true
crawlera_port = 8010
[xheaders]
profile = "mobile"`)
	conf, err := Parse(text)

	assert.Nil(t, err)
	assert.True(t, conf.Debug)
	assert.Equal(t, conf.CrawleraPort, 8010)
	assert.Len(t, conf.XHeaders, 1)
	assert.Equal(t, conf.XHeaders["X-Crawlera-Profile"], "mobile")
}

func TestParseFailed(t *testing.T) {
	text := bytes.NewBufferString(`debug = sajhfskdfjsdfk
[xheaders]
profile = mobile`)
	_, err := Parse(text)

	assert.Error(t, err)
}
