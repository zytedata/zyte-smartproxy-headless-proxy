package stats

import (
	"container/ring"
	"encoding/json"
	"sync"
	"time"

	mstats "github.com/montanaflynn/stats"
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

type timeSeries struct {
	data *ring.Ring
	lock *sync.Mutex
}

func (ts *timeSeries) MarshalJSON() ([]byte, error) {
	floats := ts.collect()
	marshalled := map[string]interface{}{
		"average":            0,
		"minimal":            0,
		"maximal":            0,
		"median":             0,
		"standard_deviation": 0,
		"percentiles":        map[uint8]float64{},
	}

	if len(floats) > 0 {
		marshalled["percentiles"] = ts.calculatePercentiles(floats)

		mean, _ := mstats.Mean(floats)
		marshalled["average"] = mean

		min, _ := mstats.Min(floats)
		marshalled["minimal"] = min

		max, _ := mstats.Max(floats)
		marshalled["maximal"] = max

		median, _ := mstats.Median(floats)
		marshalled["median"] = median

		dev, _ := mstats.StandardDeviation(floats)
		marshalled["standard_deviation"] = dev
	}

	return json.Marshal(marshalled)
}

func (ts *timeSeries) calculatePercentiles(floats []float64) map[uint8]float64 {
	percentiles := map[uint8]float64{}

	for _, perc := range statsPercentilesToCalculate {
		calculated, _ := mstats.Percentile(floats, float64(perc))
		percentiles[perc] = calculated
	}

	return percentiles
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
