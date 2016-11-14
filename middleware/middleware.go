// Copyright 2016 The Gem Authors. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package middleware

import (
	"github.com/go-gem/gem"
)

var defaultSkipper = func(c *gem.Context) bool {
	return false
}

// Skipper defines a function to skip middleware.
type Skipper func(c *gem.Context) bool
