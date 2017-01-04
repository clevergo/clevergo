// Copyright 2016 The Gem Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license
// that can be found in the LICENSE file.

package gem

// Handler for processing incoming requests.
type Handler interface {
	Handle(*Context)
}

// The HandlerFunc type is an adapter to allow the use of
// ordinary functions as HTTP handlers. If f is a function
// with the appropriate signature, HandlerFunc(f) is a
// Handler that calls f.
type HandlerFunc func(*Context)

// Handle calls f(ctx).
func (f HandlerFunc) Handle(ctx *Context) {
	f(ctx)
}

// HandlerOption option for handler.
type HandlerOption struct {
	Middlewares []Middleware
}

// NewHandlerOption returns HandlerOption instance by the
// given middlewares.
func NewHandlerOption(middlewares ...Middleware) *HandlerOption {
	return &HandlerOption{
		Middlewares: middlewares,
	}
}
