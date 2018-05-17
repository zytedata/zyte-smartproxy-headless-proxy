package proxy

import (
	"crypto/md5"
	"io"
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
	clientID         []byte
	requestStarted   time.Time
	crawleraStarted  time.Time
	crawleraFinished time.Time
}

func (sh *stateHandler) installRequest(proxy *goproxy.ProxyHttpServer, conf *config.Config) handlerTypeReq {
	return func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		state, err := newState(req.RemoteAddr, req.Header.Get("User-Agent"))
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

	hash := md5.New()
	io.WriteString(hash, remoteAddr)
	hash.Write([]byte{0})
	io.WriteString(hash, userAgent)

	return &state{
		id:             newID,
		clientID:       hash.Sum(nil),
		requestStarted: time.Now(),
	}, nil
}

func getState(ctx *goproxy.ProxyCtx) *state {
	return ctx.UserData.(*state)
}

func newStateHandler() handlerReqInterface {
	return &stateHandler{}
}
