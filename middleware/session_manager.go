package middleware

import (
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	sessionClientTimeout      = 180 * time.Second
	sessionClientTimeoutRetry = 30 * time.Second
)

type sessionManager struct {
	id                string
	requestIDChan     chan *sessionIDRequest
	brokenSessionChan chan string
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

func (s *sessionManager) requestNewSession(feedback *sessionIDRequest) {
	newSessionChan := make(chan string, 1)
	feedback.channel <- (chan<- string(newSessionChan))

	timeAfter := s.getTimeoutChannel(feedback.retry)
	for {
		select {
		case brokenSession := <-s.brokenSessionChan:
			if s.id == brokenSession {
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

func newSessionManager() *sessionManager {
	return &sessionManager{
		requestIDChan:     make(chan *sessionIDRequest),
		brokenSessionChan: make(chan string),
	}
}
