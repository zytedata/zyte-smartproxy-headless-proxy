package layers

import "github.com/9seconds/httransform"

var headersProfileToRemove = [6]string{
	"accept",
	"accept-encoding",
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

func (h *XHeadersLayer) OnResponse(state *httransform.LayerState, _ error) {

}
