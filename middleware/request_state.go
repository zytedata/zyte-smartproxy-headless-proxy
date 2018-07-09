package middleware

import (
	"crypto/hmac"
	"crypto/sha1"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/elazarl/goproxy"
	uuid "github.com/hashicorp/go-uuid"
	"github.com/juju/errors"

	"bitbucket.org/scrapinghub/crawlera-headless-proxy/stats"
)

// RequestState stores basic metadata of every request (who, when etc)
// It is attached to goproxy.ProxyCtx UserData field.
type RequestState struct {
	ID               string
	ClientID         string
	CrawleraRequests uint8
	RequestStarted   time.Time

	seenMiddlewares map[middlewareType]struct{}
	requestFinished time.Time
	crawleraTimes   []time.Time

	crawleraRequestsChan chan<- struct{}
	crawleraTimesChan    chan<- time.Duration
}

// Finish sets request as finished.
func (rs *RequestState) Finish() {
	if rs.requestFinished.IsZero() {
		rs.requestFinished = time.Now()
	}
}

// Elapsed returns duration of the request (overall).
func (rs *RequestState) Elapsed() time.Duration {
	var finishedAt time.Time
	if rs.requestFinished.IsZero() {
		finishedAt = time.Now()
	} else {
		finishedAt = rs.requestFinished
	}

	return finishedAt.Sub(rs.RequestStarted)
}

// StartCrawleraRequest adds a fact that request to Crawlera is
// performing.
func (rs *RequestState) StartCrawleraRequest() error {
	if len(rs.crawleraTimes)%2 == 0 {
		rs.crawleraTimes = append(rs.crawleraTimes, time.Now())
	} else {
		return errors.New("Crawlera request already started")
	}
	rs.CrawleraRequests++
	rs.crawleraRequestsChan <- struct{}{}

	return nil
}

// FinishCrawleraRequest adds a fact that request to Crawlera was
// performed
func (rs *RequestState) FinishCrawleraRequest() error {
	if len(rs.crawleraTimes)%2 == 1 {
		rs.crawleraTimes = append(rs.crawleraTimes, time.Now())
	} else {
		return errors.New("Crawlera request already finished")
	}

	l := len(rs.crawleraTimes) - 1
	rs.crawleraTimesChan <- rs.crawleraTimes[l].Sub(rs.crawleraTimes[l-1])

	return nil
}

// CrawleraTimes returns a list of durations on accessing Crawlera.
func (rs *RequestState) CrawleraTimes() []time.Duration {
	var crawleraTimes []time.Time
	var duration []time.Duration

	if len(rs.crawleraTimes)%2 == 1 {
		crawleraTimes = append(crawleraTimes, rs.crawleraTimes...)
		crawleraTimes = append(crawleraTimes, time.Now())
	} else {
		crawleraTimes = rs.crawleraTimes
	}

	for i := 0; i < len(crawleraTimes); i += 2 {
		duration = append(duration, crawleraTimes[i+1].Sub(crawleraTimes[i]))
	}

	return duration
}

// CrawleraElapsed returns a duration which was spent accessing Crawlera.
func (rs *RequestState) CrawleraElapsed() time.Duration {
	var duration time.Duration

	for _, val := range rs.CrawleraTimes() {
		duration += val
	}

	return duration
}

// DoCrawleraRequest sends an http.Request with http.Client wrapping it
// around with Start/Finish crawlera request.
func (rs *RequestState) DoCrawleraRequest(client *http.Client, req *http.Request) (*http.Response, error) {
	if err := rs.StartCrawleraRequest(); err != nil {
		return nil, errors.Annotate(err, "Cannot start crawlera request")
	}
	defer rs.FinishCrawleraRequest() // nolint: errcheck

	return client.Do(req)
}

func newRequestState(req *http.Request, statsContainer *stats.Stats) (*RequestState, error) {
	newID, err := uuid.GenerateUUID()
	if err != nil {
		return nil, errors.Annotate(err, "Cannot generate unique id")
	}

	remoteAddr := ""
	if host, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
		remoteAddr = host
	}

	hash := hmac.New(sha1.New, []byte(remoteAddr))
	if _, err := io.WriteString(hash, req.UserAgent()); err != nil {
		return nil, errors.Annotate(err, "Cannot generate client id")
	}

	return &RequestState{
		ID:              newID,
		ClientID:        fmt.Sprintf("%x", hash.Sum(nil)),
		RequestStarted:  time.Now(),
		seenMiddlewares: map[middlewareType]struct{}{},

		crawleraRequestsChan: statsContainer.CrawleraRequestsChan,
		crawleraTimesChan:    statsContainer.CrawleraTimesChan,
	}, nil
}

// GetRequestState returns a request state from goproxy context.
func GetRequestState(ctx *goproxy.ProxyCtx) *RequestState {
	return ctx.UserData.(*RequestState)
}
