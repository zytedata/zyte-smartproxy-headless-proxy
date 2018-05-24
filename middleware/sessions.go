package middleware

import (
	"net/http"
	"sync"

	"github.com/elazarl/goproxy"
	"github.com/karlseguin/ccache"
	log "github.com/sirupsen/logrus"

	"github.com/9seconds/crawlera-headless-proxy/config"
	"github.com/9seconds/crawlera-headless-proxy/stats"
)

const (
	sessionsChansMaxSize      = 100
	sessionsChansItemsToPrune = sessionsChansMaxSize / 2
	sessionsChansTTL          = sessionClientTimeout * 2
)

type sessionsMiddleware struct {
	UniqBase

	httpClient   *http.Client
	clients      *sync.Map
	sessionChans *ccache.Cache

	sessionsCreatedChan chan<- struct{}
	allErrorsChan       chan<- struct{}
	crawleraErrorsChan  chan<- struct{}
}

func (s *sessionsMiddleware) OnRequest() ReqType {
	return s.BaseOnRequest(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		rstate := GetRequestState(ctx)

		mgrRaw, loaded := s.clients.LoadOrStore(rstate.ClientID, newSessionManager())
		mgr := mgrRaw.(*sessionManager)
		if !loaded {
			go mgr.Start()
		}

		sessionID := mgr.getSessionID(false)
		switch sessionID.(type) {
		case string:
			s.onRequestWithSession(req, sessionID.(string))
		case chan<- string:
			s.onRequestWithoutSession(req, rstate, sessionID.(chan<- string))
		}

		return req, nil
	})
}

func (s *sessionsMiddleware) onRequestWithSession(req *http.Request, sessionID string) {
	req.Header.Set("X-Crawlera-Session", sessionID)
}

func (s *sessionsMiddleware) onRequestWithoutSession(req *http.Request,
	rstate *RequestState, sessionIDChan chan<- string) {
	req.Header.Set("X-Crawlera-Session", "create")
	s.sessionChans.Set(rstate.ID, sessionIDChan, sessionsChansTTL)

	log.WithFields(log.Fields{
		"request-id": rstate.ID,
		"client-id":  rstate.ClientID,
	}).Debug("Initialize fresh session without retries.")
}

func (s *sessionsMiddleware) OnResponse() RespType {
	return s.BaseOnResponse(func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		rstate := GetRequestState(ctx)

		// This is specific of goproxy. If request was cancelled for some
		// reason (client is closed connection for example), then resp is nil
		// and ctx keeps error. But also, if X-Crawlera-Error header is set,
		// it means that request is error.
		err := ""
		if resp != nil {
			err = resp.Header.Get("X-Crawlera-Error")
			if err == "" {
				return s.sessionRespOK(resp, rstate)
			}
			s.crawleraErrorsChan <- struct{}{}
		}
		s.allErrorsChan <- struct{}{}

		log.WithFields(log.Fields{
			"request-id": rstate.ID,
			"client-id":  rstate.ClientID,
			"respnil":    resp,
			"error":      err,
			"ctx-error":  ctx.Error,
		}).Debug("Reason of failed response.")

		return s.sessionRespError(rstate, ctx)
	})
}

func (s *sessionsMiddleware) sessionRespOK(resp *http.Response, rstate *RequestState) *http.Response {
	if item := s.sessionChans.Get(rstate.ID); item != nil {
		sessionIDChan := item.Value().(chan<- string)
		defer close(sessionIDChan)

		sessionID := resp.Header.Get("X-Crawlera-Session")
		if !item.Expired() {
			sessionIDChan <- sessionID
			log.WithFields(log.Fields{
				"request-id": rstate.ID,
				"client-id":  rstate.ClientID,
				"sessionid":  sessionID,
			}).Debug("Initialized new session.")
		}
		s.sessionsCreatedChan <- struct{}{}
	}

	return resp
}

func (s *sessionsMiddleware) sessionRespError(rstate *RequestState, ctx *goproxy.ProxyCtx) *http.Response {
	mgrRaw, ok := s.clients.Load(rstate.ClientID)
	if !ok {
		log.WithFields(log.Fields{
			"client-id": rstate.ClientID,
		}).Warn("Client bypass OnRequest handler")
		return nil
	}
	mgr := mgrRaw.(*sessionManager)

	brokenSessionID := ctx.Req.Header.Get("X-Crawlera-Session")
	mgr.getBrokenSessionChan() <- brokenSessionID

	if item := s.sessionChans.Get(rstate.ID); item != nil {
		close(item.Value().(chan<- string))
	}

	sessionID := mgr.getSessionID(true)
	switch sessionID.(type) {
	case chan<- string:
		return s.sessionRespErrorWithoutSession(ctx.Req, rstate, mgr, sessionID.(chan<- string))
	default:
		return s.sessionRespErrorWithSession(ctx.Req, rstate, mgr, sessionID.(string))
	}
}

func (s *sessionsMiddleware) sessionRespErrorWithSession(req *http.Request,
	rstate *RequestState, mgr *sessionManager, sessionID string) *http.Response {
	req.Header.Set("X-Crawlera-Session", sessionID)

	resp, err := rstate.DoCrawleraRequest(s.httpClient, req)
	if err != nil || resp.Header.Get("X-Crawlera-Error") != "" {
		mgr.getBrokenSessionChan() <- sessionID

		log.WithFields(log.Fields{
			"request-id": rstate.ID,
			"client-id":  rstate.ClientID,
			"session-id": sessionID,
		}).Info("Request failed even with new session ID after retry")

		return resp
	}

	log.WithFields(log.Fields{
		"request-id": rstate.ID,
		"client-id":  rstate.ClientID,
		"session-id": sessionID,
	}).Info("Request succeed with new session ID after retry")

	return resp
}

func (s *sessionsMiddleware) sessionRespErrorWithoutSession(req *http.Request,
	rstate *RequestState, mgr *sessionManager, sessionIDChan chan<- string) *http.Response {
	defer close(sessionIDChan)

	req.Header.Set("X-Crawlera-Session", "create")
	resp, err := rstate.DoCrawleraRequest(s.httpClient, req)

	if err == nil && resp.Header.Get("X-Crawlera-Error") == "" {
		sessionID := resp.Header.Get("X-Crawlera-Session")
		sessionIDChan <- sessionID

		log.WithFields(log.Fields{
			"request-id": rstate.ID,
			"client-id":  rstate.ClientID,
			"session-id": sessionID,
		}).Info("Got fresh session after retry")

		return resp
	}

	log.WithFields(log.Fields{
		"request-id": rstate.ID,
		"client-id":  rstate.ClientID,
	}).Info("Could not obtain new session even after retry")

	return resp
}

// NewSessionsMiddleware returns middleware which is responsible for
// automatic session management.
func NewSessionsMiddleware(conf *config.Config, proxy *goproxy.ProxyHttpServer, statsContainer *stats.Stats) Middleware {
	ware := &sessionsMiddleware{}
	ware.mtype = middlewareTypeSessions

	ware.httpClient = &http.Client{
		Timeout:   sessionClientTimeout,
		Transport: proxy.Tr,
	}
	ware.clients = &sync.Map{}
	ware.sessionChans = ccache.New(ccache.Configure().MaxSize(sessionsChansMaxSize).ItemsToPrune(sessionsChansItemsToPrune))
	ware.sessionsCreatedChan = statsContainer.SessionsCreatedChan
	ware.allErrorsChan = statsContainer.AllErrorsChan
	ware.crawleraErrorsChan = statsContainer.CrawleraErrorsChan

	return ware
}
