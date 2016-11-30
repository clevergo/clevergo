// Copyright 2016 The Gem Authors. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package middleware

import (
	"strconv"
	"testing"

	"github.com/go-gem/gem"
	"github.com/go-gem/tests"
	"github.com/valyala/fasthttp"
)

func TestBodyLimit(t *testing.T) {
	reqPayload := "name=foo"
	reqPayloadSize := len(reqPayload)

	m := NewBodyLimit(reqPayloadSize - 1)
	m.Skipper = nil

	router := gem.NewRouter()
	router.Use(m)
	router.POST("/", func(ctx *gem.Context) {
		ctx.HTML(fasthttp.StatusOK, "OK")
	})

	srv := gem.New("", router.Handler())

	if m.Skipper == nil {
		t.Error(errSkipperNil)
	}

	var err error

	// Entity too large.
	test1 := tests.New(srv)
	test1.Method = gem.MethodPost
	test1.Headers[gem.HeaderContentType] = gem.HeaderContentTypeForm
	test1.Headers[gem.HeaderContentLength] = strconv.Itoa(reqPayloadSize)
	test1.Payload = reqPayload
	test1.Expect().Status(fasthttp.StatusRequestEntityTooLarge).
		Body(fasthttp.StatusMessage(fasthttp.StatusRequestEntityTooLarge))
	if err = test1.Run(); err != nil {
		t.Error(err)
	}

	// Increase limit size.
	m.Limit = reqPayloadSize
	test2 := tests.New(srv)
	test2.Method = gem.MethodPost
	test2.Headers[gem.HeaderContentType] = gem.HeaderContentTypeForm
	test2.Headers[gem.HeaderContentLength] = strconv.Itoa(reqPayloadSize)
	test2.Payload = reqPayload
	test2.Expect().Status(fasthttp.StatusOK).
		Body(fasthttp.StatusMessage(fasthttp.StatusOK))
	if err = test2.Run(); err != nil {
		t.Error(err)
	}

	// Always skip.
	m.Skipper = alwaysSkipper
	test3 := tests.New(srv)
	test3.Method = gem.MethodPost
	test3.Expect().Status(fasthttp.StatusOK).
		Body(fasthttp.StatusMessage(fasthttp.StatusOK))
	if err = test3.Run(); err != nil {
		t.Error(err)
	}
}
