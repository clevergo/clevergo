// Copyright 2020 CleverGo. All rights reserved.
// Use of this source code is governed by a MIT style license that can be found
// in the LICENSE file.

package clevergo

import (
	"net/http"
	"runtime/debug"
	"sync"
)

// MiddlewareFunc is a function that receives a handle and returns a handle.
type MiddlewareFunc func(Handle) Handle

// WrapH wraps a HTTP handler and returns a middleware.
func WrapH(h http.Handler) MiddlewareFunc {
	return func(handle Handle) Handle {
		return func(c *Context) error {
			h.ServeHTTP(c.Response, c.Request)
			return handle(c)
		}
	}
}

// WrapHH wraps func(http.Handler) http.Handler and returns a middleware.
func WrapHH(fn func(http.Handler) http.Handler) MiddlewareFunc {
	nextHandler := new(middlewareHandler)
	handler := fn(nextHandler)
	return func(handle Handle) Handle {
		return func(c *Context) error {
			state := getMiddlewareState()
			defer func() {
				putMiddlewareState(state)
			}()
			state.ctx = c
			state.next = handle
			c.WithValue(nextHandler, state)
			handler.ServeHTTP(c.Response, c.Request)
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
	state := middlewareStatePool.Get().(*middlewareState)
	state.reset()
	return state
}

func putMiddlewareState(state *middlewareState) {
	middlewareStatePool.Put(state)
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
}

func (r *recovery) middleware(next Handle) Handle {
	return func(c *Context) (err error) {
		defer func() {
			if data := recover(); data != nil {
				err = PanicError{
					Context: c,
					Data:    data,
					Stack:   debug.Stack(),
				}
			}
		}()
		err = next(c)
		return
	}
}

// Recovery returns a recovery middleware with debug enabled by default.
func Recovery() MiddlewareFunc {
	r := &recovery{}
	return r.middleware
}

// ServerHeader is a middleware that sets Server header.
func ServerHeader(value string) MiddlewareFunc {
	return func(next Handle) Handle {
		return func(c *Context) error {
			c.SetHeader("Server", value)
			return next(c)
		}
	}
}
