package middleware

import (
	"crypto/hmac"
	"crypto/sha1"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/elazarl/goproxy"
	uuid "github.com/hashicorp/go-uuid"
	"github.com/juju/errors"
)

// RequestState stores basic metadata of every request (who, when etc)
// It is attached to goproxy.ProxyCtx UserData field.
type RequestState struct {
	ID               string
	ClientID         string
	CrawleraRequests uint8
	RequestStarted   time.Time

	seenMiddlewares map[middlewareType]struct{}
	requestFinished *time.Time
	crawleraTimes   []time.Time
}

// Finish sets request as finished.
func (rs *RequestState) Finish() {
	if rs.requestFinished != nil {
		timeNow := time.Now()
		rs.requestFinished = &timeNow
	}
}

// Elapsed returns duration of the request (overall).
func (rs *RequestState) Elapsed() time.Duration {
	var finishedAt time.Time
	if rs.requestFinished != nil {
		finishedAt = *rs.requestFinished
	} else {
		finishedAt = time.Now()
	}

	return finishedAt.Sub(rs.RequestStarted)
}

// StartCrawleraRequest adds a fact that request to Crawlera is
// performing.
func (rs *RequestState) StartCrawleraRequest() (err error) {
	if len(rs.crawleraTimes)%2 == 0 {
		rs.crawleraTimes = append(rs.crawleraTimes, time.Now())
	} else {
		err = errors.New("Crawlera request already started")
	}
	rs.CrawleraRequests++

	return
}

// FinishCrawleraRequest adds a fact that request to Crawlera was
// performed
func (rs *RequestState) FinishCrawleraRequest() (err error) {
	if len(rs.crawleraTimes)%2 == 1 {
		rs.crawleraTimes = append(rs.crawleraTimes, time.Now())
	} else {
		err = errors.New("Crawlera request already finished")
	}

	return
}

// CrawleraElapsed returns a duration which was spent accessing Crawlera.
func (rs *RequestState) CrawleraElapsed() time.Duration {
	duration := time.Duration(0)
	crawleraTimes := []time.Time{}
	copy(crawleraTimes, rs.crawleraTimes)

	if len(crawleraTimes)%2 == 1 {
		crawleraTimes = append(crawleraTimes, time.Now())
	}

	for i := 0; i < len(crawleraTimes); i += 2 {
		duration += crawleraTimes[i+1].Sub(crawleraTimes[i])
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

func newRequestState(req *http.Request) (*RequestState, error) {
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
	}, nil
}

// InitMiddlewares sets goproxy with middlewares. This basically
// generates and sets correct request state.
func InitMiddlewares(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
	state, err := newRequestState(req)
	if err != nil {
		log.Fatalf("Cannot create new state of request")
	}
	ctx.UserData = state

	return req, nil
}

// GetRequestState returns a request state from goproxy context.
func GetRequestState(ctx *goproxy.ProxyCtx) *RequestState {
	return ctx.UserData.(*RequestState)
}
