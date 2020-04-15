// Copyright 2020 CleverGo. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package clevergo

import (
	"log"
	"net/http"
	"runtime/debug"
	"strings"
	"sync"
)

// MiddlewareFunc is a function that receives a handle and returns a handle.
type MiddlewareFunc func(Handle) Handle

// WrapH wraps a HTTP handler and returns a middleware.
func WrapH(h http.Handler) MiddlewareFunc {
	return func(handle Handle) Handle {
		return func(ctx *Context) error {
			h.ServeHTTP(ctx.Response, ctx.Request)
			return handle(ctx)
		}
	}
}

// WrapHH wraps func(http.Handler) http.Handler and returns a middleware.
func WrapHH(fn func(http.Handler) http.Handler) MiddlewareFunc {
	nextHandler := new(middlewareHandler)
	handler := fn(nextHandler)
	return func(handle Handle) Handle {
		return func(ctx *Context) error {
			state := getMiddlewareState()
			defer func() {
				putMiddlewareState(state)
			}()
			state.ctx = ctx
			state.next = handle
			ctx.WithValue(nextHandler, state)
			handler.ServeHTTP(ctx.Response, ctx.Request)
			return state.err
		}
	}
}

type middlewareHandler struct {
}

func (h *middlewareHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	state := r.Context().Value(h).(*middlewareState)
	defer func(w http.ResponseWriter, r *http.Request) {
		state.ctx.Response = w
		state.ctx.Request = r
	}(state.ctx.Response, state.ctx.Request)
	state.ctx.Response = w
	state.ctx.Request = r
	state.err = state.next(state.ctx)
}

var middlewareStatePool = sync.Pool{
	New: func() interface{} {
		return new(middlewareState)
	},
}

func getMiddlewareState() *middlewareState {
	ctx := middlewareStatePool.Get().(*middlewareState)
	ctx.reset()
	return ctx
}

func putMiddlewareState(ctx *middlewareState) {
	middlewareStatePool.Put(ctx)
}

type middlewareState struct {
	ctx  *Context
	next Handle
	err  error
}

func (s *middlewareState) reset() {
	s.ctx = nil
	s.next = nil
	s.err = nil
}

// Chain wraps handle with middlewares, middlewares will be invoked in sequence.
func Chain(handle Handle, middlewares ...MiddlewareFunc) Handle {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handle = middlewares[i](handle)
	}

	return handle
}

type recovery struct {
	debug  bool
	logger *log.Logger
}

func (r *recovery) handle(ctx *Context, err interface{}) {
	ctx.Response.WriteHeader(http.StatusInternalServerError)
	r.logf(err)
	if r.debug {
		r.logf(debug.Stack())
	}
}

func (r *recovery) logf(v interface{}) {
	if r.logger != nil {
		r.logger.Printf("%s\n", v)
		return
	}

	log.Printf("%s\n", v)
}

func (r *recovery) middleware(next Handle) Handle {
	return func(ctx *Context) error {
		defer func() {
			if err := recover(); err != nil {
				r.handle(ctx, err)
			}
		}()
		return next(ctx)
	}
}

// Recovery returns a recovery middleware.
func Recovery(debug bool) MiddlewareFunc {
	return RecoveryLogger(debug, nil)
}

// RecoveryLogger returns a recovery middleware with the given logger.
func RecoveryLogger(debug bool, logger *log.Logger) MiddlewareFunc {
	r := &recovery{debug: debug, logger: logger}
	return r.middleware
}

// Skipper is a function that indicates whether current request is skippable.
type Skipper func(ctx *Context) bool

// PathSkipper returns skipper with the given patterns.
// Pattern has two forms, one is that contains a certain path, another contains a wildcard,
// both of them are case-insensitive.
//   Pattern     Path            Skippable
//   ""          "/"             false
//   "/"         "/"             true
//   "/"         "/login"        false
//   "/login"    "/login"        true
//   "/login"    "/Login"        true
//   "/login"    "/LOGIN"        true
//   "/guest*"   "/guest"        true
//   "/guest*"   "/guest/foo"    true
//   "/guest*"   "/guest/bar"    true
func PathSkipper(patterns ...string) Skipper {
	return func(ctx *Context) bool {
		for _, pattern := range patterns {
			if pattern == "" {
				continue
			}
			if pattern[len(pattern)-1] == '*' && len(ctx.Request.URL.Path) >= len(pattern)-1 {
				length := len(pattern) - 1
				if strings.EqualFold(ctx.Request.URL.Path[:length], pattern[:length]) {
					return true
				}
			}
			if strings.EqualFold(pattern, ctx.Request.URL.Path) {
				return true
			}
		}
		return false
	}
}
