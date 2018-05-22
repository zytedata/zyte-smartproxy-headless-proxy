package stats

import (
	"container/ring"
	"sync"
	"time"
)

type timeSeriesInterface interface {
	collect() []float64
}

type timeSeries struct {
	data *ring.Ring
	lock *sync.Mutex
}

func (ts *timeSeries) add(item float64) {
	ts.lock.Lock()
	defer ts.lock.Unlock()

	ts.data.Value = item
	ts.data = ts.data.Next()
}

func (ts *timeSeries) collect() []float64 {
	series := &([]float64{})

	ts.lock.Lock()
	defer ts.lock.Unlock()

	ts.data.Do(func(item interface{}) {
		if item != nil {
			*series = append(*series, item.(float64))
		}
	})

	return *series
}

type durationTimeSeries struct {
	timeSeries
}

func (d *durationTimeSeries) add(item time.Duration) {
	d.timeSeries.add(item.Seconds())
}

type uint64TimeSeries struct {
	timeSeries
}

func (u *uint64TimeSeries) add(item uint64) {
	u.timeSeries.add(float64(item))
}

func newDurationTimeSeries(capacity int) *durationTimeSeries {
	return &durationTimeSeries{
		timeSeries: timeSeries{
			data: ring.New(capacity),
			lock: &sync.Mutex{},
		},
	}
}

func newUint64TimeSeries(capacity int) *uint64TimeSeries {
	return &uint64TimeSeries{
		timeSeries: timeSeries{
			data: ring.New(capacity),
			lock: &sync.Mutex{},
		},
	}
}
