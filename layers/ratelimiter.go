package layers

import "github.com/9seconds/httransform/v2/layers"

type RateLimiterLayer struct {
	limiter chan struct{}
}

func (r *RateLimiterLayer) OnRequest(_ *layers.Context) error {
	r.limiter <- struct{}{}
	return nil
}

func (r *RateLimiterLayer) OnResponse(_ *layers.Context, err error) error {
	<-r.limiter
	return err
}

func NewRateLimiterLayer(concurrentConnections int) layers.Layer {
	return &RateLimiterLayer{
		limiter: make(chan struct{}, concurrentConnections),
	}
}
