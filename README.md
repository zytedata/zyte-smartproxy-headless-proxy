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

* Install Go >= 1.7
* Download sources

```console
$ git clone https://github.com/scrapinghub/crawlera-headless-proxy
$ cd crawlera-headless-proxy
```

* Execute make

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
  -m, --proxy-api-ip=PROXY-API-IP
                             IP to bind proxy API to. Default is the bind-ip value.
  -p, --bind-port=BIND-PORT  Port to bind to. Default is 3128.
  -w, --proxy-api-port=PROXY-API-PORT
                             Port to bind proxy api to. Default is 3130.
  -c, --config=CONFIG        Path to configuration file.
  -l, --tls-ca-certificate=TLS-CA-CERTIFICATE
                             Path to TLS CA certificate file.
  -r, --tls-private-key=TLS-PRIVATE-KEY
                             Path to TLS private key.
  -t, --no-auto-sessions     Disable automatic session management.
  -n, --concurrent-connections=CONCURRENT-CONNECTIONS
                             Number of concurrent connections.
  -a, --api-key=API-KEY      API key to Crawlera.
  -u, --crawlera-host=CRAWLERA-HOST
                             Hostname of Crawlera. Default is proxy.crawlera.com.
  -o, --crawlera-port=CRAWLERA-PORT
                             Port of Crawlera. Default is 8010.
  -v, --dont-verify-crawlera-cert
                             Do not verify Crawlera own certificate.
  -x, --xheader=XHEADER ...  Crawlera X-Headers.
  -k, --adblock-list=ADBLOCK-LIST ...
                             A list to requests to filter out (ADBlock compatible).
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
| Disable automatic session management                              | `CRAWLERA_HEADLESS_NOAUTOSESSIONS`    | `-t`, `--no-auto-sessions`        | `no_auto_sessions`                |
| Maximal ammount of concurrent connections to process              | `CRAWLERA_HEADLESS_CONCURRENCY`       | `-n`, `--concurrent-connections`  | `concurrent_connections`          |
| Additional Crawlera X-Headers.                                    | `CRAWLERA_HEADLESS_XHEADERS`          | `-x`, `--xheaders`                | Section `xheaders`                |
| Adblock-compatible filter lists.                                  | `CRAWLERA_HEADLESS_ADBLOCKLISTS`      | `-k`, `--adblock-list`            | `adblock_lists`                   |
| Which IP should proxy API listen on (default is `bind-ip` value). | `CRAWLERA_HEADLESS_PROXYAPIIP`        | `-m`, `--proxy-api-ip`            | `proxy_api_ip`                    |
| Which port proxy API should listen on.                            | `CRAWLERA_HEADLESS_PROXYAPIPORT`      | `-w`, `--proxy-api-port`          | `proxy_api_port`                  |

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

## Concurrency

There is a limiter on maxmial amount of concurrent connections
`--concurrent-connections`. This is required because default Crawlera
limits amount of concurrent connections based on billing plan of the
user. If user exceeds this amount, Crawlera returns response with status
code 429. This can be rather irritating so there is internal limiter
which is more friendly to the browsers. You need to setup an amount of
concurrent connections for your plan and crawlera-headless-proxy will
throttle your requests _before_ they will go to Crawlera. It won't send
429 back, it just holds excess requests.

## Automatic session management

Crawlera allows to use sessions and sessions are natural if we are talking
about browsers. Session binds a certain IP to some session ID so all requests
will go through the same IP, in the same way as ordinary work with browser
looks like. It can slow down your crawl but increase its quality for some
websites.

Current implementation of automatic session management is done with
assumption that only one browser is used to access this proxy. There is
no clear and simple way how to distinguish the browsers accessing this
proxy concurrently.

Basic behavior is here:

1. If session is not created, it would be created on the first request.
2. Until session is known, all other requests are on hold
3. After session id is known, other requests will start to use that session.
4. If session became broken, all requests are set on hold until new session
   will be created.
5. All requests which were failed because of broken session would be
   retried with new. If new session is not ready yet, they will wait until
   this moment.

Such retries will be done only once because they might potentially block
browser for the long time. All retries are also done with 30 seconds
timeout.

## Adblock list support

crawlera-headless-proxy supports preventive filtering by;
adblock-compatible filter lists like EasyList. If you start the tool
with such a lists, they are going to be downloaded and requests to
trackers/advertisment platforms will be filtered. This will save you a;
lot of throughput and requests passed to Crawlera.

If you do not pass any list, such filtering won't
be enabled. The list we recommend to use are
[EasyList](https://easylist.to/easylist/easylist.txt)
(please do not forget to add region-specific lists),
[EasyPrivacy](https://easylist.to/easylist/easyprivacy.txt) and
[Disconnect](https://s3.amazonaws.com/lists.disconnect.me/simple_malware.txt).

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

## Proxy API

crawlera-headless-proxy has its own HTTP Rest API which is bind to
another port. Right now only one endpoint is supported.

### `GET /stats`

This endpoint returns various statistics on current work of proxy.

Example:

```json
{
  "requests_number": 342,
  "crawlera_requests": 343,
  "sessions_created": 2,
  "clients_connected": 1,
  "clients_serving": 1,
  "traffic": 5144097,
  "overall_times": {
    "average": 1.1697893100117303,
    "minimal": 0.019068425,
    "maxmimal": 13.083649925,
    "median": 0.399386525,
    "standard_deviation": 2.7668849085566722,
    "percentiles": {
      "10": 0.197817644,
      "20": 0.232745909,
      "30": 0.276867836,
      "40": 0.329860919,
      "50": 0.399386525,
      "60": 0.467046535,
      "70": 0.576318183,
      "75": 0.603624489,
      "80": 0.633094287,
      "85": 0.699524249,
      "90": 0.839260954,
      "95": 10.736666247,
      "99": 11.715572608
    }
  },
  "crawlera_times": {
    "average": 1.1055083281666656,
    "minimal": 0.000111273,
    "maxmimal": 11.7612128,
    "median": 0.36406424049999997,
    "standard_deviation": 2.691387695152139,
    "percentiles": {
      "10": 0.186066392,
      "20": 0.219907051,
      "30": 0.258462571,
      "40": 0.301227936,
      "50": 0.36406424049999997,
      "60": 0.423234904,
      "70": 0.527960255,
      "75": 0.559111637,
      "80": 0.601874697,
      "85": 0.658456279,
      "90": 0.770335886,
      "95": 10.723328013,
      "99": 11.627872067
    }
  },
  "traffic_times": {
    "average": 15129.697058823529,
    "minimal": 8,
    "maxmimal": 515722,
    "median": 9700.5,
    "standard_deviation": 34750.735056119,
    "percentiles": {
      "10": 6588,
      "20": 7822,
      "30": 8593,
      "40": 9271.5,
      "50": 9700.5,
      "60": 10589,
      "70": 11186,
      "75": 11562,
      "80": 11830,
      "85": 12415.5,
      "90": 14410,
      "95": 46591,
      "99": 92629
    }
  },
  "uptime": 297
}
```

Here is the description of these stats:

* `requests_number` - a number of requests managed by headless
  proxy. This includes all possible requests, not only those which were
  send to Crawlera.
* `crawlera_requests` - a number of requests which were send to
  Crawlera. This also includes retries on session restoration etc.
* `sessions_created` - how many sessions were created by headless
  proxy so far.
* `clients_connected` - how many clients (requests) are connected to
  headless proxy at this moment.
* `clients_serving` - how many clients (requests) are doing requests to
  Crawlera now.
* `traffic` - an amount of traffic sent to clients in bytes. This metric
  does not includes size of headers now, only response bodies.

`*_times` describes different time series (overall response time, time
spent in crawlera) etc and provide average(mean), min and max values,
stddev and histogram of percentiles. Time series are done in window
mode, tracking only latest 3000 values.

Please pay attention that usually `requests_number` and
`crawlera_requests` are different. This is because headless proxy
filters adblock requests and also retries to recreate sessions which
implies additional Crawlera requests. So, depending on the netloc
proportion of these numbers can differ.

Also, `clients_serving <= clients_connected` because of rate limiting.
You may consider `client_serving` as requests which pass rate limiter.


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
