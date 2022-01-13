package layers

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/9seconds/httransform/v2/layers"
	"github.com/pmezard/adblock/adblock"
	log "github.com/sirupsen/logrus"
)

var errAdblockedRequest = errors.New("request was adblocked")

const adblockTimeout = 2 * time.Second

type adblockParsedResult struct {
	rules []*adblock.Rule
	err   error
}

type AdblockLayer struct {
	loaded  bool
	matcher *adblock.RuleMatcher
	cond    *sync.Cond
}

func (a *AdblockLayer) OnRequest(ctx *layers.Context) error {
	host := ctx.RequestHeaders.GetLast("host").Value()
	contentType := ctx.RequestHeaders.GetLast("content-type").Value()
	referer := ctx.RequestHeaders.GetLast("referer").Value()
	adblockRequest := &adblock.Request{
		URL:          string(ctx.Request().URI().FullURI()),
		Domain:       host,
		Timeout:      adblockTimeout,
		ContentType:  contentType,
		OriginDomain: referer,
	}
	logger := getLogger(ctx)

	if !a.loaded {
		a.cond.L.Lock()
		for !a.loaded {
			a.cond.Wait()
		}
		a.cond.L.Unlock()
	}

	matched, _, err := a.matcher.Match(adblockRequest)
	if err != nil {
		logger.WithFields(log.Fields{"err": err}).Debug("Cannot match request.")
	}

	if matched {
		return errAdblockedRequest
	}

	return nil
}

func (a *AdblockLayer) OnResponse(ctx *layers.Context, err error) error {
	if err == errAdblockedRequest {
		getMetrics(ctx).NewAdblockedRequest()
		ctx.Respond("Request was adblocked", http.StatusForbidden)
		logger := getLogger(ctx)
		logger.WithFields(log.Fields{}).Debug("Request was adblocked")

		return nil
	}

	return err
}

func (a *AdblockLayer) sync(lists []string) {
	channel := make(chan *adblockParsedResult, len(lists))
	wg := &sync.WaitGroup{}

	for _, v := range lists {
		wg.Add(1)

		go func(channel chan<- *adblockParsedResult, item string) {
			defer wg.Done()
			a.fetchList(channel, item)
		}(channel, v)
	}

	wg.Wait()
	close(channel)

	a.consumeItems(channel)
}

func (a *AdblockLayer) fetchList(channel chan<- *adblockParsedResult, item string) {
	var reader io.ReadCloser

	var err error

	result := &adblockParsedResult{}

	if strings.HasPrefix(item, "http://") || strings.HasPrefix(item, "https://") {
		reader, err = a.fetchURL(item)
	} else {
		reader, err = a.readFileSystem(item)
	}

	if err != nil {
		result.err = fmt.Errorf("cannot parse rules of item %s", item)
		channel <- result

		return
	}

	defer reader.Close()                  // nolint: errcheck
	defer io.Copy(ioutil.Discard, reader) // nolint: errcheck

	if rules, err := adblock.ParseRules(reader); err != nil {
		result.err = fmt.Errorf("cannot parse rules of item %s", item)
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

func (a *AdblockLayer) fetchURL(url string) (io.ReadCloser, error) {
	log.WithFields(log.Fields{"url": url}).Debug("Fetch adblock list")

	resp, err := http.Get(url) // nolint: gosec, bodyclose

	status := ""
	if resp != nil {
		status = resp.Status
	}

	log.WithFields(log.Fields{
		"url":    url,
		"err":    err,
		"status": status,
	}).Debug("Got adblock list response")

	if err != nil {
		return nil, fmt.Errorf("cannot fetch url %s: %w", url, err)
	}

	return resp.Body, nil
}

func (a *AdblockLayer) readFileSystem(path string) (io.ReadCloser, error) {
	log.WithFields(log.Fields{"path": path}).Debug("Open filesystem adblock list")

	fp, err := os.Open(path) // nolint: gosec

	log.WithFields(log.Fields{
		"path": path,
		"err":  err,
	}).Debug("Opened filesystem adblock list")

	if err != nil {
		return nil, fmt.Errorf("cannot open file %s: %w", path, err)
	}

	return fp, nil
}

func (a *AdblockLayer) consumeItems(channel <-chan *adblockParsedResult) {
	rules := []*adblock.Rule{}

	for item := range channel {
		if item.err != nil {
			log.Fatal(item.err.Error())
		}

		rules = append(rules, item.rules...)
	}

	a.cond.L.Lock()
	defer a.cond.L.Unlock()

	for idx, value := range rules {
		if err := a.matcher.AddRule(value, idx); err != nil {
			log.Warnf("Cannot add rule '%s': %s", value.Raw, err.Error())
		}
	}

	a.loaded = true
	a.cond.Broadcast()
}

func NewAdblockLayer(lists []string) layers.Layer {
	layer := &AdblockLayer{
		cond:    sync.NewCond(&sync.Mutex{}),
		matcher: adblock.NewMatcher(),
	}
	go layer.sync(lists)

	return layer
}
