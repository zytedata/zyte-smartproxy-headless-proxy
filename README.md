# Crawlera Headless Proxy

Crawlera Headless proxy is a proxy which main intent
is to help users with headless browsers to use
[Crawlera](https://scrapinghub.com/crawlera). These
includes different implementations of headless browsers
such as [Splash](https://scrapinghub.com/splash),
headless [Chrome](https://google.com/chrome/) and
[Firefox](https://www.mozilla.org/en-US/firefox/).
Also, this proxy should help users of such frameworks
as [Selenium](https://www.seleniumhq.org/) and
[Puppeteer](https://github.com/GoogleChrome/puppeteer) to use Crawlera
without a need to build [Squid](http://www.squid-cache.org/) chains or
install [Polipo](https://www.irif.fr/~jch/software/polipo/).

The biggest problem with headless browsers is their configuration:

1. Crawlera uses proxy authentication protocol described in
   [RFC 7235](https://tools.ietf.org/html/rfc7235#section-4.3) but it is
   rather hard to configure such authentication in headless browsers. The
   most popular way of bypassing this problem is to use Polipo which is,
   unfortunately, unsupported for a long time.
2. Crawlera uses
   [X-Headers as configuration](https://doc.scrapinghub.com/crawlera.html#request-headers).
   To use this API with headless browsers, users have to install plugins or
   extensions in their browsers and configure them to propagate such headers
   to Crawlera.
3. Also, it is rather hard and complex to maintain best practices of using
   these headers. For example,
   [support of Browser Profiles](https://doc.scrapinghub.com/crawlera.html#x-crawlera-profile)
   requires to have a minimal possible set headers. For example, it is
   recommended to remove `Accept` header by default. It is rather hard
   to do that using headless browsers API.

Crawlera Headless Proxy intended to help users to avoid such
problems. You should generally think about it as a proxy which should
be accessible by your headless browser of Selenium grid. This proxy
propagates your requests to Crawlera maintaining API key and injecting
headers into the requests. Basically, you have to do a bare minimum:

1. Get Crawlera API key
2. Run this proxy on your local machine or any machine accessible by
   headless browser, configuring it with configuration file, commandline
   parameters or environment variables.
3. Propagate TLS certificate of this proxy to your browsers or
   operating system vault.
4. Access this proxy as local proxy, plain, without any authentication.


## Installation

### Install binaries

There are some prebuilt binaries available on Release pages. Please download
required one for your operating system and CPU architecture.

### Install from sources

To install from sources, please do following:

1. Install Go >= 1.7
2. Download sources
   ```console
   $ git clone https://github.com/scrapinghub/crawlera-headless-proxy
   $ cd crawlera-headless-proxy
   ```
3. Execute make
   ```console
   $ make
   ```

This will build binary `crawlera-headless-proxy`. If you are interesed in
compiling for other OS/CPU architecture, please crosscompile:

```console
$ make crosscompile
```

### Docker container

To download prebuilt container, please do following:

```console
$ docker pull scrapinghub/crawlera-headless-proxy
```

If you want to build this image locally, please do it with make

```console
$ make docker
```

This will build image with tag `crawlera-headless-proxy`. It can
be configured by environment variables or command flags. Default
configuration file path within a container is `/config.toml`.


## Usage

### Help output

```console
$ crawlera-headless-proxy --help
usage: crawlera-headless-proxy [<flags>]

Local proxy for Crawlera to be used with headless browsers.

Flags:
      --help                 Show context-sensitive help (also try --help-long and --help-man).
  -d, --debug                Run in debug mode.
  -b, --bind-ip=BIND-IP      IP to bind to. Default is 127.0.0.1.
  -p, --bind-port=BIND-PORT  Port to bind to. Default is 3128.
  -c, --config=CONFIG        Path to configuration file.
  -l, --tls-ca-certificate=TLS-CA-CERTIFICATE
                             Path to TLS CA certificate file
  -r, --tls-private-key=TLS-PRIVATE-KEY
                             Path to TLS private key
  -a, --api-key=API-KEY      API key to Crawlera.
  -u, --crawlera-host=CRAWLERA-HOST
                             Hostname of Crawlera. Default is proxy.crawlera.com.
  -o, --crawlera-port=CRAWLERA-PORT
                             Port of Crawlera. Default is 8010.
  -v, --dont-verify-crawlera-cert
                             Do not verify Crawlera own certificate
  -x, --xheader=XHEADER ...  Crawlera X-Headers.
      --version              Show application version.
```

### Configuration

Defaults are sensible. If you run this tool without any configuration,
it will start HTTP/HTTPS proxy on `localhost:3128`. The only thing you
usually need to do is to propagate API key.

```console
$ crawlera-headless-proxy -a myapikey
```

This will start local HTTP/HTTPS proxy on `localhost:3128` and will proxy all
requests to `proxy.crawlera.com:8010` with API key `myapikey`.

Also, it is possible to configure this tool using environment variables.
Here is the complete table of configuration options.

| *Description*                                                     | *Environment variable*                | *Comandline parameter*            | *Parameter in configuration file* |
|-------------------------------------------------------------------|---------------------------------------|-----------------------------------|-----------------------------------|
| Run in debug/verbose mode.                                        | `CRAWLERA_HEADLESS_DEBUG`             | `-d`, `--debug`                   | `debug`                           |
| Which IP this tool should listen on (0.0.0.0 for all interfaces). | `CRAWLERA_HEADLESS_BINDIP`            | `-b`, `--bind-ip`                 | `bind_ip`                         |
| Which port this tool should listen.                               | `CRAWLERA_HEADLESS_BINDPORT`          | `-p`, `--bind-port`               | `bind_port`                       |
| Path to the configuration file.                                   | `CRAWLERA_HEADLESS_CONFIG`            | `-c`, `--config`                  | -                                 |
| API key of Crawlera.                                              | `CRAWLERA_HEADLESS_APIKEY`            | `-a`, `--api-key`                 | `api_key`                         |
| Hostname of Crawlera.                                             | `CRAWLERA_HEADLESS_CHOST`             | `-u`, `--crawlera-host`           | `crawlera_host`                   |
| Port of Crawlera.                                                 | `CRAWLERA_HEADLESS_CPORT`             | `-o`, `--crawlera-port`           | `crawlera_port`                   |
| Do not verify Crawlera own TLS certificate.                       | `CRAWLERA_HEADLESS_DONTVERIFY`        | `-k`, `dont-verify-crawlera-cert` | `dont_verify_crawlera_cert`       |
| Path to own TLS CA certificate.                                   | `CRAWLERA_HEADLESS_TLSCACERTPATH`     | `-l`, `tls-ca-certificate`        | `tls_ca_certificate`              |
| Path to own TLS private key.                                      | `CRAWLERA_HEADLESS_TLSPRIVATEKEYPATH` | `-r`, `tls-private-key`           | `tls_private_key`                 |
| Additional Crawlera X-Headers.                                    | `CRAWLERA_HEADLESS_XHEADERS`          | `-x`, `--xheaders`                | Section `xheaders`                |

Configuration is implemented in
[TOML language](https://github.com/toml-lang/toml). If you haven't heard about
TOML, please consider it as a hardened INI configuration files. Every
configuration goes to top level section (unnamed). X-Headers go to its
own section. Let's express following commandline in configuration file:

```console
$ crawlera-headless-proxy -b 0.0.0.0 -p 3129 -u proxy.crawlera.com -o 8010 -x profile=desktop -x cookies=disable
```

Configuration file will look like:

```toml
bind_ip = "0.0.0.0"
bind_port = 3129
crawlera_host = "proxy.crawlera.com"
crawlera_port = 8010

[xheaders]
profile = "desktop"
cookies = "disable"
```

You can use both command line flags, environment variables and
configuration files. This tool will resolve these options according to
this order (1 has max priority, 4 - minimal):

1. Environment variables
2. Commandline flags
3. Configuration file
4. Defaults

## TLS keys

Since crawlera-headless-proxy has to inject X-Headers into responses,
it works with your browser only by HTTP 1.1. Unfortunately, there is no
clear way how to hijack HTTP2 connections. Also, since it is effectively
MITM proxy, you need to use its own TLS certificate. This is hardcoded
into the binary so you have to download it and apply it to your system.
Please consult with manuals of your operating system how to do that.

Link to certificate is
Its SHA256 checksum is `100c7dd015814e7b8df16fc9e8689129682841d50f9a1b5a8a804a1eaf36322d`.

If you want to have your own certificate, please generate it. The
simpliest way to do that is to execute following command:

```console
$ openssl req -x509 -newkey rsa:4096 -keyout private-key.pem -out ca.crt -days 3650 -nodes
```

This command will generate TLS private key `private-key.pem` and
self-signed certificate `ca.crt`.


## Examples

### curl

```console
$ crawlera-headless-proxy -p 3128 -a "$MYAPIKEY" -x profile=desktop
$ curl -x localhost:3128 -sLI https://scrapinghub.com
```

### Selenium (Python)

```python
from selenium import webdriver

CRAWLERA_HEADLESS_PROXY = "localhost:3128"

profile = webdriver.DesiredCapabilities.FIREFOX.copy()
profile["proxy"] = {
    "httpProxy": CRAWLERA_HEADLESS_PROXY,
    "ftpProxy": CRAWLERA_HEADLESS_PROXY,
    "sslProxy": CRAWLERA_HEADLESS_PROXY,
    "noProxy": None,
    "proxyType": "MANUAL",
    "class": "org.openqa.selenium.Proxy",
    "autodetect": False
}

driver = webdriver.Remote("http://localhost:4444/wd/hub", profile)
driver.get("https://scrapinghub.com")
```
