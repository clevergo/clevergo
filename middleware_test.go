// Copyright 2020 CleverGo. All rights reserved.
// Use of this source code is governed by a MIT style license that can be found
// in the LICENSE file.

package clevergo

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func echoHandler(s string) Handle {
	return func(c *Context) error {
		c.WriteString(s)

		return nil
	}
}

func echoMiddleware(s string) MiddlewareFunc {
	return func(next Handle) Handle {
		return func(c *Context) error {
			c.WriteString(s + " ")
			return next(c)
		}
	}
}

func terminatedMiddleware() MiddlewareFunc {
	return func(next Handle) Handle {
		return func(c *Context) error {
			c.WriteString("terminated")
			return nil
		}
	}
}

type chainTest struct {
	handle      Handle
	middlewares []MiddlewareFunc
	body        string
}

func TestChain(t *testing.T) {
	tests := []chainTest{
		{echoHandler("foo"), []MiddlewareFunc{}, "foo"},
		{echoHandler("foo"), []MiddlewareFunc{echoMiddleware("one"), echoMiddleware("two")}, "one two foo"},
		{echoHandler("foo"), []MiddlewareFunc{echoMiddleware("one"), terminatedMiddleware()}, "one terminated"},
	}
	for _, test := range tests {
		w := httptest.NewRecorder()
		handle := Chain(test.handle, test.middlewares...)
		handle(&Context{Response: w})
		assert.Equal(t, test.body, w.Body.String())
	}
}

func ExampleChain() {
	m1 := echoMiddleware("m1")
	m2 := echoMiddleware("m2")
	handle := Chain(echoHandler("hello"), m1, m2)
	w := httptest.NewRecorder()
	handle(&Context{Response: w})
	fmt.Println(w.Body.String())
	// Output:
	// m1 m2 hello
}

func TestRecovery(t *testing.T) {
	m := Recovery()
	handle := m(func(_ *Context) error {
		panic("foobar")
	})
	w := httptest.NewRecorder()
	ctx := newContext(w, httptest.NewRequest(http.MethodGet, "/", nil))
	err := handle(ctx).(PanicError)
	assert.Equal(t, ctx, err.Context)
	assert.Equal(t, "foobar", err.Data)
	assert.NotNil(t, err.Stack)
}

func TestWrapH(t *testing.T) {
	handled := false
	handler := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		handled = true
	})
	m := WrapH(handler)
	m(fakeHandler("foo"))(&Context{})
	assert.True(t, handled, "failed to wrap handler as middleware")
}

func TestWrapHH(t *testing.T) {
	type ctxKey string
	var foo ctxKey = "foo"
	fn := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r = r.WithContext(context.WithValue(r.Context(), foo, "bar"))
			h.ServeHTTP(w, r)
		})
	}
	var handled bool
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	c := newContext(w, req)
	expectedErr := errors.New("foo")
	actualErr := WrapHH(fn)(func(c *Context) error {
		handled = true
		foo, _ := c.Value(foo).(string)
		assert.Equal(t, "bar", foo)
		return expectedErr
	})(c)
	assert.True(t, handled, "WrapHH failed")
	assert.Equal(t, expectedErr, actualErr)
}

func TestServerHeader(t *testing.T) {
	handle := func(c *Context) error {
		return nil
	}
	for _, value := range []string{"foo", "bar"} {
		m := ServerHeader(value)
		w := httptest.NewRecorder()
		m(handle)(newContext(w, nil))
		assert.Equal(t, value, w.Header().Get("Server"))
	}
}
