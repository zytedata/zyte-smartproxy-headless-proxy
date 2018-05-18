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
	httpClient *http.Client
	clients    *sync.Map
}

type sessionState struct {
	id      string
	creator string
	cond    *sync.Cond
}

func (sh *sessionHandler) installRequest(proxy *goproxy.ProxyHttpServer, conf *config.Config) handlerTypeReq {
	return func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		requestState := getState(ctx)
		sessionStateRaw, _ := sh.clients.LoadOrStore(requestState.clientID, newSessionState())
		sess := sessionStateRaw.(*sessionState)

		if sess.id == "" && sess.creator == "" {
			sess.cond.L.Lock()
			if sess.id == "" && sess.creator == "" {
				sess.id = ""
				sess.creator = requestState.id
				sess.cond.L.Unlock()
				req.Header.Set("X-Crawlera-Session", "create")

				log.WithFields(log.Fields{
					"reqid": requestState.id,
				}).Debug("Initialize fresh session without retries.")

				return req, nil
			}
			sess.cond.L.Unlock()
		}

		sess.cond.L.Lock()
		defer sess.cond.L.Unlock()
		for sess.id == "" {
			sess.cond.Wait()
		}

		// only 1 goroutine is awaken to create new session
		if sess.id == "" {
			sess.creator = requestState.id
			req.Header.Set("X-Crawlera-Session", "create")

			log.WithFields(log.Fields{
				"reqid": requestState.id,
			}).Debug("Reinitialize session without retries.")
		} else {
			req.Header.Set("X-Crawlera-Session", sess.id)
		}

		return req, nil
	}
}

func (sh *sessionHandler) installResponse(proxy *goproxy.ProxyHttpServer, conf *config.Config) handlerTypeResp {
	return func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		requestState := getState(ctx)
		sessionStateRaw, _ := sh.clients.LoadOrStore(requestState.clientID, newSessionState())
		sess := sessionStateRaw.(*sessionState)

		if resp != nil && resp.Header.Get("X-Crawlera-Error") == "" {
			return sh.handlerSessionRespOK(resp, requestState, sess)
		}

		return sh.handlerSessionRespError(requestState, sess, ctx)
	}
}

func (sh *sessionHandler) handlerSessionRespOK(resp *http.Response, requestState *state, sess *sessionState) *http.Response {
	if sess.creator == requestState.id {
		sess.cond.L.Lock()
		defer sess.cond.L.Unlock()

		sess.creator = ""
		sess.id = resp.Header.Get("X-Crawlera-Session")

		log.WithFields(log.Fields{
			"reqid":     requestState.id,
			"sessionid": sess.id,
		}).Debug("Initialized new session.")

		sess.cond.Broadcast()
	}
	return resp
}

func (sh *sessionHandler) handlerSessionRespError(requestState *state, sess *sessionState, ctx *goproxy.ProxyCtx) *http.Response {
	sess.cond.L.Lock()
	defer sess.cond.L.Unlock()

	if ctx.Req.Header.Get("X-Crawlera-Session") == sess.id {
		sess.id = ""
	}
	for !(sess.creator == "" || sess.creator == requestState.id || sess.id != "") {
		sess.cond.Wait()
	}

	var resp *http.Response
	req := ctx.Req
	if sess.id != "" {
		req.Header.Set("X-Crawlera-Session", sess.id)
		if newResp, err := sh.httpClient.Do(req); err == nil {
			resp = newResp
		}
		return resp
	}

	sess.creator = requestState.id
	defer func() { sess.creator = "" }()
	req.Header.Set("X-Crawlera-Session", "create")

	log.WithFields(log.Fields{
		"reqid": requestState.id,
	}).Info("Retry without new session, fetching new one.")

	if newResp, err := sh.httpClient.Do(req); err == nil && newResp.Header.Get("X-Crawlera-Error") == "" {
		sess.id = newResp.Header.Get("X-Crawlera-Session")

		log.WithFields(log.Fields{
			"reqid":     requestState.id,
			"sessionid": sess.id,
		}).Info("Got new session after retry.")

		sess.cond.Broadcast()

		return newResp
	}

	log.WithFields(log.Fields{
		"reqid": requestState.id,
	}).Info("Failed to get new session.")

	sess.cond.Signal()

	return resp
}

func newSessionState() *sessionState {
	return &sessionState{
		cond: sync.NewCond(&sync.Mutex{}),
	}
}

func newSessionHandler(transport http.RoundTripper) handlerReqRespInterface {
	return &sessionHandler{
		clients: &sync.Map{},
		httpClient: &http.Client{
			Timeout:   sessionClientTimeout,
			Transport: transport,
		},
	}
}
