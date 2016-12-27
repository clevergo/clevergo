// Copyright 2013 Julien Schmidt, 2016 The Gem Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package gem

import (
	"net/http"
)

// Router is a http.Handler which can be used to dispatch requests to different
// handler functions via configurable routes
type Router struct {
	trees map[string]*node

	middlewares []Middleware

	// Enables automatic redirection if the current route can't be matched but a
	// handler for the path with (without) the trailing slash exists.
	// For example if /foo/ is requested but a route only exists for /foo, the
	// client is redirected to /foo with http status code 301 for GET requests
	// and 307 for all other request methods.
	RedirectTrailingSlash bool

	// If enabled, the router tries to fix the current request path, if no
	// handle is registered for it.
	// First superfluous path elements like ../ or // are removed.
	// Afterwards the router does a case-insensitive lookup of the cleaned path.
	// If a handle can be found for this route, the router makes a redirection
	// to the corrected path with status code 301 for GET requests and 307 for
	// all other request methods.
	// For example /FOO and /..//Foo could be redirected to /foo.
	// RedirectTrailingSlash is independent of this option.
	RedirectFixedPath bool

	// If enabled, the router checks if another method is allowed for the
	// current route, if the current request can not be routed.
	// If this is the case, the request is answered with 'Method Not Allowed'
	// and HTTP status code 405.
	// If no other Method is allowed, the request is delegated to the NotFound
	// handler.
	HandleMethodNotAllowed bool

	// If enabled, the router automatically replies to OPTIONS requests.
	// Custom OPTIONS handlers take priority over automatic replies.
	HandleOPTIONS bool

	// Configurable http.Handler which is called when no matching route is
	// found. If it is not set, http.NotFound is used.
	NotFound Handler

	// Configurable http.Handler which is called when a request
	// cannot be routed and HandleMethodNotAllowed is true.
	// If it is not set, http.Error with http.StatusMethodNotAllowed is used.
	// The "Allow" header with allowed request methods is set before the handler
	// is called.
	MethodNotAllowed Handler

	// Function to handle panics recovered from http handlers.
	// It should be used to generate a error page and return the http error code
	// 500 (Internal Server Error).
	// The handler can be used to keep your server from crashing because of
	// unrecovered panics.
	PanicHandler func(*Context, interface{})
}

// NewRouter returns a new initialized Router.
// Path auto-correction, including trailing slashes, is enabled by default.
func NewRouter() *Router {
	return &Router{
		RedirectTrailingSlash:  true,
		RedirectFixedPath:      true,
		HandleMethodNotAllowed: true,
		HandleOPTIONS:          true,
	}
}

// Use register middleware.
func (r *Router) Use(middleware Middleware) {
	r.middlewares = append(r.middlewares, middleware)
}

// GET is a shortcut for router.Handle("GET", path, handle)
func (r *Router) GET(path string, handle HandlerFunc, opts ...*HandlerOption) {
	r.Handle(MethodGet, path, handle, opts...)
}

// HEAD is a shortcut for router.Handle("HEAD", path, handle)
func (r *Router) HEAD(path string, handle HandlerFunc, opts ...*HandlerOption) {
	r.Handle(MethodHead, path, handle, opts...)
}

// OPTIONS is a shortcut for router.Handle("OPTIONS", path, handle)
func (r *Router) OPTIONS(path string, handle HandlerFunc, opts ...*HandlerOption) {
	r.Handle(MethodOptions, path, handle, opts...)
}

// POST is a shortcut for router.Handle("POST", path, handle)
func (r *Router) POST(path string, handle HandlerFunc, opts ...*HandlerOption) {
	r.Handle(MethodPost, path, handle, opts...)
}

// PUT is a shortcut for router.Handle("PUT", path, handle)
func (r *Router) PUT(path string, handle HandlerFunc, opts ...*HandlerOption) {
	r.Handle(MethodPut, path, handle, opts...)
}

// PATCH is a shortcut for router.Handle("PATCH", path, handle)
func (r *Router) PATCH(path string, handle HandlerFunc, opts ...*HandlerOption) {
	r.Handle(MethodPatch, path, handle, opts...)
}

// DELETE is a shortcut for router.Handle("DELETE", path, handle)
func (r *Router) DELETE(path string, handle HandlerFunc, opts ...*HandlerOption) {
	r.Handle(MethodDelete, path, handle, opts...)
}

// Handle registers a new request handle with the given path and method.
//
// For GET, POST, PUT, PATCH and DELETE requests the respective shortcut
// functions can be used.
//
// This function is intended for bulk loading and to allow the usage of less
// frequently used, non-standardized or custom methods (e.g. for internal
// communication with a proxy).
func (r *Router) Handle(method, path string, handle HandlerFunc, opts ...*HandlerOption) {
	if path[0] != '/' {
		panic("path must begin with '/' in path '" + path + "'")
	}

	if r.trees == nil {
		r.trees = make(map[string]*node)
	}

	root := r.trees[method]
	if root == nil {
		root = new(node)
		r.trees[method] = root
	}

	var handler Handler = HandlerFunc(handle)

	if len(opts) > 0 {
		// wrapped by middlewares.
		for i := len(opts[0].Middlewares) - 1; i >= 0; i-- {
			handler = opts[0].Middlewares[i].Wrap(handler)
		}
	}

	root.addRoute(path, handler)
}

// ServeFiles serves files from the given file system root.
// The path must end with "/*filepath", files are then served from the local
// path /defined/root/dir/*filepath.
// For example if root is "/etc" and *filepath is "passwd", the local file
// "/etc/passwd" would be served.
// Internally a http.FileServer is used, therefore http.NotFound is used instead
// of the Router's NotFound handler.
// To use the operating system's file system implementation,
// use http.Dir:
//     router.ServeFiles("/src/*filepath", http.Dir("/var/www"))
func (r *Router) ServeFiles(path string, root http.FileSystem, opts ...*HandlerOption) {
	if len(path) < 10 || path[len(path)-10:] != "/*filepath" {
		panic("path must end with /*filepath in path '" + path + "'")
	}

	fileServer := http.FileServer(root)

	handle := func(ctx *Context) {
		filepath, _ := ctx.UserValue("filepath").(string)
		ctx.Request.URL.Path = filepath
		fileServer.ServeHTTP(ctx.Response, ctx.Request)
	}

	r.GET(path, handle, opts...)
}

func (r *Router) recv(ctx *Context) {
	if rcv := recover(); rcv != nil {
		r.PanicHandler(ctx, rcv)
	}
}

// Lookup allows the manual lookup of a method + path combo.
// This is e.g. useful to build a framework around this router.
// If the path was found, it returns the handle function and the path parameter
// values. Otherwise the third return value indicates whether a redirection to
// the same path with an extra / without the trailing slash should be performed.
func (r *Router) Lookup(method, path string, ctx *Context) (Handler, bool) {
	if root := r.trees[method]; root != nil {
		return root.getValue(path, ctx)
	}
	return nil, false
}

func (r *Router) allowed(path, reqMethod string, ctx *Context) (allow string) {
	if path == "*" {
		// server-wide
		for method := range r.trees {
			if method == MethodOptions {
				continue
			}

			// add request method to list of allowed methods
			if len(allow) == 0 {
				allow = method
			} else {
				allow += ", " + method
			}
		}
	} else {
		// specific path
		for method := range r.trees {
			// Skip the requested method - we already tried this one
			if method == reqMethod || method == MethodOptions {
				continue
			}

			handle, _ := r.trees[method].getValue(path, ctx)
			if handle != nil {
				// add request method to list of allowed methods
				if len(allow) == 0 {
					allow = method
				} else {
					allow += ", " + method
				}
			}
		}
	}
	if len(allow) > 0 {
		allow += ", OPTIONS"
	}
	return
}

// Handler returns a Handler that wrapped by middlewarers.
func (r *Router) Handler() Handler {
	var handler Handler = HandlerFunc(r.handle)
	for i := len(r.middlewares) - 1; i >= 0; i-- {
		handler = r.middlewares[i].Wrap(handler)
	}

	return handler
}

func (r *Router) handle(ctx *Context) {
	if r.PanicHandler != nil {
		defer r.recv(ctx)
	}

	path := ctx.Request.URL.Path

	if root := r.trees[ctx.Request.Method]; root != nil {
		if handler, tsr := root.getValue(path, ctx); handler != nil {
			handler.Handle(ctx)
			return
		} else if ctx.Request.Method != MethodConnect && path != "/" {
			code := 301 // Permanent redirect, request with GET method
			if ctx.Request.Method != MethodGet {
				// Temporary redirect, request with same method
				// As of Go 1.3, Go does not support status code 308.
				code = 307
			}

			if tsr && r.RedirectTrailingSlash {
				if len(path) > 1 && path[len(path)-1] == '/' {
					ctx.Request.URL.Path = path[:len(path)-1]
				} else {
					ctx.Request.URL.Path = path + "/"
				}
				ctx.Redirect(ctx.Request.URL.String(), code)
				return
			}

			// Try to fix the request path
			if r.RedirectFixedPath {
				fixedPath, found := root.findCaseInsensitivePath(
					CleanPath(path),
					r.RedirectTrailingSlash,
				)
				if found {
					ctx.Request.URL.Path = string(fixedPath)
					ctx.Redirect(ctx.Request.URL.String(), code)
					return
				}
			}
		}
	}

	if ctx.Request.Method == MethodOptions {
		// Handle OPTIONS requests
		if r.HandleOPTIONS {
			if allow := r.allowed(path, ctx.Request.Method, ctx); len(allow) > 0 {
				ctx.Response.Header().Set("Allow", allow)
				return
			}
		}
	} else {
		// Handle 405
		if r.HandleMethodNotAllowed {
			if allow := r.allowed(path, ctx.Request.Method, ctx); len(allow) > 0 {
				ctx.Response.Header().Set("Allow", allow)
				if r.MethodNotAllowed != nil {
					r.MethodNotAllowed.Handle(ctx)
				} else {
					http.Error(ctx.Response,
						http.StatusText(http.StatusMethodNotAllowed),
						http.StatusMethodNotAllowed,
					)
				}
				return
			}
		}
	}

	// Handle 404
	if r.NotFound != nil {
		r.NotFound.Handle(ctx)
	} else {
		http.NotFound(ctx.Response, ctx.Request)
	}
}
