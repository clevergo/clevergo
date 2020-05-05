// Copyright 2013 Julien Schmidt. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package clevergo

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

// errors
var (
	ErrRendererNotRegister = errors.New("renderer not registered")
	ErrDecoderNotRegister  = errors.New("decoder not registered")
)

var requestMethods = []string{
	http.MethodGet,
	http.MethodPost,
	http.MethodPut,
	http.MethodPatch,
	http.MethodOptions,
	http.MethodDelete,
	http.MethodHead,
	http.MethodConnect,
	http.MethodTrace,
}

// Handle is a function which handle incoming request and manage outgoing response.
type Handle func(ctx *Context) error

// HandleHandler converts http.Handler to Handle.
func HandleHandler(handler http.Handler) Handle {
	return func(ctx *Context) error {
		handler.ServeHTTP(ctx.Response, ctx.Request)
		return nil
	}
}

// HandleHandlerFunc converts http.HandlerFunc to Handle.
func HandleHandlerFunc(f http.HandlerFunc) Handle {
	return func(ctx *Context) error {
		f(ctx.Response, ctx.Request)
		return nil
	}
}

// Decoder is an interface that decodes request's input.
type Decoder interface {
	// Decode decodes request's input and stores it in the value pointed to by v.
	Decode(req *http.Request, v interface{}) error
}

// Renderer is an interface for template engine.
type Renderer interface {
	Render(w io.Writer, name string, data interface{}, ctx *Context) error
}

// Router is a http.Handler which can be used to dispatch requests to different
// handler functions via configurable routes
type Router struct {
	trees map[string]*node

	// Named routes.
	routes map[string]*Route

	maxParams uint16

	// Enables automatic redirection if the current route can't be matched but a
	// handler for the path with (without) the trailing slash exists.
	// For example if /foo/ is requested but a route only exists for /foo, the
	// client is redirected to /foo with http status code 301 for Get requests
	// and 308 for all other request methods.
	RedirectTrailingSlash bool

	// If enabled, the router tries to fix the current request path, if no
	// handle is registered for it.
	// First superfluous path elements like ../ or // are removed.
	// Afterwards the router does a case-insensitive lookup of the cleaned path.
	// If a handle can be found for this route, the router makes a redirection
	// to the corrected path with status code 301 for Get requests and 308 for
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

	// An optional http.Handler that is called on automatic OPTIONS requests.
	// The handler is only called if HandleOPTIONS is true and no OPTIONS
	// handler for the specific path was set.
	// The "Allowed" header is set before calling the handler.
	GlobalOPTIONS http.Handler

	// Cached value of global (*) allowed methods
	globalAllowed string

	// Configurable http.Handler which is called when no matching route is
	// found. If it is not set, http.NotFound is used.
	NotFound http.Handler

	// Configurable http.Handler which is called when a request
	// cannot be routed and HandleMethodNotAllowed is true.
	// If it is not set, http.Error with http.StatusMethodNotAllowed is used.
	// The "Allow" header with allowed request methods is set before the handler
	// is called.
	MethodNotAllowed http.Handler

	// Error Handler.
	ErrorHandler ErrorHandler

	// If enabled, use the request.URL.RawPath instead of request.URL.Path.
	UseRawPath bool

	middlewares []MiddlewareFunc
	handle      Handle

	Renderer Renderer

	Decoder Decoder
}

// Make sure the Router conforms with the http.Handler interface
var _ http.Handler = NewRouter()

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

// URL creates an url with the given route name and arguments.
func (r *Router) URL(name string, args ...string) (*url.URL, error) {
	if route, ok := r.routes[name]; ok {
		return route.URL(args...)
	}

	return nil, fmt.Errorf("route %q does not exist", name)
}

// Group implements IRouter.Group.
func (r *Router) Group(path string, opts ...RouteGroupOption) IRouter {
	return newRouteGroup(r, path, opts...)
}

// Use attaches global middlewares.
func (r *Router) Use(middlewares ...MiddlewareFunc) {
	r.middlewares = append(r.middlewares, middlewares...)

	r.handle = Chain(r.handleRequest, r.middlewares...)
}

// Get implements IRouter.Get.
func (r *Router) Get(path string, handle Handle, opts ...RouteOption) {
	r.Handle(http.MethodGet, path, handle, opts...)
}

// Head implements IRouter.Head.
func (r *Router) Head(path string, handle Handle, opts ...RouteOption) {
	r.Handle(http.MethodHead, path, handle)
}

// Options implements IRouter.Options.
func (r *Router) Options(path string, handle Handle, opts ...RouteOption) {
	r.Handle(http.MethodOptions, path, handle)
}

// Post implements IRouter.Post.
func (r *Router) Post(path string, handle Handle, opts ...RouteOption) {
	r.Handle(http.MethodPost, path, handle)
}

// Put implements IRouter.Put.
func (r *Router) Put(path string, handle Handle, opts ...RouteOption) {
	r.Handle(http.MethodPut, path, handle)
}

// Patch implements IRouter.Patch.
func (r *Router) Patch(path string, handle Handle, opts ...RouteOption) {
	r.Handle(http.MethodPatch, path, handle)
}

// Delete implements IRouter.Delete.
func (r *Router) Delete(path string, handle Handle, opts ...RouteOption) {
	r.Handle(http.MethodDelete, path, handle, opts...)
}

// Any implements IRouter.Any.
func (r *Router) Any(path string, handle Handle, opts ...RouteOption) {
	r.Handle(requestMethods[0], path, handle, opts...)
	// Removes route name option before registering handler by the rest of methods.
	for i, opt := range opts {
		if isRouteNameOption(opt) {
			opts = append(opts[:i], opts[i+1:]...)
		}
	}
	for i := 1; i < len(requestMethods); i++ {
		r.Handle(requestMethods[i], path, handle, opts...)
	}
}

// Handle implements IRouter.Handle.
func (r *Router) Handle(method, path string, handle Handle, opts ...RouteOption) {
	if method == "" {
		panic("method must not be empty")
	}
	if len(path) < 1 || path[0] != '/' {
		panic("path must begin with '/' in path '" + path + "'")
	}
	if handle == nil {
		panic("handle must not be nil")
	}
	if r.trees == nil {
		r.trees = make(map[string]*node)
	}

	root := r.trees[method]
	if root == nil {
		root = new(node)
		r.trees[method] = root

		r.globalAllowed = r.allowed("*", "")
	}

	route := newRoute(path, handle, opts...)
	if route.name != "" {
		if _, ok := r.routes[route.name]; ok {
			panic("route name " + route.name + " is already registered")
		}
		if r.routes == nil {
			r.routes = make(map[string]*Route)
		}
		r.routes[route.name] = route
	}
	root.addRoute(path, route)

	// Update maxParams
	if pc := countParams(path); pc > r.maxParams {
		r.maxParams = pc
	}
}

// Handler implements IRouter.Handler.
func (r *Router) Handler(method, path string, handler http.Handler, opts ...RouteOption) {
	r.Handle(method, path, HandleHandler(handler), opts...)
}

// HandlerFunc implements IRouter.HandlerFunc.
func (r *Router) HandlerFunc(method, path string, f http.HandlerFunc, opts ...RouteOption) {
	r.Handle(method, path, HandleHandlerFunc(f), opts...)
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
func (r *Router) ServeFiles(path string, root http.FileSystem, opts ...RouteOption) {
	if len(path) < 10 || path[len(path)-10:] != "/*filepath" {
		panic("path must end with /*filepath in path '" + path + "'")
	}

	fileServer := http.FileServer(root)

	r.Get(path, func(ctx *Context) error {
		ctx.Request.URL.Path = ctx.Params.String("filepath")
		fileServer.ServeHTTP(ctx.Response, ctx.Request)
		return nil
	}, opts...)
}

// Lookup allows the manual lookup of a method + path combo.
// This is e.g. useful to build a framework around this router.
// If the path was found, it returns the handle function and the path parameter
// values. Otherwise the third return value indicates whether a redirection to
// the same path with an extra / without the trailing slash should be performed.
func (r *Router) Lookup(method, path string) (*Route, Params, bool) {
	ps := make(Params, 0, r.maxParams)
	if root := r.trees[method]; root != nil {
		route, tsr := root.getValue(path, &ps, r.UseRawPath)
		return route, ps, tsr
	}
	return nil, nil, false
}

func (r *Router) allowed(path, reqMethod string) (allow string) {
	allowed := make([]string, 0, 9)

	if path == "*" { // server-wide
		// empty method is used for internal calls to refresh the cache
		if reqMethod == "" {
			for method := range r.trees {
				if method == http.MethodOptions {
					continue
				}
				// Add request method to list of allowed methods
				allowed = append(allowed, method)
			}
		} else {
			return r.globalAllowed
		}
	} else { // specific path
		for method := range r.trees {
			// Skip the requested method - we already tried this one
			if method == reqMethod || method == http.MethodOptions {
				continue
			}

			handle, _ := r.trees[method].getValue(path, nil, r.UseRawPath)
			if handle != nil {
				// Add request method to list of allowed methods
				allowed = append(allowed, method)
			}
		}
	}

	if len(allowed) > 0 {
		// Add request method to list of allowed methods
		allowed = append(allowed, http.MethodOptions)

		// Sort allowed methods.
		// sort.Strings(allowed) unfortunately causes unnecessary allocations
		// due to allowed being moved to the heap and interface conversion
		for i, l := 1, len(allowed); i < l; i++ {
			for j := i; j > 0 && allowed[j] < allowed[j-1]; j-- {
				allowed[j], allowed[j-1] = allowed[j-1], allowed[j]
			}
		}

		// return as comma separated list
		return strings.Join(allowed, ", ")
	}
	return
}

// ServeHTTP makes the router implement the http.Handler interface.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx := getContext(r, w, req)
	defer putContext(ctx)

	var err error
	if r.handle != nil {
		err = r.handle(ctx)
	} else {
		err = r.handleRequest(ctx)
	}
	if err != nil {
		r.HandleError(ctx, err)
	}
}

func (r *Router) handleRequest(ctx *Context) (err error) {
	path := ctx.Request.URL.Path
	if r.UseRawPath && ctx.Request.URL.RawPath != "" {
		path = ctx.Request.URL.RawPath
	}

	if root := r.trees[ctx.Request.Method]; root != nil {
		if route, tsr := root.getValue(path, &ctx.Params, r.UseRawPath); route != nil {
			ctx.Route = route
			err = route.handle(ctx)
			return
		} else if ctx.Request.Method != http.MethodConnect && path != "/" {
			// Moved Permanently, request with Get method
			code := http.StatusMovedPermanently
			if ctx.Request.Method != http.MethodGet {
				// Permanent Redirect, request with same method
				code = http.StatusPermanentRedirect
			}

			if tsr && r.RedirectTrailingSlash {
				if len(path) > 1 && path[len(path)-1] == '/' {
					path = path[:len(path)-1]
				} else {
					path = path + "/"
				}
				ctx.Redirect(path, code)
				return
			}

			// Try to fix the request path
			if r.RedirectFixedPath {
				fixedPath, found := root.findCaseInsensitivePath(
					CleanPath(path),
					r.RedirectTrailingSlash,
				)
				if found {
					ctx.Redirect(fixedPath, code)
					return
				}
			}
		}
	}

	if ctx.Request.Method == http.MethodOptions && r.HandleOPTIONS {
		// Handle OPTIONS requests
		if allow := r.allowed(path, http.MethodOptions); allow != "" {
			ctx.Response.Header().Set("Allow", allow)
			if r.GlobalOPTIONS != nil {
				r.GlobalOPTIONS.ServeHTTP(ctx.Response, ctx.Request)
			}
			return
		}
	} else if r.HandleMethodNotAllowed { // Handle 405
		if allow := r.allowed(path, ctx.Request.Method); allow != "" {
			ctx.Response.Header().Set("Allow", allow)
			if r.MethodNotAllowed != nil {
				r.MethodNotAllowed.ServeHTTP(ctx.Response, ctx.Request)
				return
			}
			return ErrMethodNotAllowed
		}
	}

	// Handle 404
	if r.NotFound != nil {
		r.NotFound.ServeHTTP(ctx.Response, ctx.Request)
		return
	}

	return ErrNotFound
}

// HandleError handles error.
func (r *Router) HandleError(ctx *Context, err error) {
	if r.ErrorHandler != nil {
		r.ErrorHandler.Handle(ctx, err)
		return
	}

	log.Println(err)

	switch e := err.(type) {
	case StatusError:
		ctx.Error(err.Error(), e.Status())
	default:
		ctx.Error(err.Error(), http.StatusInternalServerError)
	}
}
