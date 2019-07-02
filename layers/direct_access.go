package layers

import (
	"regexp"

	"github.com/juju/errors"
	log "github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"

	"github.com/9seconds/httransform"
)

var errDirectAccess = errors.New("Direct access to the URL")

type DirectAccessLayer struct {
	rules    []*regexp.Regexp
	executor httransform.HTTPRequestExecutor
}

func (d *DirectAccessLayer) OnRequest(state *httransform.LayerState) error {
	url := state.Request.URI()
	hostpath := make([]byte, 0, len(url.Host())+len(url.Path())+1)
	hostpath = append(hostpath, url.Host()...)
	hostpath = append(hostpath, '/')
	hostpath = append(hostpath, url.Path()...)

	for _, v := range d.rules {
		if v.Match(hostpath) {
			return errDirectAccess
		}
	}

	return nil
}

func (d *DirectAccessLayer) OnResponse(state *httransform.LayerState, err error) {
	if err == errDirectAccess {
		httransform.HTTPExecutor(state)
		if err := httransform.ParseHeaders(state.ResponseHeaders, state.Response.Header.Header()); err != nil {
			logger := getLogger(state)
			logger.WithFields(log.Fields{"err": err}).Debug("Cannot process response")
			httransform.MakeSimpleResponse(state.Response, "Malformed response headers", fasthttp.StatusBadRequest)
		}
	}
}

func NewDirectAccessLayer(regexps []string) httransform.Layer {
	rules := make([]*regexp.Regexp, len(regexps))
	for i, v := range regexps {
		rules[i] = regexp.MustCompile(v)
	}

	return &DirectAccessLayer{
		rules:    rules,
		executor: httransform.MakeStreamingReuseHTTPClient(),
	}
}
