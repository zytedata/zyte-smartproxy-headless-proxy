package layers

import (
	"sync"

	"github.com/9seconds/httransform/v2/errors"
	"github.com/9seconds/httransform/v2/executor"
	"github.com/9seconds/httransform/v2/layers"
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
		if channelUntyped != nil {
			close(channelUntyped.(chan<- string))
		}

		return err
	}

	if !isCrawleraError(ctx) {
		s.onResponseOK(ctx)
		return err
	}

	getMetrics(ctx).NewCrawleraError()

	return s.onResponseError(ctx)
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

func (s *SessionsLayer) onResponseError(ctx *layers.Context) error {
	clientID := getClientID(ctx)
	mgrRaw, _ := s.clients.Load(clientID)
	mgr := mgrRaw.(*sessionManager)

	if channelUntyped := ctx.Get(sessionChanContextType); channelUntyped != nil {
		close(channelUntyped.(chan<- string))
	}

	if brokenSessionID := ctx.ResponseHeaders.GetLast("x-crawlera-session").Value(); brokenSessionID != "" {
		mgr.getBrokenSessionChan() <- brokenSessionID
	}

	switch value := mgr.getSessionID(true).(type) {
	case chan<- string:
		return s.onResponseErrorRetryCreateSession(ctx, value)
	case string:
		return s.onResponseErrorRetryWithSession(ctx, mgr, value)
	}

	return errors.Annotate(nil, "Unexpected error in onResponseError", "session_manager", 0)
}

func (s *SessionsLayer) onResponseErrorRetryCreateSession(ctx *layers.Context, channel chan<- string) error {
	defer close(channel)

	logger := getLogger(ctx)
	ctx.RequestHeaders.Set("X-Crawlera-Session", "create", true)
	err := s.executeRequest(ctx)

	if err != nil || isCrawleraResponseError(ctx) {
		log.Warn("Could not obtain new session even after retry")
		return errors.Annotate(err, "Could not obtain new session even after retry", "session_manager", 0)
	}

	sessionID := ctx.ResponseHeaders.GetLast("X-Crawlera-Session").Value()
	channel <- sessionID

	getMetrics(ctx).NewSessionCreated()

	logger.WithFields(log.Fields{
		"session-id": sessionID,
	}).Info("Got fresh session after retry.")

	return nil
}

func (s *SessionsLayer) onResponseErrorRetryWithSession(ctx *layers.Context, mgr *sessionManager, sessionID string) error {
	ctx.RequestHeaders.Set("X-Crawlera-Session", sessionID, true)
	logger := getLogger(ctx).WithFields(log.Fields{
		"session-id": sessionID,
	})

	err := s.executeRequest(ctx)

	if err != nil || isCrawleraResponseError(ctx) {
		mgr.getBrokenSessionChan() <- sessionID
		logger.Info("Request failed even after retry")

		return errors.Annotate(err, "Request failed even after retry", "session_manager", 0)
	}

	logger.Info("Request succeed after retry")

	return nil
}

func (s *SessionsLayer) executeRequest(ctx *layers.Context) error {
	if err := ctx.RequestHeaders.Push(); err != nil {
		return errors.Annotate(err, "cannot sync request headers", "session_manager", 0)
	}

	if err := s.executor(ctx); err != nil {
		return errors.Annotate(err, "cannot execute a request", "session_manager", 0)
	}

	if err := ctx.ResponseHeaders.Pull(); err != nil {
		return errors.Annotate(err, "cannot read response headers", "session_manager", 0)
	}

	return nil
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
