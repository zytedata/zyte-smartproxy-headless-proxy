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
	IngressTraffic    uint64 `json:"ingress_traffic"`
	EgressTraffic     uint64 `json:"egress_traffic"`
	AdblockedRequests uint64 `json:"adblocked_requests"`
	CrawleraErrors    uint64 `json:"crawlera_errors"`
	AllErrors         uint64 `json:"all_errors"`

	// The owls are not what they seem
	// do not believe RWMutex. We use it as shared/exclusive lock.
	StatsLock           *sync.RWMutex       `json:"-"`
	OverallTimes        *durationTimeSeries `json:"overall_times"`
	CrawleraTimes       *durationTimeSeries `json:"crawlera_times"`
	IngressTrafficTimes *uint64TimeSeries   `json:"ingress_traffic_times"`
	EgressTrafficTimes  *uint64TimeSeries   `json:"egress_traffic_times"`

	Uptime statsUptime `json:"uptime"`
}

func (s *Stats) NewConnection() {
	s.StatsLock.RLock()
	atomic.AddUint64(&s.ClientsConnected, 1)
	atomic.AddUint64(&s.RequestsNumber, 1)
	s.StatsLock.RUnlock()
}

func (s *Stats) DropConnection() {
	s.StatsLock.RLock()
	atomic.AddUint64(&s.ClientsConnected, atomicDecrement)
	s.StatsLock.RUnlock()
}

func (s *Stats) NewGet() {
}

func (s *Stats) NewHead() {
}

func (s *Stats) NewPost() {
}

func (s *Stats) NewPut() {
}

func (s *Stats) NewDelete() {
}

func (s *Stats) NewConnect() {
}

func (s *Stats) NewOptions() {
}

func (s *Stats) NewTrace() {
}

func (s *Stats) NewPatch() {
}

func (s *Stats) NewOther() {
}

func (s *Stats) DropGet() {
}

func (s *Stats) DropHead() {
}

func (s *Stats) DropPost() {
}

func (s *Stats) DropPut() {
}

func (s *Stats) DropDelete() {
}

func (s *Stats) DropConnect() {
}

func (s *Stats) DropOptions() {
}

func (s *Stats) DropTrace() {
}

func (s *Stats) DropPatch() {
}

func (s *Stats) DropOther() {
}

func (s *Stats) NewCertificate() {
}

func (s *Stats) DropCertificate() {
}

func (s *Stats) NewCrawleraRequest() {
	s.StatsLock.RLock()
	atomic.AddUint64(&s.CrawleraRequests, 1)
	s.StatsLock.RUnlock()
}

func (s *Stats) NewSessionCreated() {
	s.StatsLock.RLock()
	atomic.AddUint64(&s.SessionsCreated, 1)
	s.StatsLock.RUnlock()
}

func (s *Stats) NewAdblockedRequest() {
	s.StatsLock.RLock()
	atomic.AddUint64(&s.AdblockedRequests, 1)
	s.StatsLock.RUnlock()
}

func (s *Stats) NewCrawleraError() {
	s.StatsLock.RLock()
	atomic.AddUint64(&s.CrawleraErrors, 1)
	atomic.AddUint64(&s.AllErrors, 1)
	s.StatsLock.RUnlock()
}

func (s *Stats) NewOtherError() {
	s.StatsLock.RLock()
	atomic.AddUint64(&s.AllErrors, 1)
	s.StatsLock.RUnlock()
}

func (s *Stats) NewCrawleraTime(elapsed time.Duration) {
	s.StatsLock.RLock()
	s.CrawleraTimes.add(elapsed)
	s.StatsLock.RUnlock()
}

func (s *Stats) NewOverallTime(elapsed time.Duration) {
	s.StatsLock.RLock()
	s.OverallTimes.add(elapsed)
	s.StatsLock.RUnlock()
}

func (s *Stats) NewIngressTraffic(size int) {
	s.StatsLock.RLock()
	atomic.AddUint64(&s.IngressTraffic, uint64(size))
	s.IngressTrafficTimes.add(uint64(size))
	s.StatsLock.RUnlock()
}

func (s *Stats) NewEgressTraffic(size int) {
	s.StatsLock.RLock()
	atomic.AddUint64(&s.EgressTraffic, uint64(size))
	s.EgressTrafficTimes.add(uint64(size))
	s.StatsLock.RUnlock()
}

// NewStats creates new initialized Stats instance.
func NewStats() *Stats {
	return &Stats{
		OverallTimes:        newDurationTimeSeries(statsRingLength),
		CrawleraTimes:       newDurationTimeSeries(statsRingLength),
		IngressTrafficTimes: newUint64TimeSeries(statsRingLength),
		EgressTrafficTimes:  newUint64TimeSeries(statsRingLength),
		Uptime:              statsUptime(time.Now()),
		StatsLock:           &sync.RWMutex{},
	}
}
