package middleware

import "time"

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
			}
		}
	}
}

func (s *sessionManager) requestNewSession(feedback *sessionIDRequest) {
	newSessionChan := make(chan string, 1)
	feedback.channel <- chan<- string(newSessionChan)

	var timeAfter <-chan time.Time
	if feedback.retry {
		timeAfter = time.After(sessionClientTimeoutRetry)
	} else {
		timeAfter = time.After(sessionClientTimeout)
	}

	for {
		select {
		case brokenSession := <-s.brokenSessionChan:
			if s.id == brokenSession {
				s.id = ""
			}
		case newSession, notClosed := <-newSessionChan:
			if notClosed {
				s.id = newSession
			}
			return
		case <-timeAfter:
			return
		}
	}
}

func newSessionManager() *sessionManager {
	return &sessionManager{
		requestIDChan:     make(chan *sessionIDRequest),
		brokenSessionChan: make(chan string),
	}
}
