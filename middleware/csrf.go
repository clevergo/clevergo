// Copyright 2016 The Gem Authors. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package middleware

import (
	"github.com/go-gem/gem"
)

// CSRF Cross-site request forgery protection middleware.
type CSRF struct {
	// Skipper defines a function to skip middleware.
	Skipper Skipper
}

// NewCSRF returns a CSRF instance with the default
// configuration.
func NewCSRF() *CSRF {
	return &CSRF{
		Skipper: defaultSkipper,
	}
}

// Handle implements Middleware's Handle function.
func (csrf *CSRF) Handle(next gem.Handler) gem.Handler {
	return gem.HandlerFunc(func(ctx *gem.Context) {
		next.Handle(ctx)
	})
}
