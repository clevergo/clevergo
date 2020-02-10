// Copyright 2020 CleverGo. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package clevergo

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

// IRouter is an router interface.
type IRouter interface {
	// Group creates a sub IRouter with the given optional route group options.
	Group(path string, opts ...RouteGroupOption) IRouter

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

	// HandleFunc registers a new request handler function with the given path, method and optional route options.
	//
	// For Get, Head, Options, Post, Put, Patch and Delete requests the respective shortcut
	// functions can be used.
	//
	// This function is intended for bulk loading and to allow the usage of less
	// frequently used, non-standardized or custom methods (e.g. for internal
	// communication with a proxy).
	Handle(method, path string, handle Handle, opts ...RouteOption)

	Handler(method, path string, handler http.Handler, opts ...RouteOption)

	HandlerFunc(method, path string, f http.HandlerFunc, opts ...RouteOption)
}

var routeParamRegexp = regexp.MustCompile(`([\:|\*])([^\:\*\/]+)`)

// Route is a HTTP request handler.
type Route struct {
	path    string
	name    string
	pattern string
	params  []routeParam
	handle  Handle
}

func newRoute(path string, handle Handle, opts ...RouteOption) *Route {
	r := &Route{
		path:    path,
		pattern: path,
		handle:  handle,
	}
	for _, opt := range opts {
		opt(r)
	}
	r.parse()
	return r
}

func (r *Route) parse() {
	matchs := routeParamRegexp.FindAllStringSubmatch(r.path, -1)
	if len(matchs) == 0 {
		return
	}

	for _, match := range matchs {
		r.params = append(r.params, routeParam{
			name:     match[2],
			required: match[1] == ":",
		})
		r.pattern = strings.Replace(r.pattern, match[0], "{"+match[2]+"}", 1)
	}
}

var errWrongArgumentsNumber = errors.New("wrong number of arguments")

// URL creates an url with the given arguments.
//
// It accepts a sequence of key/value pairs for the route variables,
// otherwise errWrongArgumentsNumber will be returned.
func (r *Route) URL(args ...string) (*url.URL, error) {
	if len(args)%2 != 0 {
		return nil, errWrongArgumentsNumber
	}

	path := r.pattern
	var value string
	for _, param := range r.params {
		value = ""
		for i := 0; i < len(args)-1; i++ {
			if args[i] == param.name {
				value = args[i+1]
				break
			}
		}
		if param.required && value == "" {
			return nil, fmt.Errorf("route %q parameter %q is required", r.name, param.name)
		}

		path = strings.Replace(path, "{"+param.name+"}", value, 1)
	}

	return &url.URL{
		Path: path,
	}, nil
}

type routeParam struct {
	name     string
	required bool
}

// RouteOption applies options to a route,
// see RouteName and RouteMiddleware.
type RouteOption func(*Route)

// RouteName is a route option for naming a route.
func RouteName(name string) RouteOption {
	return func(r *Route) {
		r.name = name
	}
}

// RouteMiddleware is a route option for chainging middlewares to a route.
func RouteMiddleware(middlewares ...MiddlewareFunc) RouteOption {
	return func(r *Route) {
		r.handle = Chain(r.handle, middlewares...)
	}
}

// RouteGroupOption applies options to a route group.
type RouteGroupOption func(*RouteGroup)

// RouteGroupMiddleware is a option for chainging middlewares to a route group.
func RouteGroupMiddleware(middlewares ...MiddlewareFunc) RouteGroupOption {
	return func(r *RouteGroup) {
		r.middlewares = append(r.middlewares, middlewares...)
	}
}

// RouteGroup implements an nested route group,
// see https://github.com/julienschmidt/httprouter/pull/89.
type RouteGroup struct {
	parent      *Router
	path        string
	middlewares []MiddlewareFunc
}

func newRouteGroup(parent *Router, path string, opts ...RouteGroupOption) *RouteGroup {
	if path[0] != '/' {
		panic("path must begin with '/' in path '" + path + "'")
	}

	// strips traling / (if present) as all added sub paths must start with a "/".
	if len(path) > 1 && path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}

	route := &RouteGroup{parent: parent, path: path}
	for _, opt := range opts {
		opt(route)
	}

	return route
}

// Group implements IRouter.Group.
func (r *RouteGroup) Group(path string, opts ...RouteGroupOption) IRouter {
	return newRouteGroup(r.parent, r.subPath(path), opts...)
}

// Handle implements IRouter.Handle.
func (r *RouteGroup) Handle(method, path string, handle Handle, opts ...RouteOption) {
	handle = Chain(handle, r.middlewares...)

	opts = append(opts, func(route *Route) {
		if route.name != "" {
			route.name = r.path + "/" + route.name
		}
	})
	r.parent.Handle(method, r.subPath(path), handle, opts...)
}

func (r *RouteGroup) Handler(method, path string, handler http.Handler, opts ...RouteOption) {
	r.Handle(method, path, HandleHandler(handler), opts...)
}

func (r *RouteGroup) HandlerFunc(method, path string, f http.HandlerFunc, opts ...RouteOption) {
	r.Handle(method, path, HandleHandlerFunc(f), opts...)
}

// Get implements IRouter.Get.
func (r *RouteGroup) Get(path string, handle Handle, opts ...RouteOption) {
	r.Handle(http.MethodGet, path, handle, opts...)
}

// Head implements IRouter.Head.
func (r *RouteGroup) Head(path string, handle Handle, opts ...RouteOption) {
	r.Handle(http.MethodHead, path, handle, opts...)
}

// Options implements IRouter.Options.
func (r *RouteGroup) Options(path string, handle Handle, opts ...RouteOption) {
	r.Handle(http.MethodOptions, path, handle, opts...)
}

// Post implements IRouter.Post.
func (r *RouteGroup) Post(path string, handle Handle, opts ...RouteOption) {
	r.Handle(http.MethodPost, path, handle, opts...)
}

// Put implements IRouter.Put.
func (r *RouteGroup) Put(path string, handle Handle, opts ...RouteOption) {
	r.Handle(http.MethodPut, path, handle, opts...)
}

// Patch implements IRouter.Patch.
func (r *RouteGroup) Patch(path string, handle Handle, opts ...RouteOption) {
	r.Handle(http.MethodPatch, path, handle, opts...)
}

// Delete implements IRouter.Delete.
func (r *RouteGroup) Delete(path string, handle Handle, opts ...RouteOption) {
	r.Handle(http.MethodDelete, path, handle, opts...)
}

func (r *RouteGroup) subPath(path string) string {
	return r.path + path
}
