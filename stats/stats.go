package stats

import (
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

const (
	statsRingLength = 3000

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
	AdblockedRequests uint64 `json:"adblocked_requests"`
	CrawleraErrors    uint64 `json:"crawlera_errors"`
	AllErrors         uint64 `json:"all_errors"`

	// The owls are not what they seem
	// do not believe RWMutex. We use it as shared/exclusive lock.
	OverallTimes  *durationTimeSeries `json:"overall_times"`
	CrawleraTimes *durationTimeSeries `json:"crawlera_times"`

	Uptime statsUptime `json:"uptime"`

	statsLock *sync.RWMutex
}

func (s *Stats) NewConnection() {
	s.statsLock.RLock()
	atomic.AddUint64(&s.ClientsConnected, 1)
	atomic.AddUint64(&s.RequestsNumber, 1)
	s.statsLock.RUnlock()
}

func (s *Stats) DropConnection() {
	s.statsLock.RLock()
	atomic.AddUint64(&s.ClientsConnected, atomicDecrement)
	s.statsLock.RUnlock()
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
	s.statsLock.RLock()
	atomic.AddUint64(&s.CrawleraRequests, 1)
	s.statsLock.RUnlock()
}

func (s *Stats) NewSessionCreated() {
	s.statsLock.RLock()
	atomic.AddUint64(&s.SessionsCreated, 1)
	s.statsLock.RUnlock()
}

func (s *Stats) NewAdblockedRequest() {
	s.statsLock.RLock()
	atomic.AddUint64(&s.AdblockedRequests, 1)
	s.statsLock.RUnlock()
}

func (s *Stats) NewCrawleraError() {
	s.statsLock.RLock()
	atomic.AddUint64(&s.CrawleraErrors, 1)
	atomic.AddUint64(&s.AllErrors, 1)
	s.statsLock.RUnlock()
}

func (s *Stats) NewOtherError() {
	s.statsLock.RLock()
	atomic.AddUint64(&s.AllErrors, 1)
	s.statsLock.RUnlock()
}

func (s *Stats) NewCrawleraTime(elapsed time.Duration) {
	s.statsLock.RLock()
	s.CrawleraTimes.add(elapsed)
	s.statsLock.RUnlock()
}

func (s *Stats) NewOverallTime(elapsed time.Duration) {
	s.statsLock.RLock()
	s.OverallTimes.add(elapsed)
	s.statsLock.RUnlock()
}

// NewStats creates new initialized Stats instance.
func NewStats() *Stats {
	return &Stats{
		OverallTimes:  newDurationTimeSeries(statsRingLength),
		CrawleraTimes: newDurationTimeSeries(statsRingLength),
		Uptime:        statsUptime(time.Now()),
		statsLock:     &sync.RWMutex{},
	}
}
