// Copyright 2020 CleverGo. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package clevergo

import "net/http"

// Middleware is a function which receives an http.Handler and returns another http.Handler.
type Middleware func(http.Handler) http.Handler

// Chain wraps handler with middlewares, middlewares will be invoked in sequence.
func Chain(handler http.Handler, middlewares ...Middleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}
