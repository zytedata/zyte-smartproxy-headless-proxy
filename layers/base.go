package layers

import (
	"crypto/hmac"
	"crypto/sha1" // nolint: gosec
	"fmt"
	"net"
	"time"

	"github.com/9seconds/httransform/v2/layers"
	log "github.com/sirupsen/logrus"

	"github.com/scrapinghub/crawlera-headless-proxy/stats"
)

const baseLayerBadStatusCode = 400

type BaseLayer struct {
	metrics *stats.Stats
}

func (b *BaseLayer) OnRequest(ctx *layers.Context) error {
	clientID := b.getClientID(ctx)
	logger := log.WithFields(log.Fields{
		"client_id":   clientID,
		"request_id":  ctx.RequestID,
		"remote_addr": ctx.RemoteAddr(),
		"method":      string(ctx.Request().Header.Method()),
		"url":         string(ctx.Request().URI().FullURI()),
	})

	ctx.Set(logLayerContextType, logger)
	ctx.Set(metricsLayerContextType, b.metrics)
	ctx.Set(startTimeLayerContextType, time.Now())
	ctx.Set(clientIDLayerContextType, clientID)

	logger.Info("New request")

	return nil
}

func (b *BaseLayer) OnResponse(ctx *layers.Context, err error) error {
	logger := getLogger(ctx)
	metrics := getMetrics(ctx)

	logger.WithFields(log.Fields{
		"response_code": ctx.Response().StatusCode(),
		"error":         err,
	}).Info("Finish request")

	switch {
	case isCrawleraError(ctx):
		metrics.NewCrawleraError()
	case ctx.Response().StatusCode() >= baseLayerBadStatusCode:
		metrics.NewOtherError()
	}

	b.calculateOverallTime(ctx)

	return err
}

func (b *BaseLayer) calculateOverallTime(ctx *layers.Context) {
	finishTime := time.Now()
	startTimeUntyped := ctx.Get(startTimeLayerContextType)

	b.metrics.NewOverallTime(finishTime.Sub(startTimeUntyped.(time.Time)))
}

func (b *BaseLayer) getClientID(ctx *layers.Context) string {
	host := ctx.RemoteAddr().String()
	if parsedHost, _, err := net.SplitHostPort(host); err == nil {
		host = parsedHost
	}

	hsh := hmac.New(sha1.New, []byte(host))
	userAgent := ctx.RequestHeaders.GetLast("user-agent").Value()
	hsh.Write([]byte(userAgent)) // nolint: errcheck

	return fmt.Sprintf("%x", hsh.Sum(nil))
}

func NewBaseLayer(metrics *stats.Stats) layers.Layer {
	return &BaseLayer{
		metrics: metrics,
	}
}
