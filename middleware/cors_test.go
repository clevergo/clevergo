// Copyright 2016 The Gem Authors. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package middleware

import (
	"testing"

	"github.com/go-gem/gem"
	"github.com/go-gem/tests"
	"github.com/valyala/fasthttp"
)

func TestCORS(t *testing.T) {
	m := NewCORS()
	m.Skipper = nil
	m.AllowMethods = []string{}
	m.AllowOrigins = []string{}

	router := gem.NewRouter()
	router.Use(m)
	router.GET("/", func(ctx *gem.Context) {
		ctx.HTML(fasthttp.StatusOK, "OK")
	})

	if m.Skipper == nil {
		t.Errorf("The skipper should not be nil")
	}
	for i := 0; i < len(m.AllowOrigins); i++ {
		if m.AllowOrigins[i] != CORSAllowOrigins[i] {
			t.Errorf("Unexpected allow origins")
		}
	}
	for i := 0; i < len(m.AllowMethods); i++ {
		if m.AllowMethods[i] != CORSAllowMethods[i] {
			t.Errorf("Unexpected allow methods")
		}
	}

	srv := gem.New("", router.Handler)

	test1 := tests.New(srv)
	test1.Expect().Status(200).Body("OK")
	if err := test1.Run(); err != nil {
		t.Error(err)
	}

	// Preflight request
	test2 := tests.New(srv)
	test2.Method = gem.MethodOptions
	test2.Expect().Status(200)
	if err := test2.Run(); err != nil {
		t.Error(err)
	}

	// Always skip.
	m.Skipper = func(ctx *gem.Context) bool {
		return true
	}
	test8 := tests.New(srv)
	test8.Expect().
		Status(200).
		Body("OK")
	if err := test8.Run(); err != nil {
		t.Error(err)
	}
}
