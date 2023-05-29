package config

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
)

// Config stores global configuration data of the application.
type Config struct {
	Debug                             bool     `toml:"debug"`
	DoNotVerifyCrawleraCert           bool     `toml:"dont_verify_crawlera_cert"`
	NoAutoSessions                    bool     `toml:"no_auto_sessions"`
	ConcurrentConnections             int      `toml:"concurrent_connections"`
	BindPort                          int      `toml:"bind_port"`
	CrawleraPort                      int      `toml:"crawlera_port"`
	ProxyAPIPort                      int      `toml:"proxy_api_port"`
	BindIP                            string   `toml:"bind_ip"`
	ProxyAPIIP                        string   `toml:"proxy_api_ip"`
	APIKey                            string   `toml:"api_key"`
	CrawleraHost                      string   `toml:"crawlera_host"`
	TLSCaCertificate                  string   `toml:"tls_ca_certificate"`
	TLSPrivateKey                     string   `toml:"tls_private_key"`
	AdblockLists                      []string `toml:"adblock_lists"`
	DirectAccessHostPathRegexps       []string `toml:"direct_access_hostpath_regexps"`
	DirectAccessExceptHostPathRegexps []string `toml:"direct_access_except_hostpath_regexps"`
	DirectAccessProxy                 string   `toml:"direct_access_proxy"`
	XHeaders                          map[string]string
}

// Bind returns a string for the http.ListenAndServe based on config
// information.
func (c *Config) Bind() string {
	return net.JoinHostPort(c.BindIP, strconv.Itoa(c.BindPort))
}

// CrawleraURL builds and returns URL to crawlera. Basically, this is required
// for http.ProxyURL to have embedded credentials etc.
func (c *Config) CrawleraURL() string {
	return fmt.Sprintf("http://%s:@%s",
		c.APIKey,
		net.JoinHostPort(c.CrawleraHost, strconv.Itoa(c.CrawleraPort)))
}

// MaybeSetNoAutoSessions defines is it is required to enable automatic
// session management or not.
func (c *Config) MaybeSetNoAutoSessions(value bool) {
	c.NoAutoSessions = c.NoAutoSessions || value
}

// MaybeSetDebug enabled debug mode of crawlera-headless-proxy (verbosity
// mostly). If given value is not defined (false) then changes nothing.
func (c *Config) MaybeSetDebug(value bool) {
	c.Debug = c.Debug || value
}

// MaybeSetConcurrentConnections sets a number of concurrent connections
// if necessary.
func (c *Config) MaybeSetConcurrentConnections(value int) {
	if value > 0 {
		c.ConcurrentConnections = value
	}
}

// MaybeDoNotVerifyCrawleraCert defines is it necessary to verify Crawlera
// TLS certificate. If given value is not defined (false) then changes nothing.
func (c *Config) MaybeDoNotVerifyCrawleraCert(value bool) {
	c.DoNotVerifyCrawleraCert = c.DoNotVerifyCrawleraCert || value
}

// MaybeSetBindIP sets an IP crawlera-headless-proxy should listen on.
// If given value is not defined (0) then changes nothing.
//
// If you want to have a global access (which is not recommended) please
// set it to 0.0.0.0.
func (c *Config) MaybeSetBindIP(value net.IP) {
	if value != nil {
		c.BindIP = value.String()
	}
}

// MaybeSetBindPort sets a port crawlera-headless-proxy should listen on.
// If given value is not defined (0) then changes nothing.
func (c *Config) MaybeSetBindPort(value int) {
	if value > 0 {
		c.BindPort = value
	}
}

// MaybeSetProxyAPIPort sets a port for own API of crawlera-headless-proxy.
// If given value is not defined (0) then changes nothing.
func (c *Config) MaybeSetProxyAPIPort(value int) {
	if value > 0 {
		c.ProxyAPIPort = value
	}
}

// MaybeSetProxyAPIIP sets an ip for own API of crawlera-headless-proxy.
// If given value is not defined ("") then changes nothing.
func (c *Config) MaybeSetProxyAPIIP(value net.IP) {
	if value != nil {
		c.ProxyAPIIP = value.String()
	}
}

// MaybeSetAPIKey sets an API key of Crawlera. If given value is not
// defined ("") then changes nothing.
func (c *Config) MaybeSetAPIKey(value string) {
	if value != "" {
		c.APIKey = value
	}
}

// MaybeSetCrawleraHost sets a host of Crawlera (usually it is
// 'proxy.crawlera.com'). If given value is not defined ("") then changes
// nothing.
func (c *Config) MaybeSetCrawleraHost(value string) {
	if value != "" {
		c.CrawleraHost = value
	}
}

// MaybeSetCrawleraPort a port Crawlera is listening to (usually it is 8010).
// If given value is not defined (0) then changes nothing.
func (c *Config) MaybeSetCrawleraPort(value int) {
	if value > 0 {
		c.CrawleraPort = value
	}
}

// MaybeSetTLSCaCertificate sets a content of the given file as TLS CA
// certificate.
func (c *Config) MaybeSetTLSCaCertificate(value string) {
	if value != "" {
		c.TLSCaCertificate = value
	}
}

// MaybeSetTLSPrivateKey sets a content of the given file as TLS
// private key.
func (c *Config) MaybeSetTLSPrivateKey(value string) {
	if value != "" {
		c.TLSPrivateKey = value
	}
}

// MaybeSetAdblockLists sets a list to URLs
func (c *Config) MaybeSetAdblockLists(value []string) {
	if len(value) > 0 {
		c.AdblockLists = value
	}
}

// MaybeSetDirectAccessHostPathRegexps sets a list of regular
// expressions for direct access.
func (c *Config) MaybeSetDirectAccessHostPathRegexps(value []string) {
	if len(value) > 0 {
		c.DirectAccessHostPathRegexps = value
	}
}

// MaybeSetDirectAccessExceptHostPathRegexps sets a list of regular
// expressions for proxied access. Takes priority over DirectAccessHostPathRegexps.
func (c *Config) MaybeSetDirectAccessExceptHostPathRegexps(value []string) {
	if len(value) > 0 {
		c.DirectAccessExceptHostPathRegexps = value
	}
}

func (c *Config) MaybeSetDirectAccessProxy(value string) {
	if len(value) > 0 {
		c.DirectAccessProxy = value
	}
}

// SetXHeader sets a header value of Crawlera X-Header. It is actually
// allowed to pass values in both ways: with full name (x-crawlera-profile)
// for example, and in the short form: just 'profile'. This effectively the
// same.
func (c *Config) SetXHeader(key, value string) {
	key = strings.ToLower(key)
	key = strings.TrimPrefix(key, "x-crawlera-")
	key = strings.Title(key)
	key = fmt.Sprintf("X-Crawlera-%s", key)

	c.XHeaders[key] = value
}

// Parse processes incoming file handler (usually, an instance of *os.File)
// and returns an instance of Config with fields set.
//
// Basically, new Config instance gets its fields in this order:
//   1. Defaults
//   2. Values from the config file.
func Parse(file io.Reader) (*Config, error) {
	buf, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("cannot read config file: %w", err)
	}

	conf := NewConfig()
	if _, err := toml.Decode(string(buf), conf); err != nil {
		return nil, fmt.Errorf("cannot parse config file: %w", err)
	}

	xheaders := conf.XHeaders
	conf.XHeaders = map[string]string{}

	for k, v := range xheaders {
		conf.SetXHeader(k, v)
	}

	return conf, nil
}

// NewConfig returns new instance of configuration data structure with
// fields set to sensible defaults.
func NewConfig() *Config {
	return &Config{
		AdblockLists: []string{},
		BindIP:       "127.0.0.1",
		BindPort:     3128, // nolint: gomnd
		ProxyAPIPort: 3129, // nolint: gomnd
		CrawleraHost: "proxy.zyte.com",
		CrawleraPort: 8011, // nolint: gomnd
		XHeaders:     map[string]string{},
	}
}
