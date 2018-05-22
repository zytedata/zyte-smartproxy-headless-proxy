package middleware

import (
	"net/http"

	"github.com/elazarl/goproxy"
	log "github.com/sirupsen/logrus"
)

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
