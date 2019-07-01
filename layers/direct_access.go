package layers

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/juju/errors"
	log "github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"

	"github.com/9seconds/httransform"
)

var errDirectAccess = errors.New("Direct access to the URL")

type DirectAccessLayer struct {
	rule     *regexp.Regexp
	executor httransform.HTTPRequestExecutor
}

func (d *DirectAccessLayer) OnRequest(state *httransform.LayerState) error {
	if d.rule.Find(state.Request.URI().Path()) != nil {
		return errDirectAccess
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

func NewDirectAccessLayer(suffixes []string) httransform.Layer {
	quotedSuffixes := make([]string, len(suffixes))
	for i, v := range suffixes {
		quotedSuffixes[i] = regexp.QuoteMeta(v)
	}

	pipedExpression := fmt.Sprintf("(?:%s)$", strings.Join(quotedSuffixes, "|"))
	rule, err := regexp.Compile(pipedExpression)
	if err != nil {
		panic(errors.Annotate(err, "Cannot make a regular expression"))
	}

	return &DirectAccessLayer{
		rule:     rule,
		executor: httransform.MakeStreamingReuseHTTPClient(),
	}
}
