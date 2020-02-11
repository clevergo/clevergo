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

	// Handler is an adapter for registering http.Handler.
	Handler(method, path string, handler http.Handler, opts ...RouteOption)

	// HandlerFunc is an adapter for registering http.HandlerFunc.
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
