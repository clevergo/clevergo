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

func TestNewError(t *testing.T) {
	tests := []struct {
		code int
		msg  string
	}{
		{http.StatusForbidden, "forbidden"},
		{http.StatusInternalServerError, "internal server error"},
	}

	for _, test := range tests {
		err := NewError(test.code, errors.New(test.msg))
		assert.Equal(t, test.code, err.Code)
		assert.Equal(t, test.msg, err.Error())
	}
}

func TestErrorHandler(t *testing.T) {
	m := ErrorHandler()
	cases := []struct {
		err  error
		code int
		body string
	}{
		{nil, http.StatusOK, ""},
		{ErrNotFound, http.StatusNotFound, "Not Found\n"},
		{errors.New("foobar"), http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError) + "\n"},
	}
	for _, test := range cases {
		handle := m(func(c *Context) error {
			return test.err
		})
		resp := httptest.NewRecorder()
		c := newContext(resp, nil)
		c.app = Pure()
		assert.Nil(t, handle(c))
		assert.Equal(t, test.code, resp.Code)
		assert.Equal(t, test.body, resp.Body.String())
	}
}

func TestPanicErrorError(t *testing.T) {
	err := PanicError{
		Data:  "foo",
		Stack: []byte("bar"),
	}
	msg := err.Error()
	assert.Contains(t, msg, "foo")
	assert.Contains(t, msg, "bar")
}
