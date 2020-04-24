// Copyright 2020 CleverGo. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package clevergo

import "net/http"

// RouteGroupOption applies options to a route group.
type RouteGroupOption func(*RouteGroup)

// RouteGroupName set the name of route group.
func RouteGroupName(name string) RouteGroupOption {
	return func(r *RouteGroup) {
		r.name = name
	}
}

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
	name        string
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

	route := &RouteGroup{parent: parent, path: path, name: path}
	for _, opt := range opts {
		opt(route)
	}

	return route
}

// Group implements IRouter.Group.
func (r *RouteGroup) Group(path string, opts ...RouteGroupOption) IRouter {
	router := newRouteGroup(r.parent, r.subPath(path), opts...)

	// inherit middlewares.
	router.middlewares = append(r.middlewares, router.middlewares...)

	return router
}

func (r *RouteGroup) nameOption() RouteOption {
	return func(route *Route) {
		if route.name != "" {
			route.name = r.name + "/" + route.name
		}
	}
}

func (r RouteGroup) middlewareOption() RouteOption {
	return func(route *Route) {
		if len(r.middlewares) > 0 {
			route.handle = Chain(route.handle, r.middlewares...)
		}
	}
}

func (r *RouteGroup) combineOptions(opts []RouteOption) []RouteOption {
	opts = append(opts, r.nameOption(), r.middlewareOption())
	return opts
}

// Handle implements IRouter.Handle.
func (r *RouteGroup) Handle(method, path string, handle Handle, opts ...RouteOption) {
	r.parent.Handle(method, r.subPath(path), handle, r.combineOptions(opts)...)
}

// Handler implements IRouter.Handler.
func (r *RouteGroup) Handler(method, path string, handler http.Handler, opts ...RouteOption) {
	r.Handle(method, path, HandleHandler(handler), opts...)
}

// HandlerFunc implements IRouter.HandlerFunc.
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

// Any implements IRouter.Any.
func (r *RouteGroup) Any(path string, handle Handle, opts ...RouteOption) {
	r.parent.Any(r.subPath(path), handle, r.combineOptions(opts)...)
}

func (r *RouteGroup) subPath(path string) string {
	return r.path + path
}
