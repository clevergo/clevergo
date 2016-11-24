// Copyright 2016 The Gem Authors. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package middleware

import (
	"github.com/go-gem/gem"
)

// JWT JSON WEB TOKEN middleware.
type JWT struct {
	// Skipper defines a function to skip middleware.
	Skipper Skipper
}

// NewJWT returns a JWT instance with the default
// configuration.
func NewJWT() *JWT {
	return &JWT{
		Skipper: defaultSkipper,
	}
}

// Handle implements Middleware's Handle function.
func (jwt *JWT) Handle(next gem.Handler) gem.Handler {
	return gem.HandlerFunc(func(ctx *gem.Context) {
		next.Handle(ctx)
	})
}
