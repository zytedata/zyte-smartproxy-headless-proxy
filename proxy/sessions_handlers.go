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
	sessionHTTPClient *http.Client

	sessionID      string
	sessionCreator string
	sessionCond    *sync.Cond
)

const sessionClientTimeout = 30 * time.Second

func installHTTPClient(transport http.RoundTripper) {
	sessionHTTPClient = &http.Client{
		Timeout:   sessionClientTimeout,
		Transport: transport,
	}
}

func handlerSessionReq(proxy *goproxy.ProxyHttpServer, conf *config.Config) handlerTypeReq {
	return func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		sessionState := getState(ctx)

		if sessionID == "" && sessionCreator == "" {
			sessionCond.L.Lock()
			if sessionID == "" && sessionCreator == "" {
				sessionID = ""
				sessionCreator = sessionState.id
				sessionCond.L.Unlock()
				req.Header.Set("X-Crawlera-Session", "create")

				log.WithFields(log.Fields{
					"reqid": sessionState.id,
				}).Debug("Initialize fresh session without retries.")

				return req, nil
			}
			sessionCond.L.Unlock()
		}

		sessionCond.L.Lock()
		defer sessionCond.L.Unlock()
		for sessionID == "" {
			sessionCond.Wait()
		}

		// only 1 goroutine is awaken to create new session
		if sessionID == "" {
			sessionCreator = sessionState.id
			req.Header.Set("X-Crawlera-Session", "create")

			log.WithFields(log.Fields{
				"reqid": sessionState.id,
			}).Debug("Reinitialize session without retries.")
		} else {
			req.Header.Set("X-Crawlera-Session", sessionID)
		}

		return req, nil
	}
}

func handlerSessionResp(proxy *goproxy.ProxyHttpServer, conf *config.Config) handlerTypeResp {
	return func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		sessionState := getState(ctx)

		if resp.Header.Get("X-Crawlera-Error") == "" {
			return handlerSessionRespOK(resp, sessionState)
		}

		return handlerSessionRespError(resp, sessionState)
	}
}

func handlerSessionRespOK(resp *http.Response, sessionState *state) *http.Response {
	if sessionCreator == sessionState.id {
		sessionCond.L.Lock()
		defer sessionCond.L.Unlock()

		sessionCreator = ""
		sessionID = resp.Header.Get("X-Crawlera-Session")

		log.WithFields(log.Fields{
			"reqid":     sessionState.id,
			"sessionid": sessionID,
		}).Debug("Initialized new session.")

		sessionCond.Broadcast()
	}

	return resp
}

func handlerSessionRespError(resp *http.Response, sessionState *state) *http.Response {
	sessionCond.L.Lock()
	defer sessionCond.L.Unlock()

	if resp.Header.Get("X-Crawlera-Session") == sessionID {
		sessionID = ""
	}
	for !(sessionCreator == "" || sessionCreator == sessionState.id || sessionID != "") {
		sessionCond.Wait()
	}

	req := resp.Request
	if sessionID != "" {
		req.Header.Set("X-Crawlera-Session", sessionID)
		if newResp, err := sessionHTTPClient.Do(req); err == nil {
			resp = newResp
		}
		return resp
	}

	sessionCreator = sessionState.id
	defer func() { sessionCreator = "" }()
	req.Header.Set("X-Crawlera-Session", "create")

	log.WithFields(log.Fields{
		"reqid": sessionState.id,
	}).Info("Retry without new session, fetching new one.")

	if newResp, err := sessionHTTPClient.Do(req); err == nil && newResp.Header.Get("X-Crawlera-Error") == "" {
		sessionID = newResp.Header.Get("X-Crawlera-Session")

		log.WithFields(log.Fields{
			"reqid":     sessionState.id,
			"sessionid": sessionID,
		}).Info("Got new session after retry.")

		sessionCond.Broadcast()

		return newResp
	}

	log.WithFields(log.Fields{
		"reqid": sessionState.id,
	}).Info("Failed to get new session.")

	sessionCond.Signal()

	return resp

}

func init() {
	sessionCond = sync.NewCond(&sync.Mutex{})
}
