package layers

import (
	"github.com/9seconds/httransform"
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

func isCrawleraError(state *httransform.LayerState) bool {
	if _, ok := state.ResponseHeaders.GetString("x-crawlera-error"); ok {
		return true
	}

	return isCrawleraResponseError(state)
}

func isCrawleraResponseError(state *httransform.LayerState) bool {
	return len(state.Response.Header.Peek("X-Crawlera-Error")) > 0
}

func getClientID(state *httransform.LayerState) string {
	clientIDUntyped, _ := state.Get(clientIDLayerContextType)
	return clientIDUntyped.(string)
}

func getLogger(state *httransform.LayerState) *log.Entry {
	loggerUntyped, _ := state.Get(logLayerContextType)
	return loggerUntyped.(*log.Entry)
}

func getMetrics(state *httransform.LayerState) *stats.Stats {
	metrics, _ := state.Get(metricsLayerContextType)
	return metrics.(*stats.Stats)
}
