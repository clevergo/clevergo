// Copyright 2016 The Gem Authors. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package gem

import (
	"testing"
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
