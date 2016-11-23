// Copyright 2016 The Gem Authors. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package middleware

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"testing"
	"time"

	"github.com/go-gem/gem"
	"github.com/valyala/fasthttp"
)

func TestBasicAuth(t *testing.T) {
	trueUsername := "foo"
	truePsw := "123456"
	encodedStr := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", trueUsername, truePsw)))

	ba := NewBasicAuth(func(username, psw string) bool {
		return trueUsername == username && truePsw == psw
	})

	router := gem.NewRouter()
	router.Use(ba)
	router.POST("/", func(c *gem.Context) {
		c.HTML(fasthttp.StatusOK, "OK")
	})

	s := gem.New("", router.Handler)

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
		ch <- s.ServeConn(rw)
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
	if resp.StatusCode() != fasthttp.StatusUnauthorized {
		t.Errorf("Expected status code %d, got %d", fasthttp.StatusUnauthorized, resp.StatusCode())
	}
	if string(resp.Body()) != fasthttp.StatusMessage(fasthttp.StatusUnauthorized) {
		t.Errorf("Expected response body %q, got %q", fasthttp.StatusMessage(fasthttp.StatusUnauthorized), resp.Body())
	}

	newLine = "\r\n"
	reqStr = "POST / HTTP/1.1" + newLine +
		"Content-Length:8" + newLine +
		"Content-Type:application/x-www-form-urlencoded" + newLine +
		"Authorization:Basic " + encodedStr + newLine +
		newLine +
		"name=123"

	rw.r.WriteString(reqStr)
	go func() {
		ch <- s.ServeConn(rw)
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
}
