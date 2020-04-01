package stats

import (
	"container/ring"
	"encoding/json"
	"sync"
	"time"

	mstats "github.com/montanaflynn/stats"
)

// nolint: gochecknoglobals
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

type timeSeriesJSON struct {
	Average           float64           `json:"average"`
	Minimal           float64           `json:"minimal"`
	Maximal           float64           `json:"maximal"`
	Median            float64           `json:"median"`
	StandardDeviation float64           `json:"standard_deviation"`
	Percentiles       map[uint8]float64 `json:"percentiles"`
}

type timeSeries struct {
	data *ring.Ring
	lock *sync.Mutex
}

func (ts *timeSeries) MarshalJSON() ([]byte, error) {
	floats := ts.collect()
	marshalled := timeSeriesJSON{}

	if len(floats) > 0 {
		marshalled.Percentiles = ts.calculatePercentiles(floats)

		mean, _ := mstats.Mean(floats) // nolint: gosec
		marshalled.Average = mean

		min, _ := mstats.Min(floats) // nolint: gosec
		marshalled.Minimal = min

		max, _ := mstats.Max(floats) // nolint: gosec
		marshalled.Maximal = max

		median, _ := mstats.Median(floats) // nolint: gosec
		marshalled.Median = median

		dev, _ := mstats.StandardDeviation(floats) // nolint: gosec
		marshalled.StandardDeviation = dev
	}

	return json.Marshal(&marshalled)
}

func (ts *timeSeries) calculatePercentiles(floats []float64) map[uint8]float64 {
	percentiles := make(map[uint8]float64, len(statsPercentilesToCalculate))

	if len(floats) >= 100 { // nolint: gomnd
		for _, perc := range statsPercentilesToCalculate {
			calculated, _ := mstats.Percentile(floats, float64(perc)) // nolint: gosec
			percentiles[perc] = calculated
		}
	}

	return percentiles
}

func (ts *timeSeries) add(item float64) {
	ts.lock.Lock()
	ts.data.Value = item
	ts.data = ts.data.Next()
	ts.lock.Unlock()
}

func (ts *timeSeries) collect() []float64 {
	series := &([]float64{})

	ts.lock.Lock()
	ts.data.Do(func(item interface{}) {
		if item != nil {
			*series = append(*series, item.(float64))
		}
	})
	ts.lock.Unlock()

	return *series
}

type durationTimeSeries struct {
	timeSeries
}

func (d *durationTimeSeries) add(item time.Duration) {
	d.timeSeries.add(item.Seconds())
}

func newDurationTimeSeries(capacity int) *durationTimeSeries {
	return &durationTimeSeries{
		timeSeries: timeSeries{
			data: ring.New(capacity),
			lock: &sync.Mutex{},
		},
	}
}
