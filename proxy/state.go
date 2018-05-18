package proxy

import (
	"crypto/hmac"
	"crypto/md5" // nolint: gas
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/elazarl/goproxy"
	uuid "github.com/hashicorp/go-uuid"
	"github.com/juju/errors"
	log "github.com/sirupsen/logrus"

	"github.com/9seconds/crawlera-headless-proxy/config"
)

type stateHandler struct {
}

type state struct {
	id               string
	clientID         string
	requestStarted   time.Time
	crawleraStarted  time.Time
	crawleraFinished time.Time
}

func (sh *stateHandler) installRequest(proxy *goproxy.ProxyHttpServer, conf *config.Config) handlerTypeReq {
	return func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		addr := ""
		if host, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
			addr = host
		}

		state, err := newState(addr, req.UserAgent())
		if err != nil {
			log.Fatalf("Cannot create new state of request")
		}
		ctx.UserData = state

		return req, nil
	}
}

func newState(remoteAddr string, userAgent string) (*state, error) {
	newID, err := uuid.GenerateUUID()
	if err != nil {
		return nil, errors.Annotate(err, "Cannot generate unique id")
	}

	hash := hmac.New(md5.New, []byte(remoteAddr))
	if _, err := io.WriteString(hash, userAgent); err != nil {
		return nil, errors.Annotate(err, "Cannot generate client id")
	}

	return &state{
		id:             newID,
		clientID:       fmt.Sprintf("%x", hash.Sum(nil)),
		requestStarted: time.Now(),
	}, nil
}

func getState(ctx *goproxy.ProxyCtx) *state {
	return ctx.UserData.(*state)
}

func newStateHandler() handlerReqInterface {
	return &stateHandler{}
}
