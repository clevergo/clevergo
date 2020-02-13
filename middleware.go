// Copyright 2020 CleverGo. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package clevergo

// MiddlewareFunc is a Handle.
type MiddlewareFunc Handle

// Chain wraps handle with middlewares, middlewares will be invoked in sequence.
func Chain(handle Handle, middlewares ...MiddlewareFunc) Handle {
	return func(ctx *Context) (err error) {
		for _, f := range middlewares {
			if err = f(ctx); err != nil {
				return
			}
		}

		return handle(ctx)
	}
}
