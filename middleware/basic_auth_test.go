// Copyright 2016 The Gem Authors. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package middleware

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/go-gem/gem"
	"github.com/go-gem/tests"
	"github.com/valyala/fasthttp"
)

func TestBasicAuth(t *testing.T) {
	trueUsername := "foo"
	truePsw := "bar"
	encodedStr := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", trueUsername, truePsw)))

	incorrectPsw := "incorrectPassword"
	incorrectEncodedStr := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", trueUsername, incorrectPsw)))

	authorization := fmt.Sprintf("%s %s", gem.HeaderBasic, encodedStr)

	m := NewBasicAuth(func(username, psw string) bool {
		return trueUsername == username && truePsw == psw
	})
	m.Skipper = nil

	router := gem.NewRouter()
	router.Use(m)
	router.GET("/", func(ctx *gem.Context) {
		ctx.HTML(fasthttp.StatusOK, fasthttp.StatusMessage(fasthttp.StatusOK))
	})

	srv := gem.New("", router.Handler())

	if m.Skipper == nil {
		t.Error(errSkipperNil)
	}

	var err error

	// Correct authorization.
	test1 := tests.New(srv)
	test1.Headers[gem.HeaderAuthorization] = authorization
	test1.Expect().Status(fasthttp.StatusOK).Body(fasthttp.StatusMessage(fasthttp.StatusOK))
	if err = test1.Run(); err != nil {
		t.Error(err)
	}

	// Empty authorization.
	test2 := tests.New(srv)
	test2.Expect().Status(fasthttp.StatusUnauthorized)
	if err = test2.Run(); err != nil {
		t.Error(err)
	}

	// Invalid base64 encoded string.
	test3 := tests.New(srv)
	test3.Headers[gem.HeaderAuthorization] = "Basic invalidBase64EncodedString"
	test3.Expect().Status(fasthttp.StatusUnauthorized)
	if err = test3.Run(); err != nil {
		t.Error(err)
	}

	// Incorrect password.
	test4 := tests.New(srv)
	test4.Headers[gem.HeaderAuthorization] = fmt.Sprintf("%s %s", gem.HeaderBasic, incorrectEncodedStr)
	test4.Expect().Status(fasthttp.StatusUnauthorized)
	if err = test4.Run(); err != nil {
		t.Error(err)
	}

	// Always skip.
	m.Skipper = alwaysSkipper
	test5 := tests.New(srv)
	test5.Expect().Status(fasthttp.StatusOK).
		Body(fasthttp.StatusMessage(fasthttp.StatusOK))
	if err = test5.Run(); err != nil {
		t.Error(err)
	}
}
