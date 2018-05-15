package proxy

import (
	"net/http"
	"sync"
	"time"

	"github.com/elazarl/goproxy"
	log "github.com/sirupsen/logrus"

	"github.com/9seconds/crawlera-headless-proxy/config"
)

const sessionClientTimeout = 30 * time.Second

type sessionHandler struct {
	httpClient     *http.Client
	sessionID      string
	sessionCreator string
	sessionCond    *sync.Cond
}

func (sh *sessionHandler) installRequest(proxy *goproxy.ProxyHttpServer, conf *config.Config) handlerTypeReq {
	return func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		sessionState := getState(ctx)

		if sh.sessionID == "" && sh.sessionCreator == "" {
			sh.sessionCond.L.Lock()
			if sh.sessionID == "" && sh.sessionCreator == "" {
				sh.sessionID = ""
				sh.sessionCreator = sessionState.id
				sh.sessionCond.L.Unlock()
				req.Header.Set("X-Crawlera-Session", "create")

				log.WithFields(log.Fields{
					"reqid": sessionState.id,
				}).Debug("Initialize fresh session without retries.")

				return req, nil
			}
			sh.sessionCond.L.Unlock()
		}

		sh.sessionCond.L.Lock()
		defer sh.sessionCond.L.Unlock()
		for sh.sessionID == "" {
			sh.sessionCond.Wait()
		}

		// only 1 goroutine is awaken to create new session
		if sh.sessionID == "" {
			sh.sessionCreator = sessionState.id
			req.Header.Set("X-Crawlera-Session", "create")

			log.WithFields(log.Fields{
				"reqid": sessionState.id,
			}).Debug("Reinitialize session without retries.")
		} else {
			req.Header.Set("X-Crawlera-Session", sh.sessionID)
		}

		return req, nil
	}
}

func (sh *sessionHandler) installResponse(proxy *goproxy.ProxyHttpServer, conf *config.Config) handlerTypeResp {
	return func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		sessionState := getState(ctx)

		if resp.Header.Get("X-Crawlera-Error") == "" {
			return sh.handlerSessionRespOK(resp, sessionState)
		}

		return sh.handlerSessionRespError(resp, sessionState)
	}
}

func (sh *sessionHandler) handlerSessionRespOK(resp *http.Response, sessionState *state) *http.Response {
	if sh.sessionCreator == sessionState.id {
		sh.sessionCond.L.Lock()
		defer sh.sessionCond.L.Unlock()

		sh.sessionCreator = ""
		sh.sessionID = resp.Header.Get("X-Crawlera-Session")

		log.WithFields(log.Fields{
			"reqid":     sessionState.id,
			"sessionid": sh.sessionID,
		}).Debug("Initialized new session.")

		sh.sessionCond.Broadcast()
	}
	return resp
}

func (sh *sessionHandler) handlerSessionRespError(resp *http.Response, sessionState *state) *http.Response {
	sh.sessionCond.L.Lock()
	defer sh.sessionCond.L.Unlock()

	if resp.Header.Get("X-Crawlera-Session") == sh.sessionID {
		sh.sessionID = ""
	}
	for !(sh.sessionCreator == "" || sh.sessionCreator == sessionState.id || sh.sessionID != "") {
		sh.sessionCond.Wait()
	}

	req := resp.Request
	if sh.sessionID != "" {
		req.Header.Set("X-Crawlera-Session", sh.sessionID)
		if newResp, err := sh.httpClient.Do(req); err == nil {
			resp = newResp
		}
		return resp
	}

	sh.sessionCreator = sessionState.id
	defer func() { sh.sessionCreator = "" }()
	req.Header.Set("X-Crawlera-Session", "create")

	log.WithFields(log.Fields{
		"reqid": sessionState.id,
	}).Info("Retry without new session, fetching new one.")

	if newResp, err := sh.httpClient.Do(req); err == nil && newResp.Header.Get("X-Crawlera-Error") == "" {
		sh.sessionID = newResp.Header.Get("X-Crawlera-Session")

		log.WithFields(log.Fields{
			"reqid":     sessionState.id,
			"sessionid": sh.sessionID,
		}).Info("Got new session after retry.")

		sh.sessionCond.Broadcast()

		return newResp
	}

	log.WithFields(log.Fields{
		"reqid": sessionState.id,
	}).Info("Failed to get new session.")

	sh.sessionCond.Signal()

	return resp
}

func newSessionHandler(transport http.RoundTripper) handlerReqRespInterface {
	return &sessionHandler{
		sessionCond: sync.NewCond(&sync.Mutex{}),
		httpClient: &http.Client{
			Timeout:   sessionClientTimeout,
			Transport: transport,
		},
	}
}
