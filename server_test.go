// Copyright 2016 The Gem Authors. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package gem

import (
	"bytes"
	"testing"
)

func TestServer_SetLogger(t *testing.T) {
	var logger Logger
	srv := New("")
	srv.SetLogger(logger)
	if srv.logger != logger {
		t.Error("failed to set logger")
	}
}

func TestVersion(t *testing.T) {
	if Version() != version {
		t.Errorf("expected version number %q, got %q", version, Version())
	}
}

func TestServerInit(t *testing.T) {
	body := []byte("foo")
	handler := HandlerFunc(func(ctx *Context) {
		ctx.Response.Write(body)
	})

	srv := New("")
	srv.init(handler)

	resp := &mockResponseWriter{}
	srv.Server.Handler.ServeHTTP(resp, nil)

	if !bytes.Equal(resp.body, body) {
		t.Errorf("expected response body %q, got %q", body, resp.body)
	}
}
