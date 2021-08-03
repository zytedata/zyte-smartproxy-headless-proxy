package layers

import (
	"errors"
	"regexp"

	log "github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"

	"github.com/9seconds/httransform/v2/layers"
	"github.com/9seconds/httransform/v2/executor"
)

var errDirectAccess = errors.New("direct access to the URL")

type DirectAccessLayer struct {
	rules    []*regexp.Regexp
	executor executor.Executor
}

func (d *DirectAccessLayer) OnRequest(ctx *layers.Context) error {
	url := ctx.Request().URI()
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

func (d *DirectAccessLayer) OnResponse(ctx *layers.Context, err error) error {
	if err == errDirectAccess {
		/*httransform.HTTPExecutor(state)

		if err := httransform.ParseHeaders(state.ResponseHeaders, state.Response.Header.Header()); err != nil {
			logger := getLogger(state)
			logger.WithFields(log.Fields{"err": err}).Debug("Cannot process response")
			httransform.MakeSimpleResponse(state.Response, "Malformed response headers", fasthttp.StatusBadRequest)
		}*/
	}
	return err
}

func NewDirectAccessLayer(regexps []string) layers.Layer {
	rules := make([]*regexp.Regexp, len(regexps))
	for i, v := range regexps {
		rules[i] = regexp.MustCompile(v)
	}

	return &DirectAccessLayer{
		rules:    rules,
		//executor: httransform.MakeStreamingReuseHTTPClient(),
	}
}
