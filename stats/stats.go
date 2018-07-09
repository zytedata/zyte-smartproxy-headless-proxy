package stats

import (
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

const (
	statsRingLength       = 3000
	statsChanBufferLength = 100

	atomicDecrement = ^uint64(0)
)

type statsUptime time.Time

func (s statsUptime) MarshalJSON() ([]byte, error) {
	seconds := int(time.Since(time.Time(s)).Seconds())
	return []byte(strconv.Itoa(seconds)), nil
}

// Stats is a collector of statistics. Its idea is to listen to all provided
// channels and generate reports (JSON data structures).
type Stats struct {
	RequestsNumber    uint64 `json:"requests_number"`
	CrawleraRequests  uint64 `json:"crawlera_requests"`
	SessionsCreated   uint64 `json:"sessions_created"`
	ClientsConnected  uint64 `json:"clients_connected"`
	ClientsServing    uint64 `json:"clients_serving"`
	Traffic           uint64 `json:"traffic"`
	AdblockedRequests uint64 `json:"adblocked_requests"`
	CrawleraErrors    uint64 `json:"crawlera_errors"`
	AllErrors         uint64 `json:"all_errors"`

	// The owls are not what they seem
	// do not believe RWMutex. We use it as shared/exclusive lock.
	StatsLock     *sync.RWMutex       `json:"-"`
	OverallTimes  *durationTimeSeries `json:"overall_times"`
	CrawleraTimes *durationTimeSeries `json:"crawlera_times"`
	TrafficTimes  *uint64TimeSeries   `json:"traffic_times"`

	Uptime statsUptime `json:"uptime"`

	RequestsNumberChan    chan struct{}      `json:"-"`
	CrawleraRequestsChan  chan struct{}      `json:"-"`
	SessionsCreatedChan   chan struct{}      `json:"-"`
	AdblockedRequestsChan chan struct{}      `json:"-"`
	CrawleraErrorsChan    chan struct{}      `json:"-"`
	AllErrorsChan         chan struct{}      `json:"-"`
	ClientsConnectedChan  chan bool          `json:"-"`
	ClientsServingChan    chan bool          `json:"-"`
	CrawleraTimesChan     chan time.Duration `json:"-"`
	OverallTimesChan      chan time.Duration `json:"-"`
	TrafficChan           chan uint64        `json:"-"`
}

// RunCollect starts a series of collector goroutines. Each goroutine
// manages its own metric indiependently.
func (s *Stats) RunCollect() {
	go s.collectRequestNumbers()
	go s.collectCrawleraRequests()
	go s.collectCrawleraErrors()
	go s.collectAllErrors()
	go s.collectAdblockedRequests()
	go s.collectSessionsCreated()
	go s.collectClientConnected()
	go s.collectClientsServing()
	go s.collectCrawleraTimes()
	go s.collectOverallTimes()
	go s.collectTraffic()
}

func (s *Stats) collectRequestNumbers() {
	for range s.RequestsNumberChan {
		s.StatsLock.RLock()
		atomic.AddUint64(&s.RequestsNumber, 1)
		s.StatsLock.RUnlock()
	}
}

func (s *Stats) collectCrawleraRequests() {
	for range s.CrawleraRequestsChan {
		s.StatsLock.RLock()
		atomic.AddUint64(&s.CrawleraRequests, 1)
		s.StatsLock.RUnlock()
	}
}

func (s *Stats) collectSessionsCreated() {
	for range s.SessionsCreatedChan {
		s.StatsLock.RLock()
		atomic.AddUint64(&s.SessionsCreated, 1)
		s.StatsLock.RUnlock()
	}
}

func (s *Stats) collectAdblockedRequests() {
	for range s.AdblockedRequestsChan {
		s.StatsLock.RLock()
		atomic.AddUint64(&s.AdblockedRequests, 1)
		s.StatsLock.RUnlock()
	}
}

func (s *Stats) collectCrawleraErrors() {
	for range s.CrawleraErrorsChan {
		s.StatsLock.RLock()
		atomic.AddUint64(&s.CrawleraErrors, 1)
		s.StatsLock.RUnlock()
	}
}

func (s *Stats) collectAllErrors() {
	for range s.AllErrorsChan {
		s.StatsLock.RLock()
		atomic.AddUint64(&s.AllErrors, 1)
		s.StatsLock.RUnlock()
	}
}

func (s *Stats) collectClientConnected() {
	for clientConnected := range s.ClientsConnectedChan {
		s.StatsLock.RLock()
		if clientConnected {
			atomic.AddUint64(&s.ClientsConnected, 1)
		} else {
			atomic.AddUint64(&s.ClientsConnected, atomicDecrement)
		}
		s.StatsLock.RUnlock()
	}
}

func (s *Stats) collectClientsServing() {
	for clientServing := range s.ClientsServingChan {
		s.StatsLock.RLock()
		if clientServing {
			atomic.AddUint64(&s.ClientsServing, 1)
		} else {
			atomic.AddUint64(&s.ClientsServing, atomicDecrement)
		}
		s.StatsLock.RUnlock()
	}
}

func (s *Stats) collectCrawleraTimes() {
	for duration := range s.CrawleraTimesChan {
		s.StatsLock.RLock()
		s.CrawleraTimes.add(duration)
		s.StatsLock.RUnlock()
	}
}

func (s *Stats) collectOverallTimes() {
	for duration := range s.OverallTimesChan {
		s.StatsLock.RLock()
		s.OverallTimes.add(duration)
		s.StatsLock.RUnlock()
	}
}

func (s *Stats) collectTraffic() {
	for traffic := range s.TrafficChan {
		s.StatsLock.RLock()
		atomic.AddUint64(&s.Traffic, traffic)
		s.TrafficTimes.add(traffic)
		s.StatsLock.RUnlock()
	}
}

// NewStats creates new initialized Stats instance.
func NewStats() *Stats {
	return &Stats{
		OverallTimes:  newDurationTimeSeries(statsRingLength),
		CrawleraTimes: newDurationTimeSeries(statsRingLength),
		TrafficTimes:  newUint64TimeSeries(statsRingLength),
		Uptime:        statsUptime(time.Now()),
		StatsLock:     &sync.RWMutex{},

		RequestsNumberChan:    make(chan struct{}, statsChanBufferLength),
		CrawleraRequestsChan:  make(chan struct{}, statsChanBufferLength),
		CrawleraErrorsChan:    make(chan struct{}, statsChanBufferLength),
		AllErrorsChan:         make(chan struct{}, statsChanBufferLength),
		AdblockedRequestsChan: make(chan struct{}, statsChanBufferLength),
		SessionsCreatedChan:   make(chan struct{}, statsChanBufferLength),
		ClientsConnectedChan:  make(chan bool, statsChanBufferLength),
		ClientsServingChan:    make(chan bool, statsChanBufferLength),
		CrawleraTimesChan:     make(chan time.Duration, statsChanBufferLength),
		OverallTimesChan:      make(chan time.Duration, statsChanBufferLength),
		TrafficChan:           make(chan uint64, statsChanBufferLength),
	}
}
