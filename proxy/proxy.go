package proxy

import (
	"context"
	"fmt"
	"net/url"

	"github.com/9seconds/httransform/v2"
	"github.com/9seconds/httransform/v2/dialers"
	"github.com/9seconds/httransform/v2/executor"
	"github.com/9seconds/httransform/v2/layers"

	"github.com/scrapinghub/crawlera-headless-proxy/config"
	customs "github.com/scrapinghub/crawlera-headless-proxy/layers"
	"github.com/scrapinghub/crawlera-headless-proxy/stats"
)

func NewProxy(conf *config.Config, statsContainer *stats.Stats, ctx *context.Context) (*httransform.Server, error) {
	_, err := url.Parse(conf.CrawleraURL())
	if err != nil {
		return nil, fmt.Errorf("incorrect crawlera url: %w", err)
	}

	dialer, err := dialers.DialerFromURL(dialers.Opts{}, conf.CrawleraURL())
	if err != nil {
		return nil, fmt.Errorf("dialer error: %w", err)
	}

	crawleraExecutor := executor.MakeDefaultExecutor(dialer)

	opts := httransform.ServerOpts{
		Layers:        makeProxyLayers(conf, crawleraExecutor, statsContainer),
		Executor:      crawleraExecutor,
		TLSCertCA:     []byte(conf.TLSCaCertificate),
		TLSPrivateKey: []byte(conf.TLSPrivateKey),
	}

	srv, err := httransform.NewServer(*ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("cannot create an instance of proxy: %w", err)
	}

	return srv, nil
}

func makeProxyLayers(conf *config.Config, crawleraExecutor executor.Executor, statsContainer *stats.Stats) []layers.Layer {
	proxyLayers := []layers.Layer{
		customs.NewBaseLayer(statsContainer),
	}

	if len(conf.AdblockLists) > 0 {
		proxyLayers = append(proxyLayers, customs.NewAdblockLayer(conf.AdblockLists))
	}

	if len(conf.DirectAccessHostPathRegexps) > 0 {
		proxyLayers = append(proxyLayers, customs.NewDirectAccessLayer(conf.DirectAccessHostPathRegexps, conf.DirectAccessExceptHostPathRegexps, conf.DirectAccessProxy))
	}

	proxyLayers = append(proxyLayers, customs.NewAuthLayer(conf.APIKey))

	if conf.ConcurrentConnections > 0 {
		proxyLayers = append(proxyLayers, customs.NewRateLimiterLayer(conf.ConcurrentConnections))
	}

	if len(conf.XHeaders) > 0 {
		proxyLayers = append(proxyLayers, customs.NewXHeadersLayer(conf.XHeaders))
	}

	if !conf.NoAutoSessions {
		proxyLayers = append(proxyLayers, customs.NewSessionsLayer(conf, crawleraExecutor))
	}

	return proxyLayers
}
