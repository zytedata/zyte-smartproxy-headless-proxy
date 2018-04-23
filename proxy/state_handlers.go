package proxy

import (
	"net/http"

	"github.com/elazarl/goproxy"
	uuid "github.com/hashicorp/go-uuid"
	"github.com/juju/errors"
	log "github.com/sirupsen/logrus"

	"github.com/9seconds/crawlera-headless-proxy/config"
)

type state struct {
	id string
}

func newState() (*state, error) {
	newId, err := uuid.GenerateUUID()
	if err != nil {
		return nil, errors.Annotate(err, "Cannot generate uniq id")
	}

	return &state{
		id: newId,
	}, nil
}

func getState(ctx *goproxy.ProxyCtx) *state {
	return ctx.UserData.(*state)
}

func handlerStateReq(proxy *goproxy.ProxyHttpServer, conf *config.Config) handlerTypeReq {
	return func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		state, err := newState()
		if err != nil {
			log.Fatalf("Cannot create new state of request")
		}
		ctx.UserData = state

		return req, nil
	}
}
