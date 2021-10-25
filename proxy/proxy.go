package proxy

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/9seconds/httransform/v2"
	"github.com/9seconds/httransform/v2/dialers"
	"github.com/9seconds/httransform/v2/executor"
	"github.com/9seconds/httransform/v2/layers"

	"github.com/scrapinghub/crawlera-headless-proxy/config"
	customs "github.com/scrapinghub/crawlera-headless-proxy/layers"
	"github.com/scrapinghub/crawlera-headless-proxy/stats"
)

func NewProxy(conf *config.Config, statsContainer *stats.Stats) (*httransform.Server, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for range signals {
			cancel()
		}
	}()

	_, err := url.Parse(conf.CrawleraURL())
	if err != nil {
		return nil, fmt.Errorf("incorrect crawlera url: %w", err)
	}

	dialer, err := dialers.DialerFromURL(dialers.Opts{}, conf.CrawleraURL())
	if err != nil {
		return nil, fmt.Errorf("dialer error: %w", err)
	}
	exe := executor.MakeDefaultExecutor(dialer)

	opts := httransform.ServerOpts{
		Layers:        makeProxyLayers(conf, exe, statsContainer),
		Executor:      exe,
		TLSCertCA:     []byte(conf.TLSCaCertificate),
		TLSPrivateKey: []byte(conf.TLSPrivateKey),
	}
	/*if conf.Debug {
		opts.TracerPool = httransform.NewTracerPool(func() httransform.Tracer {
			return &httransform.LogTracer{}
		})
	}*/

	srv, err := httransform.NewServer(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("cannot create an instance of proxy: %w", err)
	}

	return srv, nil
}

func makeProxyLayers(conf *config.Config, exe executor.Executor, statsContainer *stats.Stats) []layers.Layer {
	proxyLayers := []layers.Layer{
		customs.NewAuthLayer(conf.APIKey),
		customs.NewBaseLayer(statsContainer),
	}

	if len(conf.AdblockLists) > 0 {
		proxyLayers = append(proxyLayers, customs.NewAdblockLayer(conf.AdblockLists))
	}

	if len(conf.DirectAccessHostPathRegexps) > 0 {
		proxyLayers = append(proxyLayers, customs.NewDirectAccessLayer(conf.DirectAccessHostPathRegexps))
	}

	if conf.ConcurrentConnections > 0 {
		proxyLayers = append(proxyLayers, customs.NewRateLimiterLayer(conf.ConcurrentConnections))
	}

	if len(conf.XHeaders) > 0 {
		proxyLayers = append(proxyLayers, customs.NewXHeadersLayer(conf.XHeaders))
	}

	proxyLayers = append(proxyLayers, customs.NewRefererLayer())

	if !conf.NoAutoSessions {
		proxyLayers = append(proxyLayers, customs.NewSessionsLayer(conf, exe))
	}

	return proxyLayers
}
