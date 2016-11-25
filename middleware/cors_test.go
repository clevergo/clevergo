// Copyright 2016 The Gem Authors. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package middleware

import (
	"testing"

	"github.com/go-gem/tests"
	"github.com/valyala/fasthttp"
)

var (
	cors = NewCORS()
)

/*


var (
	cors = NewCORS()
)
func TestCORS(t *testing.T) {
	router := gem.NewRouter()
	router.Use(cors)
	router.GET("/", func(c *gem.Context) {
		c.HTML(fasthttp.StatusOK, "OK")
	})

	s := gem.New("", router.Handler)

	test := test.New(s)

	test.Expect().
		Header("Content-Type","aaa").
		Status(200).
		Body("OK")

	if err := test.Run(); err != nil {
		t.Error(err)
	}
}*/

func TestFastHTTP(t *testing.T) {
	contentType := "text/html; charset=utf-8"
	statusCode := fasthttp.StatusBadRequest
	respBody := fasthttp.StatusMessage(fasthttp.StatusBadRequest)

	// Fake server
	srv := &fasthttp.Server{
		Handler: func(ctx *fasthttp.RequestCtx) {
			ctx.SetContentType(contentType)
			ctx.SetStatusCode(statusCode)
			ctx.SetBodyString(respBody)
		},
	}

	// Create a Test instance.
	test := tests.New(srv)

	// Customize request.
	// See Test struct.
	test.Url = "/"

	// Add excepted result.
	test.Expect().
		Status(statusCode).
		Header("Content-Type", contentType).
		Body(respBody)

	// Custom checking function.
	test.Expect().Custom(func(resp fasthttp.Response) error {
		// check response.

		return nil
	})

	// Run test.
	if err := test.Run(); err != nil {
		t.Error(err)
	}
}
