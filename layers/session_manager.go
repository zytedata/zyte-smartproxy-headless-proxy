package layers

import (
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"time"

	"github.com/juju/errors"
	log "github.com/sirupsen/logrus"
)

const (
	sessionClientTimeout      = 180 * time.Second
	sessionClientTimeoutRetry = 30 * time.Second
	sessionAPITimeout         = 10 * time.Second

	sessionToDeleteQueueSize = 5000
	sessionUserAgent         = "crawlera-headless-proxy"
)

type sessionManager struct {
	id           string
	apiKey       string
	crawleraHost string

	requestIDChan     chan *sessionIDRequest
	brokenSessionChan chan string
	sessionsToDelete  chan string
}

type sessionIDRequest struct {
	channel chan<- interface{}
	retry   bool
}

func (s *sessionManager) getSessionID(retry bool) interface{} {
	respChan := make(chan interface{})
	defer close(respChan)

	s.requestIDChan <- &sessionIDRequest{
		channel: respChan,
		retry:   retry,
	}

	return <-respChan
}

func (s *sessionManager) getBrokenSessionChan() chan<- string {
	return s.brokenSessionChan
}

func (s *sessionManager) Start() {
	go s.startCrawleraAPISessionDeleter()

	for {
		select {
		case feedback := <-s.requestIDChan:
			if s.id == "" {
				s.requestNewSession(feedback)
			} else {
				feedback.channel <- s.id
			}
		case brokenSession := <-s.brokenSessionChan:
			if s.id == brokenSession {
				s.sessionsToDelete <- brokenSession
				s.id = ""
			} else {
				log.WithFields(log.Fields{
					"current-id": s.id,
					"broken-id":  brokenSession,
				}).Debug("Unknown broken session has been reported.")
			}
		}
	}
}

func (s *sessionManager) startCrawleraAPISessionDeleter() {
	for sessionID := range s.sessionsToDelete {
		if sessionID == "" {
			continue
		}

		if err := s.deleteCrawleraSession(sessionID); err != nil {
			log.WithFields(log.Fields{
				"session-id": sessionID,
				"error":      err,
			}).Warn("Cannot delete session from Crawlera")
		} else {
			log.WithFields(log.Fields{
				"session-id": sessionID,
			}).Warn("Session was deleted from Crawlera")
		}
	}
}

func (s *sessionManager) deleteCrawleraSession(sessionID string) error {
	apiURL := url.URL{
		Scheme: "http",
		Host:   s.crawleraHost,
		Path:   path.Join("sessions", sessionID),
	}
	req, _ := http.NewRequest("DELETE", apiURL.String(), http.NoBody) // nolint: gosec
	req.SetBasicAuth(s.apiKey, "")
	req.Header.Set("User-Agent", sessionUserAgent)

	client := &http.Client{Timeout: sessionAPITimeout}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()            // nolint: errcheck
	io.Copy(ioutil.Discard, resp.Body) // nolint: errcheck, gosec

	if resp.StatusCode >= http.StatusBadRequest {
		return errors.Errorf("Response status code is %d", resp.StatusCode)
	}

	return nil

}

func (s *sessionManager) requestNewSession(feedback *sessionIDRequest) {
	newSessionChan := make(chan string, 1)
	feedback.channel <- (chan<- string(newSessionChan))

	timeAfter := s.getTimeoutChannel(feedback.retry)
	for {
		select {
		case brokenSession := <-s.brokenSessionChan:
			if s.id == brokenSession {
				s.sessionsToDelete <- brokenSession
				s.id = ""
			} else {
				log.WithFields(log.Fields{
					"current-id": s.id,
					"broken-id":  brokenSession,
				}).Debug("Unknown broken session has been reported.")
			}
		case newSession, notClosed := <-newSessionChan:
			if notClosed {
				s.id = newSession
			}
			return
		case <-timeAfter:
			log.Debug("Timeout in waiting for the new session.")
			return
		}
	}
}

func (s *sessionManager) getTimeoutChannel(retry bool) <-chan time.Time {
	if retry {
		return time.After(sessionClientTimeoutRetry)
	}
	return time.After(sessionClientTimeout)
}

func newSessionManager(apiKey, crawleraHost string, crawleraPort int) *sessionManager {
	return &sessionManager{
		apiKey:            apiKey,
		crawleraHost:      net.JoinHostPort(crawleraHost, strconv.Itoa(crawleraPort)),
		requestIDChan:     make(chan *sessionIDRequest),
		brokenSessionChan: make(chan string),
		sessionsToDelete:  make(chan string, sessionToDeleteQueueSize),
	}
}
