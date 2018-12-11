package layers

import (
	"sync"

	"github.com/9seconds/httransform"
	"github.com/karlseguin/ccache"
	log "github.com/sirupsen/logrus"

	"github.com/scrapinghub/crawlera-headless-proxy/config"
)

const (
	sessionsChansMaxSize      = 100
	sessionsChansItemsToPrune = sessionsChansMaxSize / 2
	sessionsChansTTL          = sessionClientTimeout * 2
)

type SessionsLayer struct {
	apiKey       string
	crawleraHost string
	crawleraPort int
	clients      *sync.Map
	sessionChans *ccache.Cache
	executor     httransform.Executor
}

func (s *SessionsLayer) OnRequest(state *httransform.LayerState) error {
	clientID := getClientID(state)
	mgrRaw, loaded := s.clients.LoadOrStore(clientID,
		newSessionManager(s.apiKey, s.crawleraHost, s.crawleraPort))
	mgr := mgrRaw.(*sessionManager)
	if !loaded {
		go mgr.Start()
	}

	switch value := mgr.getSessionID(false).(type) {
	case string:
		state.RequestHeaders.SetString("X-Crawlera-Session", value)
	case chan<- string:
		state.RequestHeaders.SetString("X-Crawlera-Session", "create")
		s.sessionChans.Set(clientID, value, sessionsChansTTL)
	}

	return nil
}

func (s *SessionsLayer) OnResponse(state *httransform.LayerState, err error) {
	if err != nil {
		return
	}

	if !isCrawleraError(state) {
		s.onResponseOK(state)
		return
	}

	getMetrics(state).NewCrawleraError()
	s.onResponseError(state)
}

func (s *SessionsLayer) onResponseOK(state *httransform.LayerState) {
	if item := s.sessionChans.Get(getClientID(state)); item != nil {
		channel := item.Value().(chan<- string)
		defer close(channel)

		if !item.Expired() {
			sessionID, _ := state.ResponseHeaders.GetString("x-crawlera-session")
			channel <- sessionID
			getLogger(state).WithFields(log.Fields{
				"session-id": sessionID,
			}).Info("Initialized new session")
		}
	}
}

func (s *SessionsLayer) onResponseError(state *httransform.LayerState) {
	clientID := getClientID(state)
	mgrRaw, _ := s.clients.Load(clientID)
	mgr := mgrRaw.(*sessionManager)

	brokenSessionID, _ := state.ResponseHeaders.GetString("x-crawlera-session")
	mgr.getBrokenSessionChan() <- brokenSessionID

	if item := s.sessionChans.Get(clientID); item != nil {
		close(item.Value().(chan<- string))
	}

	switch value := mgr.getSessionID(true).(type) {
	case chan<- string:
		s.onResponseErrorRetryCreateSession(state, value)
	case string:
		s.onResponseErrorRetryWithSession(state, mgr, value)
	}
}

func (s *SessionsLayer) onResponseErrorRetryCreateSession(state *httransform.LayerState, channel chan<- string) {
	defer close(channel)

	logger := getLogger(state)
	state.Request.Header.Set("X-Crawlera-Session", "create")
	s.executeRequest(state)

	if isCrawleraResponseError(state) {
		log.Warn("Could not obtain new session even after retry")
		return
	}

	sessionID := state.Response.Header.Peek("X-Crawlera-Session")
	channel <- string(sessionID)

	logger.WithFields(log.Fields{
		"session-id": sessionID,
	}).Info("Got fresh session after retry.")
}

func (s *SessionsLayer) onResponseErrorRetryWithSession(state *httransform.LayerState, mgr *sessionManager, sessionID string) {
	state.Request.Header.Set("X-Crawlera-Session", sessionID)
	logger := getLogger(state).WithFields(log.Fields{
		"session-id": sessionID,
	})
	s.executeRequest(state)

	if isCrawleraResponseError(state) {
		mgr.getBrokenSessionChan() <- sessionID
		logger.Info("Request failed even with new session ID after retry")
		return
	}

	logger.Info("Request succeed with new session ID after retry")
}

func (s *SessionsLayer) executeRequest(state *httransform.LayerState) {
	state.Response.Reset()
	state.Response.Header.DisableNormalizing()
	s.executor(state)

	state.ResponseHeaders.Clear()
	httransform.ParseHeaders(state.ResponseHeaders, state.Response.Header.Header())
}

func NewSessionsLayer(conf *config.Config, executor httransform.Executor) httransform.Layer {
	return &SessionsLayer{
		crawleraHost: conf.CrawleraHost,
		crawleraPort: conf.CrawleraPort,
		apiKey:       conf.APIKey,
		clients:      &sync.Map{},
		sessionChans: ccache.New(ccache.Configure().MaxSize(sessionsChansMaxSize).ItemsToPrune(sessionsChansItemsToPrune)),
		executor:     executor,
	}
}
