package proxy

import (
	"net/http"
	"sync"
	"time"

	"github.com/elazarl/goproxy"
	log "github.com/sirupsen/logrus"

	"github.com/9seconds/crawlera-headless-proxy/config"
)

var (
	sessionCleanAfter     time.Duration
	sessionClientDuration time.Duration
	sessions              sync.Map
)

type session struct {
	lastUsed time.Time
	mutex    sync.Mutex
	id       string
}

func newSession() *session {
	return &session{
		lastUsed: time.Now(),
		id:       "",
	}
}

func applyAutoSessions(proxy *goproxy.ProxyHttpServer, conf *config.Config) {
	go func() {
		for range time.Tick(sessionCleanAfter) {
			sessions.Range(func(key, value interface{}) bool {
				sess := value.(*session)
				if time.Since(sess.lastUsed) >= sessionCleanAfter {
					sessions.Delete(key)
				}
				return true
			})
		}
	}()

	proxy.OnRequest().DoFunc(
		func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			key := req.URL.Hostname()
			sessionRaw, _ := sessions.LoadOrStore(key, newSession())
			sess := sessionRaw.(*session)

			// it is possible that 2 goroutines have called that
			// simultaneously. 2nd parameter won't help there unfortunately.
			if sess.id == "" {
				sess.mutex.Lock()
				if sess.id != "" {
					sess.mutex.Unlock()
				}
			}

			if sess.id == "" {
				log.WithFields(log.Fields{
					"hostname": key,
				}).Info("Initialize new session.")
				req.Header.Set("X-Crawlera-Session", "create")
			} else {
				req.Header.Set("X-Crawlera-Session", sess.id)
			}

			return req, nil
		})

	proxy.OnResponse().DoFunc(
		func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
			key := resp.Request.URL.Hostname()
			sessionRaw, ok := sessions.Load(key)
			if !ok {
				log.WithFields(log.Fields{
					"session-id": resp.Request.Header.Get("X-Crawlera-Session"),
					"hostname":   key,
				}).Warn("Expired session.")
				return resp
			}
			sess := sessionRaw.(*session)

			if resp.Header.Get("X-Crawlera-Error") == "" { // clean response
				sess.lastUsed = time.Now()
				if sess.id == "" {
					sess.id = resp.Header.Get("X-Crawlera-Session")
					sess.mutex.Unlock()
				}
				return resp
			}

			sess.mutex.Lock()
			defer sess.mutex.Unlock()

			resp, ok = tryToRestoreSession(resp)
			if !ok {
				sess.id = ""
				log.WithFields(log.Fields{
					"hostname": key,
				}).Warn("Failed to recreate session.")
			} else {
				sess.id = resp.Header.Get("X-Crawlera-Session")
				sess.lastUsed = time.Now()
			}

			return resp
		})
}

func tryToRestoreSession(resp *http.Response) (*http.Response, bool) {
	key := resp.Request.URL.Hostname()
	client := &http.Client{Timeout: sessionClientDuration}
	request := resp.Request
	request.Header.Set("X-Crawlera-Session", "create")

	if newResp, err := client.Do(request); err != nil {
		log.WithFields(log.Fields{
			"hostname": key,
			"error":    err,
			"url":      request.URL.String(),
		}).Debug("Failed to retry request.")
		return resp, false
	}
	if newResp.Header.Get("X-Crawlera-Error") == "" {
		return newResp, true
	}

	log.WithFields(log.Fields{
		"hostname": key,
		"error":    newResp.Header.Get("X-Crawlera-Error"),
		"url":      request.URL.String(),
	}).Debug("Failed to retry request.")

	return newResp, false
}

func init() {
	sessionCleanAfter, _ = time.ParseDuration("10m")
	sessionClientDuration, _ = time.ParseDuration("180s")
}
