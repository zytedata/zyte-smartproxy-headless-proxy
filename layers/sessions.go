package layers

import (
	"sync"

	"github.com/9seconds/httransform/v2/layers"
	"github.com/9seconds/httransform/v2/executor"
	log "github.com/sirupsen/logrus"

	"github.com/scrapinghub/crawlera-headless-proxy/config"
)

type SessionsLayer struct {
	apiKey       string
	crawleraHost string
	crawleraPort int
	clients      *sync.Map
	executor     executor.Executor
}

func (s *SessionsLayer) OnRequest(ctx *layers.Context) error {
	clientID := getClientID(ctx)
	mgrRaw, loaded := s.clients.LoadOrStore(clientID,
		newSessionManager(s.apiKey, s.crawleraHost, s.crawleraPort))
	mgr := mgrRaw.(*sessionManager)

	if !loaded {
		go mgr.Start()
	}

	switch value := mgr.getSessionID(false).(type) {
	case string:
		ctx.RequestHeaders.Set("X-Crawlera-Session", value, true)
	case chan<- string:
		ctx.RequestHeaders.Set("X-Crawlera-Session", "create", true)
		ctx.Set(sessionChanContextType, value)
	}

	return nil
}

func (s *SessionsLayer) OnResponse(ctx *layers.Context, err error) error {
	if channelUntyped := ctx.Get(sessionChanContextType); err != nil {
		close(channelUntyped.(chan<- string))
		return err
	}

	if !isCrawleraError(ctx) {
		s.onResponseOK(ctx)
		return err
	}

	getMetrics(ctx).NewCrawleraError()
	s.onResponseError(ctx)
	return err
}

func (s *SessionsLayer) onResponseOK(ctx *layers.Context) {
	if channelUntyped := ctx.Get(sessionChanContextType); channelUntyped != nil {
		sessionID := ctx.ResponseHeaders.GetLast("x-crawlera-session").Value()
		channelUntyped.(chan<- string) <- sessionID
		close(channelUntyped.(chan<- string))

		getMetrics(ctx).NewSessionCreated()

		getLogger(ctx).WithFields(log.Fields{
			"session-id": sessionID,
		}).Info("Initialized new session")
	}
}

func (s *SessionsLayer) onResponseError(ctx *layers.Context) {
	clientID := getClientID(ctx)
	mgrRaw, _ := s.clients.Load(clientID)
	mgr := mgrRaw.(*sessionManager)

	if channelUntyped := ctx.Get(sessionChanContextType); channelUntyped != nil {
		close(channelUntyped.(chan<- string))
	}

	brokenSessionID := ctx.ResponseHeaders.GetLast("x-crawlera-session").Value()
	mgr.getBrokenSessionChan() <- brokenSessionID

	switch value := mgr.getSessionID(true).(type) {
	case chan<- string:
		s.onResponseErrorRetryCreateSession(ctx, value)
	case string:
		s.onResponseErrorRetryWithSession(ctx, mgr, value)
	}
}

func (s *SessionsLayer) onResponseErrorRetryCreateSession(ctx *layers.Context, channel chan<- string) {
	defer close(channel)

	logger := getLogger(ctx)
	ctx.RequestHeaders.Set("X-Crawlera-Session", "create", true)
	s.executeRequest(ctx)

	if isCrawleraResponseError(ctx) {
		log.Warn("Could not obtain new session even after retry")
		return
	}

	sessionID := ctx.ResponseHeaders.GetLast("X-Crawlera-Session").Value()
	channel <- sessionID

	getMetrics(ctx).NewSessionCreated()

	logger.WithFields(log.Fields{
		"session-id": sessionID,
	}).Info("Got fresh session after retry.")
}

func (s *SessionsLayer) onResponseErrorRetryWithSession(ctx *layers.Context, mgr *sessionManager, sessionID string) {
	ctx.RequestHeaders.Set("X-Crawlera-Session", sessionID, true)
	logger := getLogger(ctx).WithFields(log.Fields{
		"session-id": sessionID,
	})

	s.executeRequest(ctx)

	if isCrawleraResponseError(ctx) {
		mgr.getBrokenSessionChan() <- sessionID
		logger.Info("Request failed even with new session ID after retry")

		return
	}

	logger.Info("Request succeed with new session ID after retry")
}

func (s *SessionsLayer) executeRequest(ctx *layers.Context) {
	//ctx.Response.Reset()
	//ctx.Response.Header.DisableNormalizing()
	s.executor(ctx)

	//ctx.ResponseHeaders.Clear()
	//httransform.ParseHeaders(ctx.ResponseHeaders, ctx.Response.Header.Header()) // nolint: errcheck
}

func NewSessionsLayer(conf *config.Config, executor executor.Executor) layers.Layer {
	return &SessionsLayer{
		crawleraHost: conf.CrawleraHost,
		crawleraPort: conf.CrawleraPort,
		apiKey:       conf.APIKey,
		clients:      &sync.Map{},
		executor:     executor,
	}
}
