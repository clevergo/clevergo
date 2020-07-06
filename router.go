// Copyright 2020 CleverGo. All rights reserved.
// Use of this source code is governed by a MIT style license that can be found
// in the LICENSE file.

package clevergo

import "net/http"

// Router is an router interface.
type Router interface {
	// Group creates a sub Router with the given optional route group options.
	Group(path string, opts ...RouteGroupOption) Router

	// Get registers a new GET request handler function with the given path and optional route options.
	Get(path string, handle Handle, opts ...RouteOption)

	// Head registers a new HEAD request handler function with the given path and optional route options.
	Head(path string, handle Handle, opts ...RouteOption)

	// Options registers a new Options request handler function with the given path and optional route options.
	Options(path string, handle Handle, opts ...RouteOption)

	// Post registers a new POST request handler function with the given path and optional route options.
	Post(path string, handle Handle, opts ...RouteOption)

	// Put registers a new PUT request handler function with the given path and optional route options.
	Put(path string, handle Handle, opts ...RouteOption)

	// Patch registers a new PATCH request handler function with the given path and optional route options.
	Patch(path string, handle Handle, opts ...RouteOption)

	// Delete registers a new DELETE request handler function with the given path and optional route options.
	Delete(path string, handle Handle, opts ...RouteOption)

	// Any registers a new request handler function that matches any HTTP methods with the given path and
	// optional route options. GET, HEAD, POST, PUT, DELETE, CONNECT, OPTIONS, TRACE, PATCH.
	Any(path string, handle Handle, opts ...RouteOption)

	// HandleFunc registers a new request handler function with the given path, method and optional route options.
	//
	// For Get, Head, Options, Post, Put, Patch and Delete requests the respective shortcut
	// functions can be used.
	//
	// This function is intended for bulk loading and to allow the usage of less
	// frequently used, non-standardized or custom methods (e.g. for internal
	// communication with a proxy).
	Handle(method, path string, handle Handle, opts ...RouteOption)

	// Handler is an adapter for registering http.Handler.
	Handler(method, path string, handler http.Handler, opts ...RouteOption)

	// HandlerFunc is an adapter for registering http.HandlerFunc.
	HandlerFunc(method, path string, f http.HandlerFunc, opts ...RouteOption)
}
