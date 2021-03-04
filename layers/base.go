package layers

import (
	"crypto/hmac"
	"crypto/sha1" // nolint: gosec
	"fmt"
	"net"
	"time"

	"github.com/9seconds/httransform"
	log "github.com/sirupsen/logrus"

	"github.com/scrapinghub/zyte-proxy-headless-proxy/stats"
)

const baseLayerBadStatusCode = 400

type BaseLayer struct {
	metrics *stats.Stats
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

	state.Set(logLayerContextType, logger)
	state.Set(metricsLayerContextType, b.metrics)
	state.Set(startTimeLayerContextType, time.Now())
	state.Set(clientIDLayerContextType, clientID)

	logger.Info("New request")

	return nil
}

func (b *BaseLayer) OnResponse(state *httransform.LayerState, err error) {
	logger := getLogger(state)
	metrics := getMetrics(state)

	logger.WithFields(log.Fields{
		"response_code": state.Response.Header.StatusCode(),
		"error":         err,
	}).Info("Finish request")

	switch {
	case isSmartProxyManagerError(state):
		metrics.NewSmartProxyManagerError()
	case state.Response.Header.StatusCode() >= baseLayerBadStatusCode:
		metrics.NewOtherError()
	}

	b.calculateOverallTime(state)
}

func (b *BaseLayer) calculateOverallTime(state *httransform.LayerState) {
	finishTime := time.Now()
	startTimeUntyped, _ := state.Get(startTimeLayerContextType)

	b.metrics.NewOverallTime(finishTime.Sub(startTimeUntyped.(time.Time)))
}

func (b *BaseLayer) getClientID(state *httransform.LayerState) string {
	host := state.RemoteAddr.String()
	if parsedHost, _, err := net.SplitHostPort(host); err == nil {
		host = parsedHost
	}

	hsh := hmac.New(sha1.New, []byte(host))
	userAgent, _ := state.RequestHeaders.GetBytes([]byte("user-agent"))
	hsh.Write(userAgent) // nolint: errcheck

	return fmt.Sprintf("%x", hsh.Sum(nil))
}

func NewBaseLayer(metrics *stats.Stats) httransform.Layer {
	return &BaseLayer{
		metrics: metrics,
	}
}
