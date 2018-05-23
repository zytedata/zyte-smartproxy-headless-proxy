package stats

import (
	"sync"
	"sync/atomic"
	"time"

	mstats "github.com/montanaflynn/stats"
	log "github.com/sirupsen/logrus"
)

const (
	statsRingLength       = 3000
	statsChanBufferLength = 100

	atomicDecrement = ^uint64(0)
)

var statsPercentilesToCalculate = [13]uint8{
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

// Stats is a collector of statistics. Its idea is to listen to all provided
// channels and generate reports (JSON data structures).
type Stats struct {
	requestsNumber   uint64
	crawleraRequests uint64
	sessionsCreated  uint64
	clientsConnected uint64
	clientsServing   uint64
	traffic          uint64

	// The owls are not what they seem
	// do not believe RWMutex. We use it as shared/exclusive lock.
	statsLock     *sync.RWMutex
	overallTimes  *durationTimeSeries
	crawleraTimes *durationTimeSeries
	trafficTimes  *uint64TimeSeries

	startedAt time.Time

	RequestsNumberChan   chan struct{}
	CrawleraRequestsChan chan struct{}
	SessionsCreatedChan  chan struct{}
	ClientsConnectedChan chan bool
	ClientsServingChan   chan bool
	CrawleraTimesChan    chan time.Duration
	OverallTimesChan     chan time.Duration
	TrafficChan          chan uint64
}

// JSON is intended to be serialized to JSON by proxy API.
type JSON struct {
	RequestsNumber   uint64 `json:"requests_number"`
	CrawleraRequests uint64 `json:"crawlera_requests"`
	SessionsCreated  uint64 `json:"sessions_created"`
	ClientsConnected uint64 `json:"clients_connected"`
	ClientsServing   uint64 `json:"clients_serving"`
	Traffic          uint64 `json:"traffic"`

	OverallTimes  *JSONTimes `json:"overall_times"`
	CrawleraTimes *JSONTimes `json:"crawlera_times"`
	TrafficTimes  *JSONTimes `json:"traffic_times"`

	Uptime uint `json:"uptime"`
}

// JSONTimes contains statistics on time series.
type JSONTimes struct {
	Average     float64           `json:"average"`
	Minimal     float64           `json:"minimal"`
	Maximal     float64           `json:"maxmimal"`
	Median      float64           `json:"median"`
	StdDev      float64           `json:"standard_deviation"`
	Percentiles map[uint8]float64 `json:"percentiles"`
}

// GetStatsJSON generates JSON structure from Stats. This is to be
// serialized in proxy API.
func (s *Stats) GetStatsJSON() *JSON {
	s.statsLock.Lock()
	defer s.statsLock.Unlock()

	return &JSON{
		RequestsNumber:   s.requestsNumber,
		CrawleraRequests: s.crawleraRequests,
		SessionsCreated:  s.sessionsCreated,
		ClientsConnected: s.clientsConnected,
		ClientsServing:   s.clientsServing,
		Traffic:          s.traffic,

		OverallTimes:  s.makeJSONTimes(s.overallTimes),
		CrawleraTimes: s.makeJSONTimes(s.crawleraTimes),
		TrafficTimes:  s.makeJSONTimes(s.trafficTimes),

		Uptime: uint(time.Since(s.startedAt).Seconds()),
	}
}

func (s *Stats) makeJSONTimes(data timeSeriesInterface) *JSONTimes {
	jsonData := &JSONTimes{
		Percentiles: map[uint8]float64{},
	}
	floats := data.collect()

	if len(floats) == 0 {
		return jsonData
	}
	jsonData.Percentiles = s.calculatePercentiles(floats)

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

// RunCollect starts a series of collector goroutines. Each goroutine
// manages its own metric indiependently.
func (s *Stats) RunCollect() {
	go s.collectRequestNumbers()
	go s.collectCrawleraRequests()
	go s.collectSessionsCreated()
	go s.collectClientConnected()
	go s.collectClientsServing()
	go s.collectCrawleraTimes()
	go s.collectOverallTimes()
	go s.collectTraffic()
}

func (s *Stats) collectRequestNumbers() {
	for {
		<-s.RequestsNumberChan

		s.statsLock.RLock()
		atomic.AddUint64(&s.requestsNumber, 1)
		s.statsLock.RUnlock()
	}
}

func (s *Stats) collectCrawleraRequests() {
	for {
		<-s.CrawleraRequestsChan

		s.statsLock.RLock()
		atomic.AddUint64(&s.crawleraRequests, 1)
		s.statsLock.RUnlock()
	}
}

func (s *Stats) collectSessionsCreated() {
	for {
		<-s.SessionsCreatedChan

		s.statsLock.RLock()
		atomic.AddUint64(&s.sessionsCreated, 1)
		s.statsLock.RUnlock()
	}
}

func (s *Stats) collectClientConnected() {
	for {
		clientConnected := <-s.ClientsConnectedChan

		s.statsLock.RLock()
		if clientConnected {
			atomic.AddUint64(&s.clientsConnected, 1)
		} else {
			atomic.AddUint64(&s.clientsConnected, atomicDecrement)
		}
		s.statsLock.RUnlock()
	}
}

func (s *Stats) collectClientsServing() {
	for {
		clientServing := <-s.ClientsServingChan

		s.statsLock.RLock()
		if clientServing {
			atomic.AddUint64(&s.clientsServing, 1)
		} else {
			atomic.AddUint64(&s.clientsServing, atomicDecrement)
		}
		s.statsLock.RUnlock()
	}
}

func (s *Stats) collectCrawleraTimes() {
	for {
		duration := <-s.CrawleraTimesChan

		s.statsLock.RLock()
		s.crawleraTimes.add(duration)
		s.statsLock.RUnlock()
	}
}

func (s *Stats) collectOverallTimes() {
	for {
		duration := <-s.OverallTimesChan

		s.statsLock.RLock()
		s.overallTimes.add(duration)
		s.statsLock.RUnlock()
	}
}

func (s *Stats) collectTraffic() {
	for {
		traffic := <-s.TrafficChan

		s.statsLock.RLock()
		atomic.AddUint64(&s.traffic, traffic)
		s.trafficTimes.add(traffic)
		s.statsLock.RUnlock()
	}
}

func (s *Stats) calculatePercentiles(data []float64) map[uint8]float64 {
	percentiles := map[uint8]float64{}

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

// NewStats creates new initialized Stats instance.
func NewStats() *Stats {
	return &Stats{
		overallTimes:  newDurationTimeSeries(statsRingLength),
		crawleraTimes: newDurationTimeSeries(statsRingLength),
		trafficTimes:  newUint64TimeSeries(statsRingLength),
		startedAt:     time.Now(),
		statsLock:     &sync.RWMutex{},

		RequestsNumberChan:   make(chan struct{}, statsChanBufferLength),
		CrawleraRequestsChan: make(chan struct{}, statsChanBufferLength),
		SessionsCreatedChan:  make(chan struct{}, statsChanBufferLength),
		ClientsConnectedChan: make(chan bool, statsChanBufferLength),
		ClientsServingChan:   make(chan bool, statsChanBufferLength),
		CrawleraTimesChan:    make(chan time.Duration, statsChanBufferLength),
		OverallTimesChan:     make(chan time.Duration, statsChanBufferLength),
		TrafficChan:          make(chan uint64, statsChanBufferLength),
	}
}
