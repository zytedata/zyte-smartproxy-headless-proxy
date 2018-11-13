package main

import (
	"bytes"
	"crypto/sha1" // nolint: gosec
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/juju/errors"
	log "github.com/sirupsen/logrus"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/scrapinghub/crawlera-headless-proxy/config"
	"github.com/scrapinghub/crawlera-headless-proxy/proxy"
	"github.com/scrapinghub/crawlera-headless-proxy/stats"
)

var version = "dev" // nolint: gochecknoglobals

var ( // nolint: gochecknoglobals
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
		Envar("CRAWLERA_HEADLESS_BINDIP").
		IP()
	proxyAPIIP = app.Flag("proxy-api-ip",
		"IP to bind proxy API to. Default is the bind-ip value.").
		Short('m').
		Envar("CRAWLERA_HEADLESS_PROXYAPIIP").
		IP()
	bindPort = app.Flag("bind-port",
		"Port to bind to. Default is 3128.").
		Short('p').
		Envar("CRAWLERA_HEADLESS_BINDPORT").
		Int()
	proxyAPIPort = app.Flag("proxy-api-port",
		"Port to bind proxy api to. Default is 3130.").
		Short('w').
		Envar("CRAWLERA_HEADLESS_PROXYAPIPORT").
		Int()
	configFileName = app.Flag("config",
		"Path to configuration file.").
		Short('c').
		Envar("CRAWLERA_HEADLESS_CONFIG").
		File()
	tlsCaCertificate = app.Flag("tls-ca-certificate",
		"Path to TLS CA certificate file.").
		Short('l').
		Envar("CRAWLERA_HEADLESS_TLSCACERTPATH").
		ExistingFile()
	tlsPrivateKey = app.Flag("tls-private-key",
		"Path to TLS private key.").
		Short('r').
		Envar("CRAWLERA_HEADLESS_TLSPRIVATEKEYPATH").
		ExistingFile()
	noAutoSessions = app.Flag("no-auto-sessions",
		"Disable automatic session management.").
		Short('t').
		Envar("CRAWLERA_HEADLESS_NOAUTOSESSIONS").
		Bool()
	concurrentConnections = app.Flag("concurrent-connections",
		"Number of concurrent connections.").
		Short('n').
		Envar("CRAWLERA_HEADLESS_CONCURRENCY").
		Int()
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
	doNotVerifyCrawleraCert = app.Flag("dont-verify-crawlera-cert",
		"Do not verify Crawlera own certificate.").
		Short('v').
		Envar("CRAWLERA_HEADLESS_DONTVERIFY").
		Bool()
	xheaders = app.Flag("xheader",
		"Crawlera X-Headers.").
		Short('x').
		Envar("CRAWLERA_HEADLESS_XHEADERS").
		StringMap()
	adblockLists = app.Flag("adblock-list",
		"A list to requests to filter out (ADBlock compatible).").
		Short('k').
		Envar("CRAWLERA_HEADLESS_ADBLOCKLISTS").
		Strings()
)

func main() {
	app.Version(version)
	log.SetFormatter(&log.TextFormatter{})
	log.SetLevel(log.WarnLevel)

	kingpin.MustParse(app.Parse(os.Args[1:]))

	conf, err := getConfig()
	if err != nil {
		log.Errorf("Cannot get configuration: %s", err)
		os.Exit(1)
	}
	if conf.Debug {
		log.SetLevel(log.DebugLevel)
	}
	if conf.APIKey == "" {
		log.Fatal("API key is not set")
	}
	if err = initCertificates(conf); err != nil {
		log.Fatal(err)
	}

	listen := conf.Bind()
	log.WithFields(log.Fields{
		"debug":                     conf.Debug,
		"adblock-lists":             conf.AdblockLists,
		"no-auto-sessions":          conf.NoAutoSessions,
		"apikey":                    conf.APIKey,
		"bindip":                    conf.BindIP,
		"bindport":                  conf.BindPort,
		"proxy-api-ip":              conf.ProxyAPIIP,
		"proxy-api-port":            conf.ProxyAPIPort,
		"crawlera-host":             conf.CrawleraHost,
		"crawlera-port":             conf.CrawleraPort,
		"dont-verify-crawlera-cert": conf.DoNotVerifyCrawleraCert,
		"concurrent-connections":    conf.ConcurrentConnections,
		"xheaders":                  conf.XHeaders,
	}).Debugf("Listen on %s", listen)

	statsContainer := stats.NewStats()
	go stats.RunStats(statsContainer, conf)

	if crawleraProxy, err := proxy.NewProxy(conf, statsContainer); err == nil {
		log.Fatal(http.ListenAndServe(listen, crawleraProxy))
	} else {
		log.Fatal(err)
	}
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
	conf.MaybeDoNotVerifyCrawleraCert(*doNotVerifyCrawleraCert)
	conf.MaybeSetAdblockLists(*adblockLists)
	conf.MaybeSetAPIKey(*apiKey)
	conf.MaybeSetBindIP(*bindIP)
	conf.MaybeSetBindPort(*bindPort)
	conf.MaybeSetConcurrentConnections(*concurrentConnections)
	conf.MaybeSetCrawleraHost(*crawleraHost)
	conf.MaybeSetCrawleraPort(*crawleraPort)
	conf.MaybeSetNoAutoSessions(*noAutoSessions)
	conf.MaybeSetTLSCaCertificate(*tlsCaCertificate)
	conf.MaybeSetTLSPrivateKey(*tlsPrivateKey)
	conf.MaybeSetProxyAPIIP(*proxyAPIIP)
	conf.MaybeSetProxyAPIPort(*proxyAPIPort)
	for k, v := range *xheaders {
		conf.SetXHeader(k, v)
	}

	if conf.ProxyAPIIP == "" {
		conf.ProxyAPIIP = conf.BindIP
	}

	return conf, nil
}

func initCertificates(conf *config.Config) (err error) {
	caCertificate := proxy.DefaultCertCA
	privateKey := proxy.DefaultPrivateKey

	if conf.TLSCaCertificate != "" {
		caCertificate, err = ioutil.ReadFile(conf.TLSCaCertificate)
		if err != nil {
			return errors.Annotate(err, "Cannot read TLS CA certificate")
		}
	}
	if conf.TLSPrivateKey != "" {
		privateKey, err = ioutil.ReadFile(conf.TLSPrivateKey)
		if err != nil {
			return errors.Annotate(err, "Cannot read TLS private key")
		}
	}

	caCertificate = bytes.TrimSpace(caCertificate)
	privateKey = bytes.TrimSpace(privateKey)

	log.WithFields(log.Fields{
		"ca-cert":  fmt.Sprintf("%x", sha1.Sum(caCertificate)), // nolint: gosec
		"priv-key": fmt.Sprintf("%x", sha1.Sum(privateKey)),    // nolint: gosec
	}).Debug("TLS checksums.")

	return proxy.InitCertificates(caCertificate, privateKey)
}
