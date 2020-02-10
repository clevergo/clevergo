// Copyright 2020 CleverGo. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package clevergo

// MiddlewareFunc is a function which receives an http.Handler and returns another http.Handler.
type MiddlewareFunc func(Handle) Handle

// Chain wraps handler with middlewares, middlewares will be invoked in sequence.
func Chain(handle Handle, middlewares ...MiddlewareFunc) Handle {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handle = middlewares[i](handle)
	}
	return handle
}
