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
	Average float64 `json:"average"`
	Minimal float64 `json:"minimal"`
	Maximal float64 `json:"maxmimal"`
	Median  float64 `json:"median"`
	Perc90  float64 `json:"percentile_90"`
	StdDev  float64 `json:"standard_deviation"`
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
	jsonData := &JSONTimes{}
	floats := data.collect()

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

	if perc, err := mstats.Percentile(floats, 90.0); err == nil {
		jsonData.Perc90 = perc
	} else {
		log.WithFields(log.Fields{
			"err": err,
		}).Info("Cannot findout 90 percentile value.")
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
