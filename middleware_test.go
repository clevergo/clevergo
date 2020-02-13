// Copyright 2020 CleverGo. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package clevergo

import (
	"errors"
	"fmt"
	"net/http/httptest"
	"testing"
)

func echoHandler(s string) Handle {
	return func(ctx *Context) error {
		ctx.WriteString(s)

		return nil
	}
}

func echoMiddleware(s string) MiddlewareFunc {
	return func(ctx *Context) error {
		ctx.WriteString(s + " ")
		return nil
	}
}

func terminatedMiddleware() MiddlewareFunc {
	return func(ctx *Context) error {
		ctx.WriteString("terminated")
		return errors.New("terminated")
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
	m1 := func(ctx *Context) error {
		ctx.WriteString("m1 ")
		return nil
	}
	m2 := func(ctx *Context) error {
		ctx.WriteString("m2 ")
		return nil
	}
	handle := Chain(echoHandler("hello"), m1, m2)
	w := httptest.NewRecorder()
	handle(&Context{Response: w})
	fmt.Println(w.Body.String())
	// Output:
	// m1 m2 hello
}
