// Copyright 2016 The Gem Authors. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package middleware

import (
	"errors"
	"github.com/go-gem/gem"
)

var defaultSkipper = func(ctx *gem.Context) bool {
	return false
}

var alwaysSkipper = func(ctx *gem.Context) bool {
	return true
}

var (
	errSkipperNil = errors.New("The skipper should not be nil")
)

// Skipper defines a function to skip middleware.
type Skipper func(ctx *gem.Context) bool
