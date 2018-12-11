package layers

import (
	"sync"

	"github.com/9seconds/httransform"
	log "github.com/sirupsen/logrus"

	"github.com/scrapinghub/crawlera-headless-proxy/config"
)

type SessionsLayer struct {
	apiKey       string
	crawleraHost string
	crawleraPort int
	clients      *sync.Map
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
		state.Set(sessionChanContextType, value)
	}

	return nil
}

func (s *SessionsLayer) OnResponse(state *httransform.LayerState, err error) {
	if channelUntyped, ok := state.Get(sessionChanContextType); ok && err != nil {
		close(channelUntyped.(chan<- string))
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
	if channelUntyped, ok := state.Get(sessionChanContextType); ok {
		sessionID, _ := state.ResponseHeaders.GetString("x-crawlera-session")
		channelUntyped.(chan<- string) <- sessionID
		close(channelUntyped.(chan<- string))

		getMetrics(state).NewSessionCreated()

		getLogger(state).WithFields(log.Fields{
			"session-id": sessionID,
		}).Info("Initialized new session")
	}
}

func (s *SessionsLayer) onResponseError(state *httransform.LayerState) {
	clientID := getClientID(state)
	mgrRaw, _ := s.clients.Load(clientID)
	mgr := mgrRaw.(*sessionManager)

	if channelUntyped, ok := state.Get(sessionChanContextType); ok {
		close(channelUntyped.(chan<- string))
	}

	brokenSessionID, _ := state.ResponseHeaders.GetString("x-crawlera-session")
	mgr.getBrokenSessionChan() <- brokenSessionID

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

	sessionID := string(state.Response.Header.Peek("X-Crawlera-Session"))
	channel <- sessionID
	getMetrics(state).NewSessionCreated()

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
	httransform.ParseHeaders(state.ResponseHeaders, state.Response.Header.Header()) // nolint: errcheck
}

func NewSessionsLayer(conf *config.Config, executor httransform.Executor) httransform.Layer {
	return &SessionsLayer{
		crawleraHost: conf.CrawleraHost,
		crawleraPort: conf.CrawleraPort,
		apiKey:       conf.APIKey,
		clients:      &sync.Map{},
		executor:     executor,
	}
}
