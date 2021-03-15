package proxy

import (
	"fmt"
	"net/url"
	"time"

	"github.com/9seconds/httransform"

	"github.com/zytedata/zyte-headless-proxy/config"
	"github.com/zytedata/zyte-headless-proxy/layers"
	"github.com/zytedata/zyte-headless-proxy/stats"
)

func NewProxy(conf *config.Config, statsContainer *stats.Stats) (*httransform.Server, error) {
	smpURL, err := url.Parse(conf.SmartProxyManagerURL())
	if err != nil {
		return nil, fmt.Errorf("incorrect zyte smart proxy manager url: %w", err)
	}

	executor, err := httransform.MakeProxyChainExecutor(smpURL)
	if err != nil {
		return nil, fmt.Errorf("cannot make proxy chain executor: %w", err)
	}

	smpExecutor := func(state *httransform.LayerState) {
		startTime := time.Now()

		executor(state)
		statsContainer.NewSmartProxyManagerTime(time.Since(startTime))
		statsContainer.NewSmartProxyManagerRequest()
	}

	opts := httransform.ServerOpts{
		CertCA:           []byte(conf.TLSCaCertificate),
		CertKey:          []byte(conf.TLSPrivateKey),
		Executor:         smpExecutor,
		Logger:           &Logger{},
		Metrics:          statsContainer,
		Layers:           makeProxyLayers(conf, smpExecutor, statsContainer),
		OrganizationName: "Zyte",
	}
	if conf.Debug {
		opts.TracerPool = httransform.NewTracerPool(func() httransform.Tracer {
			return &httransform.LogTracer{}
		})
	}

	srv, err := httransform.NewServer(opts)
	if err != nil {
		return nil, fmt.Errorf("cannot create an instance of proxy: %w", err)
	}

	return srv, nil
}

func makeProxyLayers(conf *config.Config, smpExecutor httransform.Executor, statsContainer *stats.Stats) []httransform.Layer {
	proxyLayers := []httransform.Layer{
		layers.NewBaseLayer(statsContainer),
	}

	if len(conf.AdblockLists) > 0 {
		proxyLayers = append(proxyLayers, layers.NewAdblockLayer(conf.AdblockLists))
	}

	if len(conf.DirectAccessHostPathRegexps) > 0 {
		proxyLayers = append(proxyLayers, layers.NewDirectAccessLayer(conf.DirectAccessHostPathRegexps))
	}

	if conf.ConcurrentConnections > 0 {
		proxyLayers = append(proxyLayers, layers.NewRateLimiterLayer(conf.ConcurrentConnections))
	}

	if len(conf.XHeaders) > 0 {
		proxyLayers = append(proxyLayers, layers.NewXHeadersLayer(conf.XHeaders))
	}

	proxyLayers = append(proxyLayers, layers.NewRefererLayer())

	if !conf.NoAutoSessions {
		proxyLayers = append(proxyLayers, layers.NewSessionsLayer(conf, smpExecutor))
	}

	return proxyLayers
}
