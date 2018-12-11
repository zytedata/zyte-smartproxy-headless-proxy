package layers

import "github.com/9seconds/httransform"

func IsCrawleraError(state *httransform.LayerState) bool {
	_, ok := state.ResponseHeaders.GetString("x-crawlera-error")
	if ok {
		return true
	}

	return len(state.Response.Header.Peek("X-Crawlera-Error")) > 0
}
