package middleware

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ClosingBuffer struct {
	bytes.Buffer
}

func (c *ClosingBuffer) Close() error {
	return nil
}

type TrafficCounterSuite struct {
	suite.Suite
}

func (t *TrafficCounterSuite) TestWork() {
	buf := *bytes.NewBufferString("testtesttest")
	cb := ClosingBuffer{Buffer: buf}

	resp := &http.Response{Body: &cb, Header: http.Header{}}
	resp.Header.Set("Expires", "Forever")

	channel := make(chan uint64, 1)
	tc := newTrafficCounter(resp, channel)
	io.Copy(ioutil.Discard, tc) // nolint: errcheck

	t.Nil(tc.Close())
	data := <-channel
	t.Equal(int(data), 26)
}

func TestTrafficCounter(t *testing.T) {
	suite.Run(t, &TrafficCounterSuite{})
}
