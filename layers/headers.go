package layers

import "github.com/9seconds/httransform"

var headersProfileToRemove = [5]string{
	"accept",
	"accept-language",
	"dnt",
	"upgrade-insecure-requests",
	"user-agent",
}

type ZyteProxyHeadersLayer struct {
	ZyteProxyHeaders map[string]string
}

func (h *ZyteProxyHeadersLayer) OnRequest(state *httransform.LayerState) error {
	for k, v := range h.ZyteProxyHeaders {
		state.RequestHeaders.SetString(k, v)
	}

	profile, ok := state.RequestHeaders.GetString("zyte-proxy-profile")
	if ok && (profile == "desktop" || profile == "mobile") {
		for _, v := range headersProfileToRemove {
			state.RequestHeaders.DeleteString(v)
		}
	}

	return nil
}

func (h *ZyteProxyHeadersLayer) OnResponse(_ *httransform.LayerState, _ error) {
}

func NewZyteProxyHeadersLayer(zyteProxyHeaders map[string]string) httransform.Layer {
	return &ZyteProxyHeadersLayer{
		ZyteProxyHeaders: zyteProxyHeaders,
	}
}
