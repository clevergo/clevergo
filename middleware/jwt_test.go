// Copyright 2016 The Gem Authors. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package middleware

import (
	"fmt"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-gem/gem"
	"github.com/go-gem/tests"
	"github.com/valyala/fasthttp"
)

var (
	signKey = []byte("foobar")

	m = NewJWT(jwt.SigningMethodHS256, func(token *jwt.Token) (interface{}, error) {
		return signKey, nil
	})

	token = jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"name": "foo",
	})
)

func jwtHandler(ctx *gem.Context) {
	ctx.HTML(fasthttp.StatusOK, fasthttp.StatusMessage(fasthttp.StatusOK))
}

func TestJWT(t *testing.T) {
	signedStr, err := token.SignedString(signKey)
	if err != nil {
		t.Fatal(err)
	}

	m.Skipper = nil

	router := gem.NewRouter()
	router.Use(m)
	router.GET("/", jwtHandler)
	router.POST("/", jwtHandler)

	srv := gem.New("", router.Handler())

	if m.Skipper == nil {
		t.Error(errSkipperNil)
	}

	// Empty jwt token.
	test1 := tests.New(srv)
	test1.Timeout = 200 * time.Microsecond
	test1.Expect().Status(fasthttp.StatusBadRequest).Body(fasthttp.StatusMessage(fasthttp.StatusBadRequest))
	if err := test1.Run(); err != nil {
		t.Error(err)
	}

	// Request with signed string in header.
	test2 := tests.New(srv)
	test2.Timeout = 200 * time.Microsecond
	test2.Headers[gem.HeaderAuthorization] = fmt.Sprintf("%s %s", gem.HeaderBearer, signedStr)
	test2.Expect().Status(fasthttp.StatusOK).Body(fasthttp.StatusMessage(fasthttp.StatusOK))
	if err := test2.Run(); err != nil {
		t.Error(err)
	}

	// Request with signed string in post form or query string.
	test3 := tests.New(srv)
	test3.Timeout = 200 * time.Microsecond
	test3.Url = "/?_jwt=" + signedStr
	test3.Expect().Status(fasthttp.StatusOK).Body(fasthttp.StatusMessage(fasthttp.StatusOK))
	if err := test3.Run(); err != nil {
		t.Error(err)
	}

	// Request with invalid signed string.
	test4 := tests.New(srv)
	test4.Timeout = 200 * time.Microsecond
	test4.Url = "/?_jwt=invalidSignedString"
	test4.Expect().Status(fasthttp.StatusUnauthorized).Body(fasthttp.StatusMessage(fasthttp.StatusUnauthorized))
	if err := test4.Run(); err != nil {
		t.Error(err)
	}

	// Custom NewClaims
	m.NewClaims = func() jwt.Claims {
		return new(jwt.MapClaims)
	}
	test5 := tests.New(srv)
	test5.Timeout = 200 * time.Microsecond
	test5.Headers[gem.HeaderAuthorization] = fmt.Sprintf("%s %s", gem.HeaderBearer, signedStr)
	test5.Expect().Status(fasthttp.StatusOK).Body(fasthttp.StatusMessage(fasthttp.StatusOK))
	if err := test5.Run(); err != nil {
		t.Error(err)
	}

	// Always skip.
	m.Skipper = alwaysSkipper
	test7 := tests.New(srv)
	test7.Timeout = 200 * time.Microsecond
	test7.Expect().Status(fasthttp.StatusOK).Body(fasthttp.StatusMessage(fasthttp.StatusOK))
	if err := test7.Run(); err != nil {
		t.Error(err)
	}
}
