// Copyright 2016 The Gem Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license
// that can be found in the LICENSE file.

package gem

import "testing"

func TestNewHandlerOption(t *testing.T) {
	var m1, m2, m3 Middleware
	option := NewHandlerOption(m1, m2, m3)

	if len(option.Middlewares) != 3 {
		t.Errorf("expected option middleware count %d, got %d", 3, len(option.Middlewares))
	}
}
