package middleware

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/elazarl/goproxy"
	"github.com/juju/errors"
	"github.com/pmezard/adblock/adblock"
	log "github.com/sirupsen/logrus"

	"github.com/9seconds/crawlera-headless-proxy/config"
)

const adblockTimeout = 2 * time.Second

type adblockMiddleware struct {
	UniqBase

	loaded  bool
	rules   []*adblock.Rule
	matcher *adblock.RuleMatcher
	cond    *sync.Cond
}

type adblockParsedResult struct {
	rules []*adblock.Rule
	err   error
}

func (ab *adblockMiddleware) OnRequest() ReqType {
	return ab.BaseOnRequest(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		var newResponse *http.Response

		host := req.URL.Hostname()
		if host == "" {
			host = req.Host
		}
		adblockRequest := &adblock.Request{
			URL:     req.URL.String(),
			Domain:  host,
			Timeout: adblockTimeout,
		}

		if !ab.loaded {
			ab.cond.L.Lock()
			for !ab.loaded {
				ab.cond.Wait()
			}
			ab.cond.L.Unlock()
		}

		matched, id, err := ab.matcher.Match(adblockRequest)
		if err != nil {
			log.WithFields(log.Fields{
				"url": req.URL.String(),
				"err": err,
			}).Debug("Cannot match request.")
		}
		if matched {
			newResponse = goproxy.NewResponse(req,
				goproxy.ContentTypeText,
				http.StatusForbidden,
				fmt.Sprintf("Adblocked by rule '%s'", ab.rules[id].Raw),
			)
		}

		return req, newResponse
	})
}

func (ab *adblockMiddleware) OnResponse() RespType {
	return ab.BaseOnResponse(func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		return resp
	})
}

func (ab *adblockMiddleware) consumeItems(channel <-chan *adblockParsedResult) {
	for item := range channel {
		if item.err != nil {
			log.Fatal(item.err.Error())
		}
		ab.rules = append(ab.rules, item.rules...)
	}

	ab.cond.L.Lock()
	defer ab.cond.L.Unlock()

	for idx, value := range ab.rules {
		if err := ab.matcher.AddRule(value, idx); err != nil {
			log.Warnf("Cannot add rule '%s': %s", value.Raw, err.Error())
		}
	}

	ab.loaded = true
	ab.cond.Broadcast()
}

func (ab *adblockMiddleware) fetchList(channel chan<- *adblockParsedResult, item string) {
	var reader io.ReadCloser
	var err error
	result := &adblockParsedResult{}

	if strings.HasPrefix(item, "http://") || strings.HasPrefix(item, "https://") {
		reader, err = ab.fetchURL(item)
	} else {
		reader, err = ab.readFileSystem(item)
	}

	if err != nil {
		result.err = errors.Annotatef(err, "Cannot parse rules of item %s", item)
		channel <- result
		return
	}
	defer reader.Close()                  // nolint: errcheck
	defer io.Copy(ioutil.Discard, reader) // nolint: errcheck

	if rules, err := adblock.ParseRules(reader); err != nil {
		result.err = errors.Annotatef(err, "Cannot parse rules of item %s", item)
	} else {
		for _, rule := range rules {
			if !rule.HasUnsupportedOpts() {
				result.rules = append(result.rules, rule)
			} else {
				log.WithFields(log.Fields{
					"rule": rule.Raw,
				}).Debug("Skip unsupported adblock rule")
			}
		}
	}

	channel <- result
}

func (ab *adblockMiddleware) fetchURL(url string) (io.ReadCloser, error) {
	log.WithFields(log.Fields{"url": url}).Debug("Fetch adblock list")
	resp, err := http.Get(url)
	log.WithFields(log.Fields{
		"url":    url,
		"err":    err,
		"status": resp.Status,
	}).Debug("Got adblock list response")

	if err != nil {
		return nil, errors.Annotatef(err, "Cannot fetch URL %s", url)
	}

	return resp.Body, nil
}

func (ab *adblockMiddleware) readFileSystem(path string) (io.ReadCloser, error) {
	log.WithFields(log.Fields{"path": path}).Debug("Open filesystem adblock list")
	fp, err := os.Open(path)
	log.WithFields(log.Fields{
		"path": path,
		"err":  err,
	}).Debug("Opened filesystem adblock list")

	if err != nil {
		return nil, errors.Annotatef(err, "Cannot open file %s", path)
	}

	return fp, nil
}

// NewAdblockMiddleware returns a middleware which uses adblock lists to
// filter out some requests which should not be performed BEFORE they
// will go to Crawlera.
func NewAdblockMiddleware(conf *config.Config, proxy *goproxy.ProxyHttpServer) Middleware {
	ware := &adblockMiddleware{}
	ware.mtype = middlewareTypeAdblock

	ware.matcher = adblock.NewMatcher()
	ware.cond = sync.NewCond(&sync.Mutex{})

	go func() {
		channel := make(chan *adblockParsedResult, len(conf.AdblockLists))
		wg := &sync.WaitGroup{}

		for _, v := range conf.AdblockLists {
			wg.Add(1)
			go func(channel chan<- *adblockParsedResult, item string) {
				defer wg.Done()
				ware.fetchList(channel, item)
			}(channel, v)
		}

		wg.Wait()
		close(channel)

		ware.consumeItems(channel)
	}()

	return ware
}
