// Copyright 2016 The Gem Authors. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package gem

import (
	"testing"

	"github.com/go-gem/sessions"
)

func TestVersion(t *testing.T) {
	if version != Version() {
		t.Errorf("Expected version: %q, got %q.\n", version, Version())
	}
}

func TestName(t *testing.T) {
	if name != Name() {
		t.Errorf("Expected name: %q, got %q.\n", name, Name())
	}
}

func TestServer_SetLogger(t *testing.T) {
	var logger Logger
	srv := New("", func(ctx *Context) {})
	srv.SetLogger(logger)
	if srv.logger != logger {
		t.Errorf("Failed to set logger")
	}
}

func TestServer_SetSessionsStoret(t *testing.T) {
	var store sessions.Store
	srv := New("", func(ctx *Context) {})
	srv.SetSessionsStore(store)
	if srv.sessionsStore != store {
		t.Errorf("Failed to set sessions store")
	}
}
