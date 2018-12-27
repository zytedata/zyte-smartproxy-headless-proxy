package layers

import "github.com/9seconds/httransform"

type RateLimiterLayer struct {
	limiter chan struct{}
}

func (r *RateLimiterLayer) OnRequest(state *httransform.LayerState) error {
	r.limiter <- struct{}{}
	return nil
}

func (r *RateLimiterLayer) OnResponse(_ *httransform.LayerState, _ error) {
	<-r.limiter
}

func NewRateLimiterLayer(concurrentConnections int) httransform.Layer {
	return &RateLimiterLayer{
		limiter: make(chan struct{}, concurrentConnections),
	}
}
