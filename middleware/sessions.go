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

		// Empty session id and absent creator means that no session is
		// created. This is our chance to create new one.
		//
		// For example, if session ID is != "", it means that we can use it
		// but not empty creator means that session create process is in
		// progress and since we want to use as less sessions as possible, it
		// is better to wait for it.
		if sess.id == "" && sess.creator == "" {
			sess.cond.L.Lock()
			// Classic: https://en.wikipedia.org/wiki/Double-checked_locking
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

		// If we pass to this point, we either have non-empty session id or
		// non-empty creator. As it was told before, if we have non-empty
		// creator, it means that session is creating now and we have to wait
		// for notification that it can be used.
		//
		// Here is the contract: creator HAVE to be consistent. Gorouitne
		// which holds 'creator' has to set it to empty at the end of its
		// work.
		sess.cond.L.Lock()
		defer sess.cond.L.Unlock()
		for sess.id == "" {
			sess.cond.Wait()
		}

		// If session is empty (it can be that we exit from condition above)
		// and before pass to this instruction something is changed. This
		// means we have to create new session.
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

		// This is specific of goproxy. If request was cancelled for some
		// reason (client is closed connection for example), then resp is nil
		// and ctx keeps error. But also, if X-Crawlera-Error header is set,
		// it means that request is error.
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
	// it is perfectly fine that session 'dies' during some requests are
	// performing and they can be finished fine. Anyway, we have to
	// send OK response back to client but we do not need to update the
	// session.
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

	// Response was bad and we have to create new session. First, let's
	// flush its ID.
	if ctx.Req.Header.Get("X-Crawlera-Session") == sess.id {
		sess.id = ""
	}

	// Now it can be that session is recreating now. Or we are creator.
	// Or someone in parallel has recreated the session. Let's wait
	// for any condition and proceed next to figure out what to do.
	for !(sess.creator == "" || sess.creator == rstate.ID || sess.id != "") {
		sess.cond.Wait()
	}

	var resp *http.Response
	req := ctx.Req
	// New session ID appears. Great, let's retry with it. But if it fail,
	// we'll return failed response because we cannot block here forever.
	if sess.id != "" {
		req.Header.Set("X-Crawlera-Session", sess.id)
		if newResp, err := rstate.DoCrawleraRequest(s.httpClient, req); err == nil {
			resp = newResp
		}
		return resp
	}

	// If session ID is empty, we have only 1 situation when we can get here:
	// if we have to create new session. Let's do that.
	sess.creator = rstate.ID
	defer func() { sess.creator = "" }()
	req.Header.Set("X-Crawlera-Session", "create")

	log.WithFields(log.Fields{
		"request-id": rstate.ID,
		"client-id":  rstate.ClientID,
	}).Info("Retry without new session, fetching new one.")

	if newResp, err := rstate.DoCrawleraRequest(s.httpClient, req); err == nil && newResp.Header.Get("X-Crawlera-Error") == "" {
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

// NewSessionsMiddleware returns middleware which is responsible for
// automatic session management.
func NewSessionsMiddleware(conf *config.Config, proxy *goproxy.ProxyHttpServer) Middleware {
	ware := &sessionsMiddleware{}
	ware.mtype = middlewareTypeSessions

	ware.httpClient = &http.Client{
		Timeout:   sessionClientTimeout,
		Transport: proxy.Tr,
	}
	ware.clients = &sync.Map{}

	return ware
}
