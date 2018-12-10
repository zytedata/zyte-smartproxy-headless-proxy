package layers

import "github.com/9seconds/httransform"

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

func (h *XHeadersLayer) OnRequest(state *httransform.LayerState) error {
	for k, v := range h.XHeaders {
		state.RequestHeaders.SetString(k, v)
	}

	profile, ok := state.RequestHeaders.GetString("x-crawlera-profile")
	if ok && (profile == "desktop" || profile == "mobile") {
		for _, v := range headersProfileToRemove {
			state.RequestHeaders.DeleteString(v)
		}
	}

	return nil
}

func (h *XHeadersLayer) OnResponse(_ *httransform.LayerState, _ error) {
}

func NewXHeadersLayer(xheaders map[string]string) httransform.Layer {
	return &XHeadersLayer{
		XHeaders: xheaders,
	}
}
