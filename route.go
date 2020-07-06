// Copyright 2020 CleverGo. All rights reserved.
// Use of this source code is governed by a MIT style license that can be found
// in the LICENSE file.

package clevergo

import (
	"errors"
	"fmt"
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

func isRouteNameOption(opt RouteOption) bool {
	r := &Route{}
	opt(r)
	return r.name != ""
}

// RouteMiddleware is a route option for chainging middlewares to a route.
func RouteMiddleware(middlewares ...MiddlewareFunc) RouteOption {
	return func(r *Route) {
		r.handle = Chain(r.handle, middlewares...)
	}
}
