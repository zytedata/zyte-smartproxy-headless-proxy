package proxy2

import (
	"net/url"

	"github.com/9seconds/httransform"
	"github.com/juju/errors"

	"github.com/scrapinghub/crawlera-headless-proxy/config"
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

	return nil, nil
}
