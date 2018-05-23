package middleware

import "time"

const sessionClientTimeout = 180 * time.Second

type sessionManager struct {
	id                string
	requestIDChan     chan chan<- interface{} // this can return either string or chan<- string
	brokenSessionChan chan string
}

func (s *sessionManager) getSessionID() interface{} {
	respChan := make(chan interface{})
	defer close(respChan)

	s.requestIDChan <- respChan

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
				feedback <- s.id
			}
		case brokenSession := <-s.brokenSessionChan:
			if s.id == brokenSession {
				s.id = ""
			}
		}
	}
}

func (s *sessionManager) requestNewSession(feedback chan<- interface{}) {
	newSessionChan := make(chan string, 1)
	feedback <- chan<- string(newSessionChan)

	timeAfter := time.After(sessionClientTimeout)
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
		requestIDChan:     make(chan chan<- interface{}),
		brokenSessionChan: make(chan string),
	}
}
