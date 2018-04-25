package proxy

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

var (
	adblockRules   []*adblock.Rule
	adblockMatcher *adblock.RuleMatcher
)

const adblockTimeout = 2 * time.Second

type adblockParsedResult struct {
	rules []*adblock.Rule
	err   error
}

func installAdblocker(list []string) {
	channel := make(chan *adblockParsedResult, len(list))
	wg := &sync.WaitGroup{}

	for _, v := range list {
		wg.Add(1)
		go func(channel chan<- *adblockParsedResult, item string) {
			defer wg.Done()

			var reader io.ReadCloser
			var err error
			result := &adblockParsedResult{}
			if strings.HasPrefix(item, "http://") || strings.HasPrefix(item, "https://") {
				reader, err = adblockFetchURL(item)
			} else {
				reader, err = adblockReadFileSystem(item)
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
				result.rules = rules
			}
			channel <- result
		}(channel, v)
	}
	wg.Wait()
	close(channel)

	for item := range channel {
		if item.err != nil {
			log.Fatalf(item.err.Error())
		}
		adblockRules = append(adblockRules, item.rules...)
	}
	for idx, value := range adblockRules {
		if err := adblockMatcher.AddRule(value, idx); err != nil {
			log.Infof("Cannot add rule '%s': %s", value.Raw, err.Error())
		}
	}
}

func handlerAdblockReq(proxy *goproxy.ProxyHttpServer, conf *config.Config) handlerTypeReq {
	return func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		var newResponse *http.Response

		host := req.URL.Host
		if host == "" {
			host = req.Host
		}
		adblockRequest := &adblock.Request{
			URL:     req.URL.String(),
			Domain:  host,
			Timeout: adblockTimeout,
		}
		matched, id, err := adblockMatcher.Match(adblockRequest)
		if err != nil {
			log.WithFields(log.Fields{
				"url": req.URL.String(),
				"err": err,
			}).Debug("Cannot match request.")
		}
		if matched {
			newResponse = goproxy.NewResponse(req,
				goproxy.ContentTypeText,
				http.StatusNotFound,
				fmt.Sprintf("Adblocked by rule '%s'", adblockRules[id].Raw),
			)
		}

		return req, newResponse
	}
}

func adblockFetchURL(url string) (io.ReadCloser, error) {
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

func adblockReadFileSystem(path string) (io.ReadCloser, error) {
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

func init() {
	adblockMatcher = adblock.NewMatcher()
}
