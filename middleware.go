// Copyright 2016 The Gem Authors. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package gem

// Middleware interface.
type Middleware interface {
	Wrap(next Handler) Handler
}
