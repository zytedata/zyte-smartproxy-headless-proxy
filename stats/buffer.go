package stats

import (
	"container/ring"
	"sync"
	"time"
)

type circularTimeBuffer struct {
	data *ring.Ring
	lock *sync.Mutex
}

func (cb *circularTimeBuffer) add(item time.Duration) {
	cb.lock.Lock()
	defer cb.lock.Unlock()

	cb.data.Value = float64(item)
	cb.data = cb.data.Next()
}

func (cb *circularTimeBuffer) collect() []float64 {
	durations := &([]float64{})

	cb.lock.Lock()
	defer cb.lock.Unlock()

	cb.data.Do(func(item interface{}) {
		if item != nil {
			*durations = append(*durations, item.(float64))
		}
	})

	return *durations
}

func newCircularTimeBuffer(capacity int) *circularTimeBuffer {
	return &circularTimeBuffer{
		data: ring.New(capacity),
		lock: &sync.Mutex{},
	}
}
