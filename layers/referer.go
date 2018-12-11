package layers

import (
	"net"
	"net/url"
	"strconv"
	"time"

	"github.com/9seconds/httransform"
	"github.com/karlseguin/ccache"
)

const (
	refererLayerTTL  = 10 * time.Second
	defaultHTTPPort  = 80
	defaultHTTPSPort = 443
)

type RefererLayer struct {
	referers *ccache.LayeredCache
}

func (r *RefererLayer) OnRequest(state *httransform.LayerState) error {
	clientID := getClientID(state)
	host, _ := state.RequestHeaders.GetString("host")
	referer, _ := state.RequestHeaders.GetString("referer")

	referer = r.clean(referer)
	if referer == "" {
		referer = r.get(clientID, host)
		if referer == "" {
			referer = r.clean(string(state.Request.URI().FullURI()))
		}
	}

	r.set(clientID, host, referer)
	state.RequestHeaders.SetString("Referer", referer)

	return nil
}

func (r *RefererLayer) OnResponse(_ *httransform.LayerState, _ error) {
}

func (r *RefererLayer) clean(referer string) string {
	urlCopy, err := url.Parse(referer)
	if err != nil {
		return referer
	}

	urlCopy.Fragment = ""
	urlCopy.RawQuery = ""
	urlCopy.ForceQuery = false
	urlCopy.Host = r.prepareHost(urlCopy)

	return urlCopy.String()
}

func (r *RefererLayer) prepareHost(data *url.URL) string {
	host, portStr, err := net.SplitHostPort(data.Host)
	if err != nil {
		return data.Host
	}

	if port, err := strconv.Atoi(portStr); err == nil {
		if (data.Scheme == "http" && port == defaultHTTPPort) || (data.Scheme == "https" && port == defaultHTTPSPort) {
			return host
		}
	}

	return data.Host
}

func (r *RefererLayer) get(primary, secondary string) string {
	if item := r.referers.Get(primary, secondary); item != nil && !item.Expired() {
		return item.Value().(string)
	}

	return ""
}

func (r *RefererLayer) set(primary, secondary, item string) {
	r.referers.Set(primary, secondary, item, refererLayerTTL)
}

func NewRefererLayer() httransform.Layer {
	return &RefererLayer{
		referers: ccache.Layered(ccache.Configure()),
	}
}
