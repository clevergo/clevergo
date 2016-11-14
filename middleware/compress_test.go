// Copyright 2016 The Gem Authors. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package middleware

import (
	"bufio"
	"bytes"
	"net"
	"testing"
	"time"

	"github.com/go-gem/gem"
	"github.com/valyala/fasthttp"
)

func newContext() *gem.Context {
	return &gem.Context{
		RequestCtx: &fasthttp.RequestCtx{},
	}
}

func TestGzip(t *testing.T) {
	s := gem.New()

	router := gem.NewRouter()
	router.Use(NewGzip(CompressBestCompression))
	router.GET("/", func(c *gem.Context) {
		c.HTML(fasthttp.StatusOK, "Compress")
	})

	s.Init(router.Handler)

	rw := &readWriter{}
	br := bufio.NewReader(&rw.w)
	var resp fasthttp.Response
	ch := make(chan error)

	rw.r.WriteString("GET / HTTP/1.1\r\n\r\n")
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
	if len(resp.Header.Peek("Content-Encoding")) > 0 {
		t.Errorf("Expected no compress response, got compressed response")
	}

	// Accept-Encoding: gzip
	rw.r.WriteString("GET / HTTP/1.1\r\nAccept-Encoding: gzip\r\n\r\n")
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
	if ce := resp.Header.Peek("Content-Encoding"); len(ce) == 0 || !bytes.Equal(ce, []byte("gzip")) {
		t.Errorf("Expected Content-Encoding %q, got %q", "gzip", ce)
	}

	// Accept-Encoding: deflate, gzip
	rw.r.WriteString("GET / HTTP/1.1\r\nAccept-Encoding: deflate, gzip\r\n\r\n")
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
	if ce := resp.Header.Peek("Content-Encoding"); len(ce) == 0 || !bytes.Equal(ce, []byte("gzip")) {
		t.Errorf("Expected Content-Encoding %q, got %q", "gzip", ce)
	}

	// Accept-Encoding: deflate
	rw.r.WriteString("GET / HTTP/1.1\r\nAccept-Encoding: deflate\r\n\r\n")
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
	if ce := resp.Header.Peek("Content-Encoding"); len(ce) == 0 || !bytes.Equal(ce, []byte("deflate")) {
		t.Errorf("Expected Content-Encoding %q, got %q", "deflate", ce)
	}
}

type readWriter struct {
	net.Conn
	r bytes.Buffer
	w bytes.Buffer
}

var zeroTCPAddr = &net.TCPAddr{
	IP: net.IPv4zero,
}

func (rw *readWriter) Close() error {
	return nil
}

func (rw *readWriter) Read(b []byte) (int, error) {
	return rw.r.Read(b)
}

func (rw *readWriter) Write(b []byte) (int, error) {
	return rw.w.Write(b)
}

func (rw *readWriter) RemoteAddr() net.Addr {
	return zeroTCPAddr
}

func (rw *readWriter) LocalAddr() net.Addr {
	return zeroTCPAddr
}

func (rw *readWriter) SetReadDeadline(t time.Time) error {
	return nil
}

func (rw *readWriter) SetWriteDeadline(t time.Time) error {
	return nil
}
