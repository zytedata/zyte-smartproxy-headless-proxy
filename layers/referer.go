package layers

import (
	"net"
	"net/url"
	"strconv"
	"time"

	"github.com/9seconds/httransform/v2/layers"
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

func (r *RefererLayer) OnRequest(ctx *layers.Context) error {
	clientID := getClientID(ctx)
	host := ctx.RequestHeaders.GetLast("host").Value()
	referer := ctx.RequestHeaders.GetLast("referer").Value()

	referer = r.clean(referer)
	if referer == "" {
		referer = r.get(clientID, host)
		if referer == "" {
			referer = r.clean(string(ctx.Request().URI().FullURI()))
		}
	}

	r.set(clientID, host, referer)
	ctx.RequestHeaders.Set("Referer", referer, true)

	return nil
}

func (r *RefererLayer) OnResponse(_ *layers.Context, err error) error {
	return err
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

func NewRefererLayer() layers.Layer {
	return &RefererLayer{
		referers: ccache.Layered(ccache.Configure()),
	}
}
