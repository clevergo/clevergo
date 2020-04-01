// Copyright 2020 CleverGo. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package clevergo

import (
	"log"
	"net/http"
	"runtime/debug"
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
			state := &middlewareCtx{ctx: ctx, handle: handle}
			ctx.WithValue(nextHandler, state)
			handler.ServeHTTP(ctx.Response, ctx.Request)
			return state.err
		}
	}
}

type middlewareHandler struct {
}

func (h *middlewareHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	state := r.Context().Value(h).(*middlewareCtx)
	defer func(w http.ResponseWriter, r *http.Request) {
		state.ctx.Response = w
		state.ctx.Request = r
	}(state.ctx.Response, state.ctx.Request)
	state.ctx.Response = w
	state.ctx.Request = r
	state.next()
}

type middlewareCtx struct {
	ctx    *Context
	handle Handle
	err    error
}

func (m middlewareCtx) next() {
	m.err = m.handle(m.ctx)
}

// Chain wraps handle with middlewares, middlewares will be invoked in sequence.
func Chain(handle Handle, middlewares ...MiddlewareFunc) Handle {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handle = middlewares[i](handle)
	}

	return handle
}

type recovery struct {
	debug bool
}

func (r *recovery) handle(ctx *Context, err interface{}) {
	ctx.Response.WriteHeader(http.StatusInternalServerError)
	log.Println(err)
	if r.debug {
		debug.PrintStack()
	}
}

// Recovery returns a recovery middleware.
func Recovery(debug bool) MiddlewareFunc {
	m := &recovery{debug: debug}
	return func(next Handle) Handle {
		return func(ctx *Context) error {
			defer func() {
				if err := recover(); err != nil {
					m.handle(ctx, err)
				}
			}()
			return next(ctx)
		}
	}
}
