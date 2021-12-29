package layers

import "github.com/9seconds/httransform/v2/layers"

var headersProfileToRemove = [5]string{
	"accept",
	"accept-language",
	"dnt",
	"upgrade-insecure-requests",
	"user-agent",
}

type XHeadersLayer struct {
	XHeaders map[string]string
}

func (h *XHeadersLayer) OnRequest(ctx *layers.Context) error {
	for k, v := range h.XHeaders {
		ctx.RequestHeaders.Set(k, v, true)
	}

	profile := ctx.RequestHeaders.GetLast("x-crawlera-profile").Value()
	switch profile {
	case "desktop", "mobile":
		for _, v := range headersProfileToRemove {
			ctx.RequestHeaders.Remove(v)
		}
	}

	return nil
}

func (h *XHeadersLayer) OnResponse(_ *layers.Context, err error) error {
	return err
}

func NewXHeadersLayer(xheaders map[string]string) layers.Layer {
	return &XHeadersLayer{
		XHeaders: xheaders,
	}
}
