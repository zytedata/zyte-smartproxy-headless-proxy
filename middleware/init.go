package middleware

import (
	"net/http"

	"github.com/elazarl/goproxy"
	log "github.com/sirupsen/logrus"

	"github.com/9seconds/crawlera-headless-proxy/stats"
)

// InitMiddlewares sets goproxy with middlewares. This basically
// generates and sets correct request state.
func InitMiddlewares(statsContainer *stats.Stats) ReqType {
	return func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		state, err := newRequestState(req, statsContainer)
		if err != nil {
			log.Fatalf("Cannot create new state of request")
		}
		ctx.UserData = state

		return req, nil
	}
}
