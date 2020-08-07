// Copyright 2020 CleverGo. All rights reserved.
// Use of this source code is governed by a MIT style license that can be found
// in the LICENSE file.

package clevergo

import (
	"bytes"
	"errors"
	stdlog "log"
	"net/http"
	"net/http/httptest"
	"testing"

	"clevergo.tech/log"
	"github.com/stretchr/testify/assert"
)

func TestLoggingLogger(t *testing.T) {
	l := &logging{}
	assert.Nil(t, l.logger)
	LoggingLogger(defaultLogger)(l)
	assert.Equal(t, defaultLogger, l.logger)
}

func TestLogging(t *testing.T) {
	m := Logging(LoggingLogger(defaultLogger))
	cases := []struct {
		err error
	}{
		{nil},
		{ErrNotFound},
		{errors.New("foobar")},
	}
	for _, test := range cases {
		handled := true
		handle := m(func(c *Context) error {
			handled = true
			c.WriteHeader(http.StatusOK)
			return test.err
		})
		resp := httptest.NewRecorder()
		c := newContext(resp, httptest.NewRequest(http.MethodGet, "/", nil))
		c.app = Pure()
		assert.Equal(t, test.err, handle(c))
		assert.True(t, handled)
	}
}

func TestBufferedResponseWrite(t *testing.T) {
	data := []byte("foobar")
	w := &bufferedResponse{}
	w.Write(data)
	assert.Equal(t, data, w.buf.Bytes())
}

func TestBufferedResponseWriteString(t *testing.T) {
	data := "foobar"
	w := &bufferedResponse{}
	w.WriteString(data)
	assert.Equal(t, data, w.buf.String())
}

func TestBufferedResponse(t *testing.T) {
	w := httptest.NewRecorder()
	resp := newBufferedResponse(w)
	assert.Equal(t, w, resp.ResponseWriter)
	assert.Equal(t, http.StatusOK, resp.statusCode)
	assert.False(t, resp.wroteHeader)
}

func TestBufferedResponseWriteHeader(t *testing.T) {
	w := httptest.NewRecorder()
	resp := newBufferedResponse(w)
	resp.WriteHeader(http.StatusNotFound)
	assert.Equal(t, http.StatusNotFound, resp.statusCode)
	assert.True(t, resp.wroteHeader)

	resp.WriteHeader(http.StatusOK)
	assert.Equal(t, http.StatusNotFound, resp.statusCode)
}

type nullWriter struct {
	err error
}

func (*nullWriter) Header() http.Header {
	return http.Header{}
}

func (*nullWriter) WriteHeader(statusCode int) {
}

func (w *nullWriter) Write(p []byte) (int, error) {
	return 0, w.err
}

func TestBufferedResponseEmit(t *testing.T) {
	output := &bytes.Buffer{}
	expectedErr := errors.New("failed to write response")
	w := &nullWriter{expectedErr}
	c := newContext(w, httptest.NewRequest(http.MethodGet, "/", nil))
	c.app = Pure()
	c.app.Logger = log.New(output, "", stdlog.LstdFlags)
	Logging()(fakeHandler("buffered response test"))(c)
	assert.Contains(t, output.String(), expectedErr.Error())
}
