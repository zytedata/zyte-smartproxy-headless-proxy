package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/elazarl/goproxy"
	log "github.com/sirupsen/logrus"

	"github.com/9seconds/crawlera-headless-proxy/config"
)

const sessionClientTimeout = 30 * time.Second

type sessionsMiddleware struct {
	UniqBase

	httpClient *http.Client
	clients    *sync.Map
}

type sessionState struct {
	id      string
	creator string
	cond    *sync.Cond
}

func (s *sessionsMiddleware) OnRequest() ReqType {
	return s.BaseOnRequest(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		sess := s.getSession(ctx)
		rstate := GetRequestState(ctx)

		if sess.id == "" && sess.creator == "" {
			sess.cond.L.Lock()
			if sess.id == "" && sess.creator == "" {
				sess.id = ""
				sess.creator = rstate.ID
				sess.cond.L.Unlock()
				req.Header.Set("X-Crawlera-Session", "create")

				log.WithFields(log.Fields{
					"request-id": rstate.ID,
					"client-id":  rstate.ClientID,
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

		if sess.id == "" {
			sess.creator = rstate.ID
			req.Header.Set("X-Crawlera-Session", "create")

			log.WithFields(log.Fields{
				"request-id": rstate.ID,
				"client-id":  rstate.ClientID,
			}).Debug("Reinitialize session without retries.")
		} else {
			req.Header.Set("X-Crawlera-Session", sess.id)
		}

		return req, nil
	})
}

func (s *sessionsMiddleware) OnResponse() RespType {
	return s.BaseOnResponse(func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		sess := s.getSession(ctx)
		rstate := GetRequestState(ctx)

		err := ""
		if resp != nil {
			err = resp.Header.Get("X-Crawlera-Error")
			if err == "" {
				return s.sessionRespOK(resp, rstate, sess)
			}
		}

		log.WithFields(log.Fields{
			"request-id": rstate.ID,
			"client-id":  rstate.ClientID,
			"respnil":    resp,
			"error":      err,
			"ctx-error":  ctx.Error,
		}).Debug("Reason of failed response.")

		return s.sessionRespError(rstate, sess, ctx)
	})
}

func (s *sessionsMiddleware) sessionRespOK(resp *http.Response, rstate *RequestState, sess *sessionState) *http.Response {
	if sess.creator == rstate.ID {
		sess.cond.L.Lock()
		defer sess.cond.L.Unlock()

		sess.creator = ""
		sess.id = resp.Header.Get("X-Crawlera-Session")

		log.WithFields(log.Fields{
			"request-id": rstate.ID,
			"client-id":  rstate.ClientID,
			"sessionid":  sess.id,
		}).Debug("Initialized new session.")

		sess.cond.Broadcast()
	}
	return resp
}

func (s *sessionsMiddleware) sessionRespError(rstate *RequestState, sess *sessionState, ctx *goproxy.ProxyCtx) *http.Response {
	sess.cond.L.Lock()
	defer sess.cond.L.Unlock()

	if ctx.Req.Header.Get("X-Crawlera-Session") == sess.id {
		sess.id = ""
	}
	for !(sess.creator == "" || sess.creator == rstate.ID || sess.id != "") {
		sess.cond.Wait()
	}

	var resp *http.Response
	req := ctx.Req
	if sess.id != "" {
		req.Header.Set("X-Crawlera-Session", sess.id)
		if newResp, err := rstate.DoRequest(s.httpClient, req); err == nil {
			resp = newResp
		}
		return resp
	}

	sess.creator = rstate.ID
	defer func() { sess.creator = "" }()
	req.Header.Set("X-Crawlera-Session", "create")

	log.WithFields(log.Fields{
		"request-id": rstate.ID,
		"client-id":  rstate.ClientID,
	}).Info("Retry without new session, fetching new one.")

	if newResp, err := rstate.DoRequest(s.httpClient, req); err == nil && newResp.Header.Get("X-Crawlera-Error") == "" {
		sess.id = newResp.Header.Get("X-Crawlera-Session")

		log.WithFields(log.Fields{
			"request-id": rstate.ID,
			"client-id":  rstate.ClientID,
			"sessionid":  sess.id,
		}).Info("Got new session after retry.")

		sess.cond.Broadcast()

		return newResp
	}

	log.WithFields(log.Fields{
		"request-id": rstate.ID,
		"client-id":  rstate.ClientID,
	}).Info("Failed to get new session.")

	sess.cond.Signal()

	return resp
}

func (s *sessionsMiddleware) getSession(ctx *goproxy.ProxyCtx) *sessionState {
	clientID := GetRequestState(ctx).ClientID
	sessionStateRaw, _ := s.clients.LoadOrStore(clientID, newSessionState())
	return sessionStateRaw.(*sessionState)
}

func newSessionState() *sessionState {
	return &sessionState{
		cond: sync.NewCond(&sync.Mutex{}),
	}
}

func NewSessionsMiddleware(conf *config.Config, proxy *goproxy.ProxyHttpServer) *sessionsMiddleware {
	ware := &sessionsMiddleware{}
	ware.conf = conf
	ware.proxy = proxy
	ware.mtype = middlewareTypeSessions

	ware.httpClient = &http.Client{
		Timeout:   sessionClientTimeout,
		Transport: proxy.Tr,
	}
	ware.clients = &sync.Map{}

	return ware
}
