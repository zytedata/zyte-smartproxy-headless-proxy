package stats

import (
	"time"
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

type StatsJSON struct {
	RequestsNumber   uint64 `json:"requests_number"`
	CrawleraRequests uint64 `json:"crawlera_requests"`
	SessionsCreated  uint64 `json:"sessions_created"`
	ClientsConnected uint64 `json:"clients_connected"`
	ClientsServing   uint64 `json:"clients_serving"`

	OverallTimes  *StatsJSONTimes `json:"overall_times"`
	CrawleraTimes *StatsJSONTimes `json:"crawlera_times"`

	Uptime uint `json:"uptime"`
}

type StatsJSONTimes struct {
	Average float64 `json:"average"`
	Minimal float64 `json:"minimal"`
	Maximal float64 `json:"maxmimal"`
	Median  float64 `json:"median"`
	Perc90  float64 `json:"percentile_90"`
	StdDev  float64 `json:"standard_deviation"`
}

func (s *Stats) GetStatsJSON() *StatsJSON {
	return &StatsJSON{
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

func (s *Stats) makeJSONTimes(data *circularTimeBuffer) *StatsJSONTimes {
	jsonData := &StatsJSONTimes{}
	return jsonData
}

func (s *Stats) Collect() {
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
