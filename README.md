# Crawlera Headless Proxy

[![Build Status](https://travis-ci.org/scrapinghub/crawlera-headless-proxy.svg?branch=master)](https://travis-ci.org/scrapinghub/crawlera-headless-proxy)
[![Go Report Card](https://goreportcard.com/badge/github.com/scrapinghub/crawlera-headless-proxy)](https://goreportcard.com/report/github.com/scrapinghub/crawlera-headless-proxy)

Crawlera Headless proxy is a proxy which main intent
is to help users with headless browsers to use
[Crawlera](https://scrapinghub.com/crawlera). This
includes different implementations of headless browsers
such as [Splash](https://scrapinghub.com/splash),
headless [Chrome](https://google.com/chrome/), and
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
   requires to have a minimal possible set of headers. For example, it is
   recommended to remove `Accept` header by default. It is rather hard
   to do that using headless browsers API.
4. Crawlera works best with browsers only with some preconditions which
   users have to repeat everytime: session usage, some recommended headers
   like `Referer` etc.

Crawlera Headless Proxy intended to help users to avoid such
problems. You should generally think about it as a proxy which should
be accessible by your headless browser of Selenium grid. This proxy
propagates your requests to Crawlera maintaining API key and injecting
headers into the requests. Basically, you have to do a bare minimum:

1. Get Crawlera API key
2. Run this proxy on your local machine or any machine accessible by
   headless browser, configuring it with a configuration file, command line
   parameters or environment variables.
3. Propagate TLS certificate of this proxy to your browsers or
   operating system vault.
4. Access this proxy as a local proxy, plain, without any authentication.


## Installation

Check out the [Using Headless Browsers with Zyte Smart Proxy Manager](https://docs.zyte.com/smart-proxy-manager/headless.html) manual.

### Install binaries

There are some prebuilt binaries available on Release pages. Please download
required one for your operating system and CPU architecture.

### Install from sources

#### Install prerequisites

You need to have a distributions of Go programming language, git, bash
and make installed. We use Go >= 1.11 so please be sure that you have
fresh enough version.

To install them on Ubuntu/Debian, please execute the following command:

```console
$ sudo apt install -y bash make git golang-go
```

If you have Ubuntu older than 18.10, please install go with snap package:

```console
$ sudo snap install --classic go
```

To install them on OS X with [Homebrew](https://brew.sh/),
please execute the following command:

```console
$ brew install go make git
```


#### Install from HomeBrew

If you use [HomeBrew](https://brew.sh), you can use it to install headless
proxy:

```console
$ curl -L https://raw.githubusercontent.com/zytedata/zyte-smartproxy-headless-proxy/master/crawlera-headless-proxy.rb > crawlera-headless-proxy.rb && brew install --HEAD crawlera-headless-proxy.rb
```


#### Build binary

```console
$ git clone https://github.com/zytedata/zyte-smartproxy-headless-proxy.git
$ cd zyte-smartproxy-headless-proxy
```

The next step is to execute make

```console
$ make
```

This will build binary `crawlera-headless-proxy`. If you are interested
in compiling for other OS/CPU architecture, please cross-compile:

```console
$ make crosscompile
```

You'll find a set of compiled biaries in `./ccbuilds` directory after
the process is finished.


### Docker container

To download prebuilt container, please do the following:

```console
$ docker pull zytedata/zyte-smartproxy-headless-proxy
```

If you want to build this image locally, please do it with make (also,
be sure that [docker is installed](https://docs.docker.com/install/)).

```console
$ make docker
```

This will build an image with tag `crawlera-headless-proxy`. It can
be configured by environment variables or command flags. Default
configuration file path within a container is `/config.toml`.

If you want to have a smaller image (but build time will grow a lot),
you can build it with `docker-slim` make target.

```console
$ make docker-slim
```


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
  -z, --direct-access-hostpath-regexps=DIRECT-ACCESS-HOSTPATH-REGEXPS ...
                             A list of regexps for hostpath for direct access, bypassing Crawlera.
      --version              Show application version.
```

Docker example:
```console
$ docker run --name crawlera-headless-proxy -p 3128:3128 zytedata/zyte-smartproxy-headless-proxy --help
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
Here is the complete table of configuration options and corresponding
environment variables.

| *Description*                                                                    | *Environment variable*                | *Comandline parameter*                   | *Parameter in configuration file* | *Default value*      |
|----------------------------------------------------------------------------------|---------------------------------------|------------------------------------------|-----------------------------------|----------------------|
| Run in debug/verbose mode.                                                       | `CRAWLERA_HEADLESS_DEBUG`             | `-d`, `--debug`                          | `debug`                           | `false`              |
| Which IP this tool should listen on (0.0.0.0 for all interfaces).                | `CRAWLERA_HEADLESS_BINDIP`            | `-b`, `--bind-ip`                        | `bind_ip`                         | `127.0.0.1`          |
| Which port this tool should listen.                                              | `CRAWLERA_HEADLESS_BINDPORT`          | `-p`, `--bind-port`                      | `bind_port`                       | 3128                 |
| Path to the configuration file.                                                  | `CRAWLERA_HEADLESS_CONFIG`            | `-c`, `--config`                         | -                                 |                      |
| API key of Crawlera.                                                             | `CRAWLERA_HEADLESS_APIKEY`            | `-a`, `--api-key`                        | `api_key`                         |                      |
| Hostname of Crawlera.                                                            | `CRAWLERA_HEADLESS_CHOST`             | `-u`, `--crawlera-host`                  | `crawlera_host`                   | `proxy.crawlera.com` |
| Port of Crawlera.                                                                | `CRAWLERA_HEADLESS_CPORT`             | `-o`, `--crawlera-port`                  | `crawlera_port`                   | 8010                 |
| Do not verify Crawlera own TLS certificate.                                      | `CRAWLERA_HEADLESS_DONTVERIFY`        | `-v`, `--dont-verify-crawlera-cert`      | `dont_verify_crawlera_cert`       | `false`              |
| Path to own TLS CA certificate.                                                  | `CRAWLERA_HEADLESS_TLSCACERTPATH`     | `-l`, `--tls-ca-certificate`             | `tls_ca_certificate`              | <embeded>            |
| Path to own TLS private key.                                                     | `CRAWLERA_HEADLESS_TLSPRIVATEKEYPATH` | `-r`, `--tls-private-key`                | `tls_private_key`                 | <embeded>            |
| Disable automatic session management                                             | `CRAWLERA_HEADLESS_NOAUTOSESSIONS`    | `-t`, `--no-auto-sessions`               | `no_auto_sessions`                | `false`              |
| Maximal ammount of concurrent connections to process                             | `CRAWLERA_HEADLESS_CONCURRENCY`       | `-n`, `--concurrent-connections`         | `concurrent_connections`          | 0                    |
| Additional Crawlera X-Headers.                                                   | `CRAWLERA_HEADLESS_XHEADERS`          | `-x`, `--xheaders`                       | Section `xheaders`                |                      |
| Adblock-compatible filter lists.                                                 | `CRAWLERA_HEADLESS_ADBLOCKLISTS`      | `-k`, `--adblock-list`                   | `adblock_lists`                   |                      |
| Regular expressions for hostpath URL part for direct access, bypassing Crawlera. | `CRAWLERA_HEADLESS_DIRECTACCESS`      | `-z`, `--direct-access-hostpath-regexps` | `direct_access_hostpath_regexps`  |                      |
| Which IP should proxy API listen on (default is `bind-ip` value).                | `CRAWLERA_HEADLESS_PROXYAPIIP`        | `-m`, `--proxy-api-ip`                   | `proxy_api_ip`                    | <same as `bind_ip`>  |
| Which port proxy API should listen on.                                           | `CRAWLERA_HEADLESS_PROXYAPIPORT`      | `-w`, `--proxy-api-port`                 | `proxy_api_port`                  | 3130                 |

0 concurrent connections means unlimited. Embedded TLS key/certificate
means that headless proxy will use ones from the repository.

Configuration is implemented in [TOML
language](https://github.com/toml-lang/toml). If you haven't heard about
TOML, please consider it as a hardened INI configuration file. Every
configuration goes to top-level section (unnamed). X-Headers go to its
own section. Let's express following command line in the configuration
file:

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

You can use both command line flags, environment variables, and
configuration files. This tool will resolve these options according to
this order (1 has max priority, 4 - minimal):

1. Environment variables
2. Commandline flags
3. Configuration file
4. Defaults

Docker example:
```console
$ docker run --name crawlera-headless-proxy -p 3128:3128 zytedata/zyte-smartproxy-headless-proxy -a $APIKEY -d -x profile=pass -x cookies=disable -x no-bancheck=1 --direct-access-hostpath-regexps=".*?\.(?:txt|json|css|less|js|mjs|cjs|gif|ico|jpe?g|svg|png|webp|mkv|mp4|mpe?g|webm|eot|ttf|woff2?)$" --adblock-list="https://easylist.to/easylist/easylist.txt" --adblock-list="https://easylist.to/easylist/easyprivacy.txt"
```


## Concurrency

There is a limiter on maximal amount of concurrent connections
`--concurrent-connections`. This is required because default Crawlera
limits the number of concurrent connections based on the billing
plan of the user. If the user exceeds this amount, Crawlera returns
a response with status code 429. This can be rather irritating so
there is internal limiter which is more friendly to the browsers. You
need to set up a number of concurrent connections for your plan and
crawlera-headless-proxy will throttle your requests before they will go
to Crawlera. It won't send 429 back, it just holds excess requests.


## Automatic session management

Crawlera allows using sessions and sessions are natural if we are
talking about browsers. Session binds a certain IP to some session ID so
all requests will go through the same IP, in the same way as ordinary
work with browser looks like. It can slow down your crawl but increase
its quality for some websites.

The current implementation of automatic session management is done with
the assumption that only one browser is used to access this proxy. There
is no clear and simple way how to distinguish the browsers accessing
this proxy concurrently.

Basic behavior is here:

1. If the session is not created, it would be created on the first request.
2. Until session is known, all other requests are on hold
3. After session id is known, other requests will start to use that session.
4. If the session became broken, all requests are set on hold until the
   new session will be created.
5. All requests which were failed because of a broken session would
   be retried with new. If a new session is not ready yet, they will
   wait until this moment.

Such retries will be done only once because they might potentially block
browser for a long time. All retries are also done with 30 seconds
timeout.


## Adblock list support

crawlera-headless-proxy supports preventive filtering by;
adblock-compatible filter lists like EasyList. If you start the tool
with such lists, they are going to be downloaded and requests to
trackers/advertising platforms will be filtered. This will save you a
lot of throughput and requests passed to Crawlera.

If you do not pass any list, such filtering won't
be enabled. The list we recommend to use is
[EasyList](https://easylist.to/easylist/easylist.txt)
(please do not forget to add region-specific lists),
[EasyPrivacy](https://easylist.to/easylist/easyprivacy.txt) and
[Disconnect](https://s3.amazonaws.com/lists.disconnect.me/simple_malware.txt).


## Direct access

Sometimes you want to save a capacity and execute requests bypassing
Crawlera. These requests can include some static assets, text files or
anything else where Crawlera is not necessary.

You can specify a list of regular expressions which matches host + path
parts of URL for direct access from headless proxy, ignoring Crawlera.


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
simplest way to do that is to execute the following command:

```console
$ openssl req -x509 -newkey rsa:4096 -keyout private-key.pem -out ca.crt -days 3650 -nodes
```

This command will generate TLS private key `private-key.pem` and
self-signed certificate `ca.crt`.


## Proxy API

crawlera-headless-proxy has its own HTTP Rest API which is bind to
another port. Right now only one endpoint is supported.

### `GET /stats`

This endpoint returns various statistics on the current work of proxy.

Example:

```json
{
  "requests_number": 423,
  "crawlera_requests": 426,
  "crawlera_errors": 0,
  "all_errors": 6,
  "adblocked_requests": 0,
  "sessions_created": 4,
  "clients_connected": 1,
  "clients_serving": 1,
  "traffic": 6326557,
  "overall_times": {
    "average": 0.37859728122037895,
    "minimal": 0.016320158,
    "maxmimal": 6.96558913,
    "median": 0.1117137805,
    "standard_deviation": 1.001460285777158,
    "percentiles": {
      "10": 0.05237131,
      "20": 0.071472272,
      "30": 0.088965026,
      "40": 0.101607119,
      "50": 0.1117137805,
      "60": 0.125672599,
      "70": 0.137716451,
      "75": 0.146478028,
      "80": 0.154273865,
      "85": 0.162262952,
      "90": 0.180582867,
      "95": 3.514414853,
      "99": 3.729193071
    }
  },
  "crawlera_times": {
    "average": 0.30196985308000035,
    "minimal": 3.1394e-05,
    "maxmimal": 3.750836014,
    "median": 0.080731409,
    "standard_deviation": 0.8410949224993787,
    "percentiles": {
      "10": 0.036049358,
      "20": 0.049492537,
      "30": 0.062905696,
      "40": 0.072465399,
      "50": 0.080731409,
      "60": 0.0885455695,
      "70": 0.09895164,
      "75": 0.103160955,
      "80": 0.110743335,
      "85": 0.118884673,
      "90": 0.129430856,
      "95": 3.494953838,
      "99": 3.694614379
    }
  },
  "traffic_times": {
    "average": 15099.18138424821,
    "minimal": 336,
    "maxmimal": 516239,
    "median": 10383,
    "standard_deviation": 31541.145341694657,
    "percentiles": {
      "10": 7511,
      "20": 8441,
      "30": 9230,
      "40": 9833,
      "50": 10383,
      "60": 10889,
      "70": 11398,
      "75": 11853,
      "80": 12327,
      "85": 13121,
      "90": 15153,
      "95": 44791,
      "99": 73846
    }
  },
  "uptime": 123
}
```

Here is the description of these stats:

* `requests_number` - a number of requests managed by headless proxy.
     This includes all possible requests, not only those which were
     sent to Crawlera.
* `crawlera_requests` - a number of requests which were sent to Crawlera.
     This also includes retries on session restoration etc.
* `sessions_created` - how many sessions were created by headless
     proxy so far.
* `clients_connected` - how many clients (requests) are connected to
     the headless proxy at this moment.
* `clients_serving` - how many clients (requests) are doing requests
     to Crawlera now.
* `traffic` - an amount of traffic sent to clients in bytes.
     This metric does include headers and body sizes.
* `crawlera_errors` - a number of responses where `X-Crawlera-Error`
     header is set.
* `all_errors` - a number of responses with errors (canceled,
     timeouts and crawlera_errors).
* `adblocked_requests` - a number of requests which were
     blocked by Adblock lists.
*_`times` describes different time series (overall response time,
     time spent in crawlera) etc and provide average(mean), min and
     max values, standard deviation and histogram of percentiles.
     Time series are done in window mode, tracking only latest 3000 values.

Please pay attention that usually requests_number and crawlera_requests
are different. This is because headless proxy filters adblock requests
and also retries to recreate sessions which imply additional Crawlera
requests. So, depending on the netloc proportion of these numbers can
differ.

Also, `clients_serving <= clients_connected` because of rate limiting. You
may consider client_serving as requests which pass rate limiter.


## Crawlera X-Headers

Crawlera is configured using the special headers, which usually are
called x-headers (they have `X-` prefix in their name). You can find a
[full list](https://doc.scrapinghub.com/crawlera.html#request-headers)
of them in documentation.

There are 2 different ways of providing these header to the headless proxy:

1. Use the full name
2. Use the short version

For example, user wants to use desktop browser profile. Corresponding
header is `X-Crawlera-Profile`. So, user can add following line to the
configuration:

```toml
[xheaders]
x-crawlera-profile = "desktop"
```

or pass it via command line:

```console
$ crawlera-headless-proxy ... -x x-crawlera-profile=desktop
```

but it makes no sense to use `X-Crawlera-` prefix all the time, please omit it.

```toml
[xheaders]
profile = "desktop"
```

or pass it via command line:

```console
$ crawlera-headless-proxy ... -x profile=desktop
```


# Examples

## curl

```console
$ crawlera-headless-proxy -p 3128 -a "$MYAPIKEY" -x profile=desktop
$ curl -x localhost:3128 -sLI https://scrapinghub.com
```

## Selenium (Python)

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
