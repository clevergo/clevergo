// Copyright 2016 The Gem Authors. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package gem

import (
	"testing"

	"github.com/go-gem/log"
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
	var logger log.Logger
	s := New()
	s.SetLogger(logger)
	if s.logger != logger {
		t.Errorf("Failed to set logger")
	}
}

func TestServer_SetSessionsStoret(t *testing.T) {
	var store sessions.Store
	s := New()
	s.SetSessionsStore(store)
	if s.sessionsStore != store {
		t.Errorf("Failed to set sessions store")
	}
}
