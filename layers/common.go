package layers

import (
	"github.com/9seconds/httransform/v2/layers"
	log "github.com/sirupsen/logrus"

	"github.com/scrapinghub/crawlera-headless-proxy/stats"
)

const (
	logLayerContextType       = "log"
	metricsLayerContextType   = "metrics"
	startTimeLayerContextType = "start_time"
	clientIDLayerContextType  = "client_id"
	sessionChanContextType    = "session_chan"
)

func isCrawleraError(ctx *layers.Context) bool {
	if ctx.ResponseHeaders.GetLast("x-crawlera-error") != nil {
		return true
	}

	return isCrawleraResponseError(ctx)
}

func isCrawleraResponseError(ctx *layers.Context) bool {
	return len(ctx.ResponseHeaders.GetLast("X-Crawlera-Error").Value()) > 0
}

func getClientID(ctx *layers.Context) string {
	clientIDUntyped := ctx.Get(clientIDLayerContextType)
	return clientIDUntyped.(string)
}

func getLogger(ctx *layers.Context) *log.Entry {
	loggerUntyped := ctx.Get(logLayerContextType)
	return loggerUntyped.(*log.Entry)
}

func getMetrics(ctx *layers.Context) *stats.Stats {
	metrics := ctx.Get(metricsLayerContextType)
	return metrics.(*stats.Stats)
}
