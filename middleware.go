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

// WrapHandler wraps a HTTP handler and returns a middleware.
func WrapHandler(h http.Handler) MiddlewareFunc {
	return func(handle Handle) Handle {
		return func(ctx *Context) error {
			h.ServeHTTP(ctx.Response, ctx.Request)
			return handle(ctx)
		}
	}
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
