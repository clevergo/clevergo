// Copyright 2020 CleverGo. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package clevergo

// MiddlewareFunc is a function that receives a handle and returns a handle.
type MiddlewareFunc func(Handle) Handle

// Chain wraps handle with middlewares, middlewares will be invoked in sequence.
func Chain(handle Handle, middlewares ...MiddlewareFunc) Handle {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handle = middlewares[i](handle)
	}

	return handle
}
