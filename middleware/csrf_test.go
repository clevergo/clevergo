// Copyright 2016 The Gem Authors. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package middleware

import (
	"fmt"
	"testing"

	"github.com/go-gem/gem"
	"github.com/go-gem/tests"
	"github.com/valyala/fasthttp"
)

func TestCSRF(t *testing.T) {
	m := NewCSRF()

	var encodedToken string
	cookie := &fasthttp.Cookie{}
	cookie.SetKey(m.CookieKey)

	router := gem.NewRouter()
	router.Use(m)
	router.GET("/", func(ctx *gem.Context) {
		encodedToken, _ = ctx.UserValue(m.ContextKey).(string)
		ctx.Response.Header.Cookie(cookie)

		ctx.HTML(fasthttp.StatusOK, fasthttp.StatusMessage(fasthttp.StatusOK))
	})
	router.POST("/", func(ctx *gem.Context) {
		ctx.HTML(fasthttp.StatusOK, fasthttp.StatusMessage(fasthttp.StatusOK))
	})

	srv := gem.New("", router.Handler)

	var err error

	// safe method.
	test1 := tests.New(srv)
	test1.Expect().Status(fasthttp.StatusOK).Body(fasthttp.StatusMessage(fasthttp.StatusOK))
	if err = test1.Run(); err != nil {
		t.Error(err)
	}

	// unsafe method with empty token.
	test2 := tests.New(srv)
	test2.Method = gem.MethodPost
	test2.Expect().Status(fasthttp.StatusBadRequest).Body("Unable to verify your data submission.")
	if err = test2.Run(); err != nil {
		t.Error(err)
	}

	trueToken := string(cookie.Value())

	// unsafe method with invalid token(base64 error).
	test3 := tests.New(srv)
	test3.Method = gem.MethodPost
	test3.Headers["Cookie"] = fmt.Sprintf("%s=%s", m.CookieKey, trueToken)
	test3.Url = fmt.Sprintf("/?%s=%s", m.FormKey, "INVALID-TOKEN")
	test3.Expect().Status(fasthttp.StatusBadRequest).Body("Unable to verify your data submission.")
	if err = test3.Run(); err != nil {
		t.Error(err)
	}

	// unsafe method with invalid token(length).
	test4 := tests.New(srv)
	test4.Method = gem.MethodPost
	test4.Headers["Cookie"] = fmt.Sprintf("%s=%s", m.CookieKey, trueToken)
	test4.Url = fmt.Sprintf("/?%s=%s", m.FormKey, "INVALIDTOKEN")
	test4.Expect().Status(fasthttp.StatusBadRequest).Body("Unable to verify your data submission.")
	if err = test4.Run(); err != nil {
		t.Error(err)
	}

	// unsafe method with invalid token.
	test5 := tests.New(srv)
	test5.Method = gem.MethodPost
	test5.Headers["Cookie"] = fmt.Sprintf("%s=%s", m.CookieKey, trueToken)
	test5.Url = fmt.Sprintf("/?%s=%s", m.FormKey, "0fBsg5gwl21YKNbvvMXVBx8OzEO8iBohZBT5C3cYhL/ky2WplJQn.A==")
	test5.Expect().Status(fasthttp.StatusBadRequest).Body("Unable to verify your data submission.")
	if err = test5.Run(); err != nil {
		t.Error(err)
	}

	// unsafe method with valid token.

	test6 := tests.New(srv)
	test6.Method = gem.MethodPost
	test6.Headers["Cookie"] = fmt.Sprintf("%s=%s", m.CookieKey, trueToken)
	test6.Url = fmt.Sprintf("/?%s=%s", m.FormKey, encodedToken)
	test6.Expect().Status(fasthttp.StatusOK).Body(fasthttp.StatusMessage(fasthttp.StatusOK))
	if err = test6.Run(); err != nil {
		t.Error(err)
	}

	// Always skip.
	m.Skipper = alwaysSkipper
	test7 := tests.New(srv)
	test7.Expect().Status(fasthttp.StatusOK).Body(fasthttp.StatusMessage(fasthttp.StatusOK))
	if err = test7.Run(); err != nil {
		t.Error(err)
	}
}

func TestCSRF_Handle(t *testing.T) {
	m := CSRF{
		CookieKey:     "",
		CookieOptions: nil,
		HeaderKey:     "",
		FormKey:       "",
		ContextKey:    "",
		MaskLen:       0,
		TokenLen:      0,
		Skipper:       nil,
	}

	var next gem.Handler
	_ = m.Handle(next)

	if m.CookieKey != CSRFCookieKey {
		t.Errorf("expected CookieKey: %v, got %v", CSRFCookieKey, m.CookieKey)
	}
	if m.CookieOptions != CSRFCookieOptions {
		t.Errorf("expected CookieOptions: %v, got %v", CSRFCookieOptions, m.CookieOptions)
	}
	if m.HeaderKey != CSRFHeaderKey {
		t.Errorf("expected HeaderKey: %v, got %v", CSRFHeaderKey, m.HeaderKey)
	}
	if m.FormKey != CSRFFormKey {
		t.Errorf("expected FormKey: %v, got %v", CSRFFormKey, m.FormKey)
	}
	if m.ContextKey != CSRFContextKey {
		t.Errorf("expected ContextKey: %v, got %v", CSRFContextKey, m.ContextKey)
	}
	if m.MaskLen != CSRFMaskLen {
		t.Errorf("expected MaskLen: %v, got %v", CSRFMaskLen, m.MaskLen)
	}
	if m.TokenLen != CSRFTokenLen {
		t.Errorf("expected TokenLen: %v, got %v", CSRFTokenLen, m.TokenLen)
	}
	if m.Skipper == nil {
		t.Error(errSkipperNil)
	}
}
