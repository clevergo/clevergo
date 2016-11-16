// Copyright 2016 The Gem Authors. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package middleware

import (
	"bufio"
	"testing"
	"time"

	"github.com/go-gem/gem"
	"github.com/go-gem/gem/bytes"
	"github.com/valyala/fasthttp"
)

func TestBodyLimit(t *testing.T) {
	bl := NewBodyLimit(7 * bytes.B)

	s := gem.New()

	router := gem.NewRouter()
	router.Use(bl)
	router.POST("/", func(c *gem.Context) {
		c.HTML(fasthttp.StatusOK, "OK")
	})

	s.Init(router.Handler)

	rw := &readWriter{}
	br := bufio.NewReader(&rw.w)
	var resp fasthttp.Response
	ch := make(chan error)

	newLine := "\r\n"
	reqStr := "POST / HTTP/1.1" + newLine +
		"Content-Length:8" + newLine +
		"Content-Type:application/x-www-form-urlencoded" + newLine +
		newLine +
		"name=123"

	rw.r.WriteString(reqStr)
	go func() {
		ch <- s.Server.ServeConn(rw)
	}()
	select {
	case err := <-ch:
		if err != nil {
			t.Fatalf("return error %s", err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("timeout")
	}
	if err := resp.Read(br); err != nil {
		t.Fatalf("Unexpected error when reading response: %s", err)
	}
	if resp.StatusCode() != fasthttp.StatusRequestEntityTooLarge {
		t.Errorf("Expected status code %d, got %d", fasthttp.StatusRequestEntityTooLarge, resp.StatusCode())
	}

	router = gem.NewRouter()
	router.Use(bl)
	router.POST("/", func(c *gem.Context) {
		c.HTML(fasthttp.StatusOK, "OK")
	})

	s.Init(router.Handler)

	// Increase limit size to 8B.
	bl.Limit = 8 * bytes.B
	rw.r.WriteString(reqStr)
	go func() {
		ch <- s.Server.ServeConn(rw)
	}()
	select {
	case err := <-ch:
		if err != nil {
			t.Fatalf("return error %s", err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("timeout")
	}
	if err := resp.Read(br); err != nil {
		t.Fatalf("Unexpected error when reading response: %s", err)
	}
	if resp.StatusCode() != fasthttp.StatusOK {
		t.Errorf("Expected status code %d, got %d", fasthttp.StatusOK, resp.StatusCode())
	}

	// Skip body limit middleware
	bl.Skipper = func(c *gem.Context) bool {
		return true
	}

	rw.r.WriteString(reqStr)
	go func() {
		ch <- s.Server.ServeConn(rw)
	}()
	select {
	case err := <-ch:
		if err != nil {
			t.Fatalf("return error %s", err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("timeout")
	}
	if err := resp.Read(br); err != nil {
		t.Fatalf("Unexpected error when reading response: %s", err)
	}
	if resp.StatusCode() != fasthttp.StatusOK {
		t.Errorf("Expected status code %d, got %d", fasthttp.StatusOK, resp.StatusCode())
	}
}
