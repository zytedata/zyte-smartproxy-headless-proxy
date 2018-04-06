package main

//go:generate scripts/generate_certs.sh

import (
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/9seconds/crawlera-headless-proxy/config"
	"github.com/9seconds/crawlera-headless-proxy/proxy"
)

var (
	app = kingpin.New("crawlera-headless-proxy",
		"Local proxy for Crawlera to be used with headless browsers.")

	debug = app.Flag("debug",
		"Run in debug mode.").
		Short('d').
		Envar("CRAWLERA_HEADLESS_DEBUG").
		Bool()
	bindIP = app.Flag("bind-ip",
		"IP to bind to. Default is 127.0.0.1.").
		Short('b').
		Envar("CRAWLERA_HEADLESS_IP").
		IP()
	bindPort = app.Flag("port",
		"Port to bind to. Default is 3128.").
		Short('p').
		Envar("CRAWLERA_HEADLESS_PORT").
		Int()
	configFileName = app.Flag("config",
		"Path to configuration file.").
		Short('c').
		Envar("CRAWLERA_HEADLESS_CONFIGPATH").
		File()
	apiKey = app.Flag("api-key",
		"API key to Crawlera.").
		Short('a').
		Envar("CRAWLERA_HEADLESS_APIKEY").
		String()
	crawleraHost = app.Flag("crawlera-host",
		"Hostname of Crawlera. Default is proxy.crawlera.com.").
		Short('u').
		Envar("CRAWLERA_HEADLESS_CHOST").
		String()
	crawleraPort = app.Flag("crawlera-port",
		"Port of Crawlera. Default is 8010.").
		Short('o').
		Envar("CRAWLERA_HEADLESS_CPORT").
		Int()
	VerifyCrawleraCert = app.Flag("verify-crawlera-cert",
		"Verify Crawlera certificate on proxy responses").
		Short('v').
		Envar("CRAWLERA_HEADLESS_VERIFYCERT").
		Bool()
	xheaders = app.Flag("xheader",
		"Crawlera X-Headers.").
		Short('x').
		Envar("CRAWLERA_HEADLESS_XHEADERS").
		StringMap()
)

func init() {
	app.Version("0.0.1")
	log.SetFormatter(&log.TextFormatter{})
	log.SetLevel(log.WarnLevel)
}

func main() {
	kingpin.MustParse(app.Parse(os.Args[1:]))

	conf, err := getConfig()
	if err != nil {
		log.Errorf("Cannot get configuration: %s", err)
		os.Exit(1)
	}
	if conf.Debug {
		log.SetLevel(log.DebugLevel)
	}

	listen := conf.Bind()
	log.WithFields(log.Fields{
		"debug":                conf.Debug,
		"bindip":               conf.BindIP,
		"bindport":             conf.BindPort,
		"apikey":               conf.APIKey,
		"crawlera-host":        conf.CrawleraHost,
		"crawlera-port":        conf.CrawleraPort,
		"xheaders":             conf.XHeaders,
		"verify-crawlera-cert": conf.VerifyCrawleraCert,
	}).Debugf("Listen on %s", listen)

	log.Fatal(http.ListenAndServe(listen, proxy.NewProxy(conf)))
}

func getConfig() (*config.Config, error) {
	conf := config.NewConfig()
	if *configFileName != nil {
		newConf, err := config.Parse(*configFileName)
		if err != nil {
			return nil, err
		}
		conf = newConf
	}

	conf.MaybeSetDebug(*debug)
	conf.MaybeVerifyCrawleraCert(*VerifyCrawleraCert)
	conf.MaybeSetBindIP(*bindIP)
	conf.MaybeSetBindPort(*bindPort)
	conf.MaybeSetAPIKey(*apiKey)
	conf.MaybeSetCrawleraHost(*crawleraHost)
	conf.MaybeSetCrawleraPort(*crawleraPort)
	for k, v := range *xheaders {
		conf.SetXHeader(k, v)
	}

	return conf, nil
}
