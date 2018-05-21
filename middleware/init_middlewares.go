package middleware

import (
	"crypto/hmac"
	"crypto/md5"
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

type RequestState struct {
	ID               string
	ClientID         string
	CrawleraRequests uint8
	RequestStarted   time.Time

	seenMiddlewares map[middlewareType]struct{}
	requestFinished *time.Time
	crawleraTimes   []time.Time
}

func (rs *RequestState) Finish() {
	if rs.requestFinished != nil {
		timeNow := time.Now()
		rs.requestFinished = &timeNow
	}
}

func (rs *RequestState) Elapsed() time.Duration {
	finishedAt := time.Now()
	if rs.requestFinished != nil {
		finishedAt = *rs.requestFinished
	}

	return finishedAt.Sub(rs.RequestStarted)
}

func (rs *RequestState) StartCrawleraRequest() (err error) {
	if len(rs.crawleraTimes)%2 == 0 {
		rs.crawleraTimes = append(rs.crawleraTimes, time.Now())
	} else {
		err = errors.New("Crawlera request already started")
	}
	rs.CrawleraRequests++

	return
}

func (rs *RequestState) FinishCrawleraRequest() (err error) {
	if len(rs.crawleraTimes)%2 == 1 {
		rs.crawleraTimes = append(rs.crawleraTimes, time.Now())
	} else {
		err = errors.New("Crawlera request already finished")
	}

	return
}

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

func (rs *RequestState) DoRequest(client *http.Client, req *http.Request) (*http.Response, error) {
	if err := rs.StartCrawleraRequest(); err != nil {
		return nil, errors.Annotate(err, "Cannot start crawlera request")
	}
	defer rs.FinishCrawleraRequest()

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

	hash := hmac.New(md5.New, []byte(remoteAddr))
	if _, err := io.WriteString(hash, req.UserAgent()); err != nil {
		return nil, errors.Annotate(err, "Cannot generate client id")
	}

	return &RequestState{
		ID:             newID,
		ClientID:       fmt.Sprintf("%x", hash.Sum(nil)),
		RequestStarted: time.Now(),
	}, nil
}

func InitMiddlewares(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
	state, err := newRequestState(req)
	if err != nil {
		log.Fatalf("Cannot create new state of request")
	}
	ctx.UserData = state

	return req, nil
}

func GetRequestState(ctx *goproxy.ProxyCtx) *RequestState {
	return ctx.UserData.(*RequestState)
}
