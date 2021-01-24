package proxy

import (
	"fmt"
	"net/url"
	"time"

	"github.com/9seconds/httransform"

	"github.com/zytegroup/zyte-proxy-headless-proxy/config"
	"github.com/zytegroup/zyte-proxy-headless-proxy/layers"
	"github.com/zytegroup/zyte-proxy-headless-proxy/stats"
)

func NewProxy(conf *config.Config, statsContainer *stats.Stats) (*httransform.Server, error) {
	zyteProxyURL, err := url.Parse(conf.ZyteProxyURL())
	if err != nil {
		return nil, fmt.Errorf("incorrect Zyte Smart Proxy Manager url: %w", err)
	}

	executor, err := httransform.MakeProxyChainExecutor(zyteProxyURL)
	if err != nil {
		return nil, fmt.Errorf("cannot make proxy chain executor: %w", err)
	}

	zyteProxyExecutor := func(state *httransform.LayerState) {
		startTime := time.Now()

		executor(state)
		statsContainer.NewZyteProxyTime(time.Since(startTime))
		statsContainer.NewZyteProxyRequest()
	}

	opts := httransform.ServerOpts{
		CertCA:           []byte(conf.TLSCaCertificate),
		CertKey:          []byte(conf.TLSPrivateKey),
		Executor:         zyteProxyExecutor,
		Logger:           &Logger{},
		Metrics:          statsContainer,
		Layers:           makeProxyLayers(conf, zyteProxyExecutor, statsContainer),
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

func makeProxyLayers(conf *config.Config, zyteProxyExecutor httransform.Executor, statsContainer *stats.Stats) []httransform.Layer {
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

	if len(conf.ZyteProxyHeaders) > 0 {
		proxyLayers = append(proxyLayers, layers.NewZyteProxyHeadersLayer(conf.ZyteProxyHeaders))
	}

	proxyLayers = append(proxyLayers, layers.NewRefererLayer())

	if !conf.NoAutoSessions {
		proxyLayers = append(proxyLayers, layers.NewSessionsLayer(conf, zyteProxyExecutor))
	}

	return proxyLayers
}
