// Copyright 2020 CleverGo. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package clevergo

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func echoHandler(s string) Handle {
	return func(ctx *Context) error {
		ctx.WriteString(s)

		return nil
	}
}

func echoMiddleware(s string) MiddlewareFunc {
	return func(next Handle) Handle {
		return func(ctx *Context) error {
			ctx.WriteString(s + " ")
			return next(ctx)
		}
	}
}

func terminatedMiddleware() MiddlewareFunc {
	return func(next Handle) Handle {
		return func(ctx *Context) error {
			ctx.WriteString("terminated")
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
		if test.body != w.Body.String() {
			t.Errorf("expected body %q, got %q", test.body, w.Body.String())
		}
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
	m := Recovery(true)
	router := NewRouter()
	out := &bytes.Buffer{}
	log.SetOutput(out)
	router.Use(m)
	router.Get("/", func(_ *Context) error {
		panic("foobar")
	})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status code %d, got %d", http.StatusInternalServerError, w.Code)
	}
	if !strings.Contains(out.String(), "foobar") {
		t.Errorf("expected output contains %s, got %s", "foobar", out.String())
	}
}

func TestWrapH(t *testing.T) {
	handled := false
	handler := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		handled = true
	})
	m := WrapH(handler)
	m(fakeHandler("foo"))(&Context{})
	if !handled {
		t.Error("failed to wrap handler as middleware")
	}
}
