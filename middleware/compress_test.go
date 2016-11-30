// Copyright 2016 The Gem Authors. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package middleware

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/go-gem/gem"
	"github.com/go-gem/tests"
	"github.com/valyala/fasthttp"
)

func TestCompress(t *testing.T) {
	router := gem.NewRouter()
	router.Use(NewCompress(CompressBestCompression))
	router.GET("/", func(ctx *gem.Context) {
		ctx.HTML(fasthttp.StatusOK, fasthttp.StatusMessage(fasthttp.StatusOK))
	})

	srv := gem.New("", router.Handler())

	var err error

	// Expected uncompressed response.
	test1 := tests.New(srv)
	test1.Expect().Status(fasthttp.StatusOK).Custom(func(resp fasthttp.Response) error {
		if len(resp.Header.Peek("Content-Encoding")) > 0 {
			return fmt.Errorf("Expected uncompressed response, got compressed response")
		}
		return nil
	})
	if err = test1.Run(); err != nil {
		t.Error(err)
	}

	// Expected gzip compressed response.
	test2 := tests.New(srv)
	test2.Timeout = time.Second
	test2.Headers[gem.HeaderAcceptEncoding] = gem.HeaderAcceptEncodingGzip
	test2.Expect().Status(fasthttp.StatusOK).Custom(func(resp fasthttp.Response) error {
		if ce := resp.Header.Peek(gem.HeaderContentEncoding); len(ce) == 0 || !bytes.Equal(ce, gem.HeaderAcceptEncodingGzipBytes) {
			return fmt.Errorf("Expected Content-Encoding %q, got %q", gem.HeaderAcceptEncodingGzipBytes, ce)
		}

		return nil
	})
	if err = test2.Run(); err != nil {
		t.Error(err)
	}

	// Expected deflate compressed response.
	test3 := tests.New(srv)
	test3.Timeout = time.Second
	test3.Headers[gem.HeaderAcceptEncoding] = gem.HeaderAcceptEncodingDeflate
	test3.Expect().Status(fasthttp.StatusOK).Custom(func(resp fasthttp.Response) error {
		if ce := resp.Header.Peek(gem.HeaderContentEncoding); len(ce) == 0 || !bytes.Equal(ce, gem.HeaderAcceptEncodingDeflateBytes) {
			return fmt.Errorf("Expected Content-Encoding %q, got %q", gem.HeaderAcceptEncodingDeflateBytes, ce)
		}

		return nil
	})
	if err = test3.Run(); err != nil {
		t.Error(err)
	}
}
