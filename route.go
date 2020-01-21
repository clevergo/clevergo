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

var routeParamRegexp = regexp.MustCompile(`([\:|\*])([^\:\*\/]+)`)

// Route is a HTTP request handler.
type Route struct {
	path    string
	name    string
	pattern string
	params  []routeParam
	handler http.Handler
}

func newRoute(path string, handler http.Handler, opts ...RouteOption) *Route {
	r := &Route{
		path:    path,
		pattern: path,
		handler: handler,
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
func RouteMiddleware(middlewares ...Middleware) RouteOption {
	return func(r *Route) {
		r.handler = Chain(r.handler, middlewares...)
	}
}

// RouteGroupOption applies options to a route group.
type RouteGroupOption func(*RouteGroup)

// RouteGroupMiddleware is a option for chainging middlewares to a route group.
func RouteGroupMiddleware(middlewares ...Middleware) RouteGroupOption {
	return func(r *RouteGroup) {
		r.middlewares = append(r.middlewares, middlewares...)
	}
}

// RouteGroup implements an nested route group,
// see https://github.com/julienschmidt/httprouter/pull/89.
type RouteGroup struct {
	parent      *Router
	path        string
	middlewares []Middleware
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

// Group creates route group with the given path and optional route options.
func (r *RouteGroup) Group(path string, opts ...RouteGroupOption) *RouteGroup {
	return newRouteGroup(r.parent, r.subPath(path), opts...)
}

// HandleFunc is a shortcut of RouteGroup.Handle(http.MethodDelete, path, http.HandlerFunc(handle), opts ...)
func (r *RouteGroup) HandleFunc(method, path string, handle http.HandlerFunc, opts ...RouteOption) {
	r.Handle(method, path, http.HandlerFunc(handle), opts...)
}

// Handle registers a new request handler with the given path, method and optional route options.
func (r *RouteGroup) Handle(method, path string, handler http.Handler, opts ...RouteOption) {
	handler = Chain(handler, r.middlewares...)

	opts = append(opts, func(route *Route) {
		if route.name != "" {
			route.name = r.path + "/" + route.name
		}
	})
	r.parent.Handle(method, r.subPath(path), handler, opts...)
}

// Get is a shortcut of RouteGroup.HandleFunc(http.MethodGet, path, handle, opts ...)
func (r *RouteGroup) Get(path string, handle http.HandlerFunc, opts ...RouteOption) {
	r.HandleFunc(http.MethodGet, path, handle, opts...)
}

// Head is a shortcut of RouteGroup.HandleFunc(http.MethodHead, path, handle, opts ...)
func (r *RouteGroup) Head(path string, handle http.HandlerFunc, opts ...RouteOption) {
	r.HandleFunc(http.MethodHead, path, handle, opts...)
}

// Options is a shortcut of RouteGroup.HandleFunc(http.MethodOptions, path, handle, opts ...)
func (r *RouteGroup) Options(path string, handle http.HandlerFunc, opts ...RouteOption) {
	r.HandleFunc(http.MethodOptions, path, handle, opts...)
}

// Post is a shortcut of RouteGroup.HandleFunc(http.MethodPost, path, handle, opts ...)
func (r *RouteGroup) Post(path string, handle http.HandlerFunc, opts ...RouteOption) {
	r.HandleFunc(http.MethodPost, path, handle, opts...)
}

// Put is a shortcut of RouteGroup.HandleFunc(http.MethodPut, path, handle, opts ...)
func (r *RouteGroup) Put(path string, handle http.HandlerFunc, opts ...RouteOption) {
	r.HandleFunc(http.MethodPut, path, handle, opts...)
}

// Patch is a shortcut of RouteGroup.HandleFunc(http.MethodPatch, path, handle, opts ...)
func (r *RouteGroup) Patch(path string, handle http.HandlerFunc, opts ...RouteOption) {
	r.HandleFunc(http.MethodPatch, path, handle, opts...)
}

// Delete is a shortcut of RouteGroup.HandleFunc(http.MethodDelete, path, handle, opts ...)
func (r *RouteGroup) Delete(path string, handle http.HandlerFunc, opts ...RouteOption) {
	r.HandleFunc(http.MethodDelete, path, handle, opts...)
}

func (r *RouteGroup) subPath(path string) string {
	return r.path + path
}
