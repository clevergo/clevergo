// Copyright 2016 The Gem Authors. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package middleware

import (
	"fmt"
	"strings"
	"testing"

	"github.com/go-gem/gem"
	"github.com/go-gem/tests"
	"github.com/valyala/fasthttp"
)

func TestCORS(t *testing.T) {
	m := NewCORS()
	m.Skipper = nil
	m.AllowMethods = nil
	m.AllowOrigins = nil

	m.init()

	if m.Skipper == nil {
		t.Error(errSkipperNil)
	}
	for i := 0; i < len(m.AllowOrigins); i++ {
		if m.AllowOrigins[i] != CORSAllowOrigins[i] {
			t.Error("Unexpected allow origins")
		}
	}
	for i := 0; i < len(m.AllowMethods); i++ {
		if m.AllowMethods[i] != CORSAllowMethods[i] {
			t.Error("Unexpected allow methods")
		}
	}
}

func TestCORS_Handle(t *testing.T) {
	origin1 := "http://127.0.0.1"
	origin2 := "http://127.0.0.2"
	origin3 := "http://127.0.0.3"

	m := NewCORS()
	m.AllowOrigins = []string{
		origin1, origin2,
	}
	m.AllowCredentials = true
	m.ExposeHeaders = []string{
		"X-My-Custom-Header",
		"X-Another-Custom-Header",
	}
	exposeHeadersStr := strings.Join(m.ExposeHeaders, ", ")

	router := gem.NewRouter()
	router.Use(m)
	router.GET("/", func(ctx *gem.Context) {
		ctx.HTML(fasthttp.StatusOK, "OK")
	})

	srv := gem.New("", router.Handler())

	// simple request with empty origin.
	test1 := tests.New(srv)
	test1.Expect().Status(200).Body("OK")
	if err := test1.Run(); err != nil {
		t.Error(err)
	}

	// simple request with allowed origin
	test2 := tests.New(srv)
	test2.Url = origin1
	test2.Headers[gem.HeaderOrigin] = origin2
	test2.Expect().Status(200).Custom(func(resp fasthttp.Response) error {
		origin := string(resp.Header.Peek(gem.HeaderAccessControlAllowOrigin))
		if origin != origin2 {
			return fmt.Errorf("expected %s: %q, got %q", gem.HeaderAccessControlAllowCredentials, origin2, origin)
		}
		credentials := string(resp.Header.Peek(gem.HeaderAccessControlAllowCredentials))
		if credentials != "true" {
			return fmt.Errorf("expected %s: %q, got %q", gem.HeaderAccessControlAllowCredentials, "true", credentials)
		}
		exposeHeaders := string(resp.Header.Peek(gem.HeaderAccessControlExposeHeaders))
		if exposeHeaders != exposeHeadersStr {
			return fmt.Errorf("expected %s: %q, got %q", gem.HeaderAccessControlExposeHeaders, exposeHeadersStr, exposeHeaders)
		}
		return nil
	})
	if err := test2.Run(); err != nil {
		t.Error(err)
	}

	// simple request with not allow origin
	test3 := tests.New(srv)
	test3.Url = origin1
	test3.Headers[gem.HeaderOrigin] = origin3
	test3.Expect().Status(200).Custom(func(resp fasthttp.Response) error {
		credentials := string(resp.Header.Peek(gem.HeaderAccessControlAllowCredentials))
		if credentials != "" {
			return fmt.Errorf("expected empty %s, got %q", gem.HeaderAccessControlAllowCredentials, credentials)
		}
		exposeHeaders := string(resp.Header.Peek(gem.HeaderAccessControlExposeHeaders))
		if exposeHeaders != "" {
			return fmt.Errorf("expected empty %s, got %q", gem.HeaderAccessControlExposeHeaders, exposeHeaders)
		}
		return nil
	})
	if err := test3.Run(); err != nil {
		t.Error(err)
	}

	// Always skip.
	m.Skipper = func(ctx *gem.Context) bool {
		return true
	}
	test4 := tests.New(srv)
	test4.Expect().
		Status(200).
		Body("OK")
	if err := test4.Run(); err != nil {
		t.Error(err)
	}
}

func TestCORS_Handle2(t *testing.T) {
	origin1 := "http://127.0.0.1"
	origin2 := "http://127.0.0.2"
	origin3 := "http://127.0.0.3"

	m := NewCORS()
	m.AllowOrigins = []string{
		origin1, origin2,
	}
	m.AllowCredentials = true
	m.AllowHeaders = []string{
		"X-My-Custom-Header",
		"X-Another-Custom-Header",
	}
	allowHeadersStr := strings.Join(m.AllowHeaders, ", ")
	m.MaxAge = 3600

	router := gem.NewRouter()
	router.Use(m)
	router.GET("/", func(ctx *gem.Context) {
		ctx.HTML(fasthttp.StatusOK, "OK")
	})

	srv := gem.New("", router.Handler())

	// simple request with empty origin.
	test1 := tests.New(srv)
	test1.Method = gem.MethodOptions
	test1.Expect().Status(200)
	if err := test1.Run(); err != nil {
		t.Error(err)
	}

	// simple request with allowed origin
	test2 := tests.New(srv)
	test2.Url = origin1
	test2.Method = gem.MethodOptions
	test2.Headers[gem.HeaderOrigin] = origin2
	test2.Expect().Status(200).Custom(func(resp fasthttp.Response) error {
		origin := string(resp.Header.Peek(gem.HeaderAccessControlAllowOrigin))
		if origin != origin2 {
			return fmt.Errorf("expected %s: %q, got %q", gem.HeaderAccessControlAllowCredentials, origin2, origin)
		}
		credentials := string(resp.Header.Peek(gem.HeaderAccessControlAllowCredentials))
		if credentials != "true" {
			return fmt.Errorf("expected %s: %q, got %q", gem.HeaderAccessControlAllowCredentials, "true", credentials)
		}
		maxAge := string(resp.Header.Peek(gem.HeaderAccessControlMaxAge))
		if maxAge != "3600" {
			return fmt.Errorf("expected %s: %q, got %q", gem.HeaderAccessControlMaxAge, "3600", maxAge)
		}
		allowHeaders := string(resp.Header.Peek(gem.HeaderAccessControlAllowHeaders))
		if allowHeaders != allowHeadersStr {
			return fmt.Errorf("expected %s: %q, got %q", gem.HeaderAccessControlAllowHeaders, allowHeadersStr, allowHeaders)
		}
		return nil
	})
	if err := test2.Run(); err != nil {
		t.Error(err)
	}

	// simple request with not allow origin
	test3 := tests.New(srv)
	test3.Url = origin1
	test3.Method = gem.MethodOptions
	test3.Headers[gem.HeaderOrigin] = origin3
	test3.Expect().Status(200).Custom(func(resp fasthttp.Response) error {
		credentials := string(resp.Header.Peek(gem.HeaderAccessControlAllowCredentials))
		if credentials != "" {
			return fmt.Errorf("expected empty %s, got %q", gem.HeaderAccessControlAllowCredentials, credentials)
		}
		allowHeaders := string(resp.Header.Peek(gem.HeaderAccessControlAllowHeaders))
		if allowHeaders != "" {
			return fmt.Errorf("expected empty %s, got %q", gem.HeaderAccessControlAllowHeaders, allowHeaders)
		}
		return nil
	})
	if err := test3.Run(); err != nil {
		t.Error(err)
	}

	// preflight request
	test4 := tests.New(srv)
	test4.Method = gem.MethodOptions
	test4.Url = origin1
	test4.Headers[gem.HeaderOrigin] = origin3
	if err := test4.Run(); err != nil {
		t.Error(err)
	}
}
