package stats

import (
	"time"

	mstats "github.com/montanaflynn/stats"
	log "github.com/sirupsen/logrus"
)

const (
	statsRingLength       = 2000
	statsChanBufferLength = 100
)

var statsPercentilesToCalculate = [13]int8{ // int8 to prevent marshalling to string
	10,
	20,
	30,
	40,
	50,
	60,
	70,
	75,
	80,
	85,
	90,
	95,
	99,
}

type Stats struct {
	requestsNumber   uint64
	crawleraRequests uint64
	sessionsCreated  uint64
	clientsConnected uint64
	clientsServing   uint64

	overallTimes  *circularTimeBuffer
	crawleraTimes *circularTimeBuffer

	startedAt time.Time

	RequestsNumberChan   chan struct{}
	CrawleraRequestsChan chan struct{}
	SessionsCreatedChan  chan struct{}
	ClientsConnectedChan chan bool
	ClientsServingChan   chan bool
	CrawleraTimesChan    chan time.Duration
	OverallTimesChan     chan time.Duration
}

type JSON struct {
	RequestsNumber   uint64 `json:"requests_number"`
	CrawleraRequests uint64 `json:"crawlera_requests"`
	SessionsCreated  uint64 `json:"sessions_created"`
	ClientsConnected uint64 `json:"clients_connected"`
	ClientsServing   uint64 `json:"clients_serving"`

	OverallTimes  *JSONTimes `json:"overall_times"`
	CrawleraTimes *JSONTimes `json:"crawlera_times"`

	Uptime uint `json:"uptime"`
}

type JSONTimes struct {
	Average     float64          `json:"average"`
	Minimal     float64          `json:"minimal"`
	Maximal     float64          `json:"maxmimal"`
	Median      float64          `json:"median"`
	StdDev      float64          `json:"standard_deviation"`
	Percentiles map[int8]float64 `json:"percentiles"`
}

func (s *Stats) GetStatsJSON() *JSON {
	return &JSON{
		RequestsNumber:   s.requestsNumber,
		CrawleraRequests: s.crawleraRequests,
		SessionsCreated:  s.sessionsCreated,
		ClientsConnected: s.clientsConnected,
		ClientsServing:   s.clientsServing,

		OverallTimes:  s.makeJSONTimes(s.overallTimes),
		CrawleraTimes: s.makeJSONTimes(s.crawleraTimes),

		Uptime: uint(time.Since(s.startedAt).Seconds()),
	}
}

func (s *Stats) makeJSONTimes(data *circularTimeBuffer) *JSONTimes {
	floats := data.collect()
	jsonData := &JSONTimes{
		Percentiles: s.calculatePercentiles(floats),
	}

	if len(floats) == 0 {
		return jsonData
	}

	if avg, err := mstats.Mean(floats); err == nil {
		jsonData.Average = avg
	} else {
		log.WithFields(log.Fields{
			"err": err,
		}).Info("Cannot findout mean/average value.")
	}

	if min, err := mstats.Min(floats); err == nil {
		jsonData.Minimal = min
	} else {
		log.WithFields(log.Fields{
			"err": err,
		}).Info("Cannot findout min value.")
	}

	if max, err := mstats.Max(floats); err == nil {
		jsonData.Maximal = max
	} else {
		log.WithFields(log.Fields{
			"err": err,
		}).Info("Cannot findout max value.")
	}

	if median, err := mstats.Median(floats); err == nil {
		jsonData.Median = median
	} else {
		log.WithFields(log.Fields{
			"err": err,
		}).Info("Cannot findout median value.")
	}

	if dev, err := mstats.StandardDeviation(floats); err == nil {
		jsonData.StdDev = dev
	} else {
		log.WithFields(log.Fields{
			"err": err,
		}).Info("Cannot findout standard deviation value.")
	}

	return jsonData
}

func (s *Stats) Collect() { // nolint: gocyclo
	for {
		select {
		case <-s.RequestsNumberChan:
			s.requestsNumber++
		case <-s.CrawleraRequestsChan:
			s.crawleraRequests++
		case <-s.SessionsCreatedChan:
			s.sessionsCreated++
		case clientState := <-s.ClientsConnectedChan:
			if clientState {
				s.clientsConnected++
			} else {
				s.clientsConnected--
			}
		case clientState := <-s.ClientsServingChan:
			if clientState {
				s.clientsServing++
			} else {
				s.clientsServing--
			}
		case duration := <-s.CrawleraTimesChan:
			s.crawleraTimes.add(duration)
		case duration := <-s.OverallTimesChan:
			s.overallTimes.add(duration)
		}
	}
}

func (s *Stats) calculatePercentiles(data []float64) map[int8]float64 {
	percentiles := map[int8]float64{}

	for _, perc := range statsPercentilesToCalculate {
		if calculated, err := mstats.Percentile(data, float64(perc)); err == nil {
			percentiles[perc] = calculated
		} else {
			log.WithFields(log.Fields{
				"err":        err,
				"percentile": perc,
			}).Info("Cannot findout percentile.")
		}
	}

	return percentiles
}

func NewStats() *Stats {
	return &Stats{
		overallTimes:  newCircularTimeBuffer(statsRingLength),
		crawleraTimes: newCircularTimeBuffer(statsRingLength),
		startedAt:     time.Now(),

		RequestsNumberChan:   make(chan struct{}, statsChanBufferLength),
		CrawleraRequestsChan: make(chan struct{}, statsChanBufferLength),
		SessionsCreatedChan:  make(chan struct{}, statsChanBufferLength),
		ClientsConnectedChan: make(chan bool, statsChanBufferLength),
		ClientsServingChan:   make(chan bool, statsChanBufferLength),
		CrawleraTimesChan:    make(chan time.Duration, statsChanBufferLength),
		OverallTimesChan:     make(chan time.Duration, statsChanBufferLength),
	}
}
