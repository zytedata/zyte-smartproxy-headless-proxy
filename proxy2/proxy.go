package proxy2

import (
	"net/url"

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
	crawleraURL.Scheme = "http"
	executor, err := httransform.MakeProxyChainExecutor(crawleraURL)
	if err != nil {
		return nil, errors.Annotate(err, "Cannot make proxy chain executor")
	}

	opts := httransform.ServerOpts{
		CertCA:  []byte(conf.TLSCaCertificate),
		CertKey: []byte(conf.TLSPrivateKey),
	}

	srv, err := httransform.NewServer(opts,
		makeProxyLayers(conf, statsContainer),
		executor,
		&Logger{},
		statsContainer,
	)
	if err != nil {
		return nil, errors.Annotate(err, "Cannot create an instance of proxy")
	}

	return srv, nil
}

func makeProxyLayers(conf *config.Config, statsContainer *stats.Stats) []httransform.Layer {
	proxyLayers := []httransform.Layer{
		&layers.BaseLayer{
			Metrics: statsContainer,
		},
	}
	if len(conf.XHeaders) > 0 {
		proxyLayers = append(proxyLayers, &layers.XHeadersLayer{XHeaders: conf.XHeaders})
	}

	return proxyLayers
}
