// Copyright 2020 CleverGo. All rights reserved.
// Use of this source code is governed by a MIT style license that can be found
// in the LICENSE file.

package clevergo

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

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
		assert.Equal(t, test.err, handle(newContext(resp, httptest.NewRequest(http.MethodGet, "/", nil))))
		assert.True(t, handled)
	}
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
