package proxy

import (
	"net/url"
	"time"

	"github.com/9seconds/httransform"
	"github.com/juju/errors"

	"github.com/scrapinghub/crawlera-headless-proxy/config"
	"github.com/scrapinghub/crawlera-headless-proxy/layers"
	"github.com/scrapinghub/crawlera-headless-proxy/stats"
)

func NewProxy(conf *config.Config, statsContainer *stats.Stats) (*httransform.Server, error) {
	crawleraURL, err := url.Parse(conf.CrawleraURL())
	if err != nil {
		return nil, errors.Annotate(err, "Incorrect Crawlera URL")
	}

	executor, err := httransform.MakeProxyChainExecutor(crawleraURL)
	if err != nil {
		return nil, errors.Annotate(err, "Cannot make proxy chain executor")
	}
	crawleraExecutor := func(state *httransform.LayerState) {
		startTime := time.Now()
		executor(state)
		statsContainer.NewCrawleraTime(time.Since(startTime))
		statsContainer.NewCrawleraRequest()
	}

	opts := httransform.ServerOpts{
		CertCA:   []byte(conf.TLSCaCertificate),
		CertKey:  []byte(conf.TLSPrivateKey),
		Executor: crawleraExecutor,
		Logger:   &Logger{},
		Metrics:  statsContainer,
		Layers:   makeProxyLayers(conf, crawleraExecutor, statsContainer),
	}
	if conf.Debug {
		opts.TracerPool = httransform.NewTracerPool(func() httransform.Tracer {
			return &httransform.LogTracer{}
		})
	}

	srv, err := httransform.NewServer(opts)
	if err != nil {
		return nil, errors.Annotate(err, "Cannot create an instance of proxy")
	}

	return srv, nil
}

func makeProxyLayers(conf *config.Config, crawleraExecutor httransform.Executor, statsContainer *stats.Stats) []httransform.Layer {
	proxyLayers := []httransform.Layer{
		layers.NewBaseLayer(statsContainer),
	}

	if len(conf.AdblockLists) > 0 {
		proxyLayers = append(proxyLayers, layers.NewAdblockLayer(conf.AdblockLists))
	}
	if len(conf.PathSuffixesWithDirectAccess) > 0 {
		proxyLayers = append(proxyLayers, layers.NewDirectAccessLayer(conf.PathSuffixesWithDirectAccess))
	}
	if conf.ConcurrentConnections > 0 {
		proxyLayers = append(proxyLayers, layers.NewRateLimiterLayer(conf.ConcurrentConnections))
	}

	if len(conf.XHeaders) > 0 {
		proxyLayers = append(proxyLayers, layers.NewXHeadersLayer(conf.XHeaders))
	}

	proxyLayers = append(proxyLayers, layers.NewRefererLayer())

	if !conf.NoAutoSessions {
		proxyLayers = append(proxyLayers, layers.NewSessionsLayer(conf, crawleraExecutor))
	}

	return proxyLayers
}
