package layers

import (
	"encoding/base64"
	"fmt"

	"github.com/9seconds/httransform/v2/layers"
)

type AuthLayer struct {
	user string
}

func (h *AuthLayer) OnRequest(ctx *layers.Context) error {
	ctx.RequestHeaders.Append("proxy-authorization", fmt.Sprintf("Basic %s", h.user))
	return nil
}

func (h *AuthLayer) OnResponse(_ *layers.Context, err error) error {
	return err
}

func NewAuthLayer(user string) layers.Layer {
	encodedUser := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:", user)))

	return &AuthLayer{
		user: encodedUser,
	}
}
