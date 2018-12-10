package layers

import (
	"crypto/hmac"
	"crypto/sha1"
	"fmt"
	"net"
	"time"

	"github.com/9seconds/httransform"
	log "github.com/sirupsen/logrus"

	"github.com/scrapinghub/crawlera-headless-proxy/stats"
)

const (
	LogLayerContextType       = "log"
	MetricsLayerContextType   = "metrics"
	StartTimeLayerContextType = "start_time"
	ClientIDLayerContextType  = "client_id"
)

type BaseLayer struct {
	Metrics *stats.Stats
}

func (b *BaseLayer) OnRequest(state *httransform.LayerState) error {
	clientID := b.getClientID(state)
	logger := log.WithFields(log.Fields{
		"client_id":   clientID,
		"request_id":  state.RequestID,
		"remote_addr": state.RemoteAddr,
		"method":      string(state.Request.Header.Method()),
		"url":         string(state.Request.URI().FullURI()),
	})

	state.Set(LogLayerContextType, &logger)
	state.Set(MetricsLayerContextType, b.Metrics)
	state.Set(StartTimeLayerContextType, time.Now())
	state.Set(ClientIDLayerContextType, clientID)

	logger.Info("New request")

	return nil
}

func (b *BaseLayer) OnResponse(state *httransform.LayerState, err error) {
	loggerUntyped, _ := state.Get(LogLayerContextType)
	logger := loggerUntyped.(*log.Logger)

	logger.WithFields(log.Fields{
		"response_code": state.Response.Header.StatusCode(),
	}).Info("Finish request")

	b.calculateOverallTime(state)
}

func (b *BaseLayer) calculateOverallTime(state *httransform.LayerState) {
	finishTime := time.Now()
	startTimeUntyped, _ := state.Get(StartTimeLayerContextType)

	b.Metrics.NewOverallTime(finishTime.Sub(startTimeUntyped.(time.Time)))
}

func (b *BaseLayer) getClientID(state *httransform.LayerState) string {
	host := state.RemoteAddr.String()
	if parsedHost, _, err := net.SplitHostPort(host); err == nil {
		host = parsedHost
	}

	hsh := hmac.New(sha1.New, []byte(host))
	userAgent, _ := state.RequestHeaders.GetBytes([]byte("user-agent"))
	hsh.Write(userAgent)

	return fmt.Sprintf("%x", hsh.Sum(nil))
}
