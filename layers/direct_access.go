package layers

import (
	"regexp"

	"github.com/9seconds/httransform/v2/dialers"
	"github.com/9seconds/httransform/v2/errors"
	"github.com/9seconds/httransform/v2/executor"
	"github.com/9seconds/httransform/v2/layers"
	log "github.com/sirupsen/logrus"
)

var errDirectAccess = errors.Annotate(nil, "direct access to the URL", "direct_executor", 0)

type DirectAccessLayer struct {
	rules      []*regexp.Regexp
	exceptions []*regexp.Regexp
	executor   executor.Executor
}

func (d *DirectAccessLayer) OnRequest(ctx *layers.Context) error {
	url := ctx.Request().URI()
	hostpath := make([]byte, 0, len(url.Host())+len(url.Path())+1)
	hostpath = append(hostpath, url.Host()...)
	hostpath = append(hostpath, '/')
	hostpath = append(hostpath, url.Path()...)

	for _, v := range d.exceptions {
		if v.Match(hostpath) {
			return nil
		}
	}

	for _, v := range d.rules {
		if v.Match(hostpath) {
			return errDirectAccess
		}
	}

	return nil
}

func (d *DirectAccessLayer) OnResponse(ctx *layers.Context, err error) error {
	if err == errDirectAccess {
		if err := ctx.RequestHeaders.Push(); err != nil {
			return errors.Annotate(err, "cannot sync request headers", "direct_executor", 0)
		}

		if err := d.executor(ctx); err != nil {
			return errors.Annotate(err, "cannot execute a direct request", "direct_executor", 0)
		}

		if err := ctx.ResponseHeaders.Pull(); err != nil {
			return errors.Annotate(err, "cannot read response headers", "direct_executor", 0)
		}

		logger := getLogger(ctx)
		logger.WithFields(log.Fields{}).Debug("Request was direct accessed")

		return nil
	}

	return err
}

func NewDirectAccessLayer(regexps []string, exceptRegxeps []string, proxy string) layers.Layer {
	rules := make([]*regexp.Regexp, len(regexps))
	for i, v := range regexps {
		rules[i] = regexp.MustCompile(v)
	}

	exceptions := make([]*regexp.Regexp, len(exceptRegxeps))
	for i, v := range exceptRegxeps {
		exceptions[i] = regexp.MustCompile(v)
	}

	return &DirectAccessLayer{
		rules:      rules,
		exceptions: exceptions,
		executor:   executor.MakeDefaultExecutor(createBaseDialer(proxy)),
	}
}

func createBaseDialer(proxy string) dialers.Dialer {
	if proxy != "" {
		dialer, err := dialers.DialerFromURL(dialers.Opts{}, proxy)
		if err == nil {
			return dialer
		}
	}

	return dialers.NewBase(dialers.Opts{})
}
