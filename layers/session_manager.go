package layers

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	sessionClientTimeout      = 180 * time.Second
	sessionClientTimeoutRetry = 30 * time.Second
	sessionAPITimeout         = 10 * time.Second
	sessionTTL                = 5 * time.Minute

	sessionUserAgent = "zyte-proxy-headless-proxy"
)

type sessionManager struct {
	id            string
	apiKey        string
	zyteProxyHost string
	lastUsed      time.Time

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
	go s.startZyteProxyAPISessionDeleter()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case feedback := <-s.requestIDChan:
			if s.id == "" {
				s.requestNewSession(feedback)
			} else {
				feedback.channel <- s.id
				s.lastUsed = time.Now()
			}
		case brokenSession := <-s.brokenSessionChan:
			if s.id == brokenSession {
				s.sessionsToDelete <- brokenSession
				s.id = ""
				s.lastUsed = time.Time{}
			} else {
				log.WithFields(log.Fields{
					"current-id": s.id,
					"broken-id":  brokenSession,
				}).Debug("Unknown broken session has been reported.")
			}
		case <-ticker.C:
			if !s.lastUsed.IsZero() && s.id != "" && time.Since(s.lastUsed) >= sessionTTL {
				log.WithFields(log.Fields{
					"id": s.id,
				}).Debug("Delete session by timeout")

				s.sessionsToDelete <- s.id
				s.id = ""
				s.lastUsed = time.Time{}
			}
		}
	}
}

func (s *sessionManager) startZyteProxyAPISessionDeleter() {
	for sessionID := range s.sessionsToDelete {
		if sessionID == "" {
			continue
		}

		if err := s.deleteZyteProxySession(sessionID); err != nil {
			log.WithFields(log.Fields{
				"session-id": sessionID,
				"error":      err,
			}).Warn("Cannot delete session from Zyte Smart Proxy Manager")
		} else {
			log.WithFields(log.Fields{
				"session-id": sessionID,
			}).Warn("Session was deleted from Zyte Smart Proxy Manager")
		}
	}
}

func (s *sessionManager) deleteZyteProxySession(sessionID string) error {
	apiURL := url.URL{
		Scheme: "http",
		Host:   s.zyteProxyHost,
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

	defer resp.Body.Close() // nolint: errcheck

	io.Copy(ioutil.Discard, resp.Body) // nolint: errcheck, gosec

	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("response status code is %d", resp.StatusCode)
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
				s.lastUsed = time.Time{}
			} else {
				log.WithFields(log.Fields{
					"current-id": s.id,
					"broken-id":  brokenSession,
				}).Debug("Unknown broken session has been reported.")
			}
		case newSession, notClosed := <-newSessionChan:
			if notClosed {
				s.id = newSession
				s.lastUsed = time.Now()
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

func newSessionManager(apiKey, zyteProxyHost string, zyteProxyPort int) *sessionManager {
	return &sessionManager{
		apiKey:            apiKey,
		zyteProxyHost:     net.JoinHostPort(zyteProxyHost, strconv.Itoa(zyteProxyPort)),
		requestIDChan:     make(chan *sessionIDRequest),
		brokenSessionChan: make(chan string),
		sessionsToDelete:  make(chan string, 1),
	}
}
