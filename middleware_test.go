// Copyright 2020 CleverGo. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package clevergo

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func echoHandler(s string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, s)
	})
}

func echoMiddleware(s string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, s+" ")
			next.ServeHTTP(w, r)
		})
	}
}

func terminatedMiddleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, "terminated")
		})
	}
}

type chainTest struct {
	handler     http.Handler
	middlewares []Middleware
	body        string
}

func TestChain(t *testing.T) {
	tests := []chainTest{
		{echoHandler("foo"), []Middleware{}, "foo"},
		{echoHandler("foo"), []Middleware{echoMiddleware("one"), echoMiddleware("two")}, "one two foo"},
		{echoHandler("foo"), []Middleware{echoMiddleware("one"), terminatedMiddleware()}, "one terminated"},
	}
	for _, test := range tests {
		w := httptest.NewRecorder()
		handler := Chain(test.handler, test.middlewares...)
		handler.ServeHTTP(w, nil)
		if test.body != w.Body.String() {
			t.Errorf("expected body %q, got %q", test.body, w.Body.String())
		}
	}
}

func ExampleChain() {
	m1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, "m1 ")
			next.ServeHTTP(w, r)
		})
	}
	m2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, "m2 ")
			next.ServeHTTP(w, r)
		})
	}
	handler := echoHandler("hello")
	handler = Chain(handler, m1, m2)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, nil)
	fmt.Println(w.Body.String())
	// Output:
	// m1 m2 hello
}
