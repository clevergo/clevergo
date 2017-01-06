// Copyright 2016 The Gem Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license
// that can be found in the LICENSE file.

/*
Package gem is a high performance web framework, it is friendly to REST APIs.

Note: This package requires go1.8 or above.

Features

1. High performance

2. Friendly to REST API

3. Full test of all APIs

4. Pretty and fast router

5. HTTP/2 support

6. Leveled logging - included four levels `debug`, `info`, `error` and `fatal`, there are many third-party implements the Logger:

	logrus - Structured, pluggable logging for Go - https://github.com/Sirupsen/logrus
	go-logging - Golang logging library - https://github.com/op/go-logging
	gem-log - default logger - https://github.com/go-gem/log

7. Middlewares

	CSRF Middleware - Cross-Site Request Forgery protection - https://github.com/go-gem/middleware-csrf
	CORS Middleware -  Cross-Origin Resource Sharing - https://github.com/go-gem/middleware-cors
	AUTH Middleware - HTTP Basic and HTTP Digest authentication - https://github.com/go-gem/middleware-auth
	JWT Middleware - JSON WEB TOKEN authentication - https://github.com/go-gem/middleware-jwt
	Compress Middleware - compress response body - https://github.com/go-gem/middleware-compress
	Request Body Limit Middleware - limit request body size - https://github.com/go-gem/middleware-body-limit

8. Frozen APIs since the stable version `2.0.0` was released


Install

use go get command to install:

	$ go get -u github.com/go-gem/gem

Quick Start

a simple HTTP server:

	package main

	import (
	    "log"

	    "github.com/go-gem/gem"
	)

	func index(ctx *gem.Context) {
	    ctx.HTML(200, "hello world")
	}

	func main() {
	    // Create server.
	    srv := gem.New(":8080")

	    // Create router.
	    router := gem.NewRouter()
	    // Register handler
	    router.GET("/", index)

	    // Start server.
	    log.Println(srv.ListenAndServe(router.Handler()))
	}

Logger

AFAIK, the following leveled logging packages are compatible with Gem web framework:

1. logrus - structured, pluggable logging for Go - https://github.com/Sirupsen/logrus

2. go-logging - golang logging library - https://github.com/op/go-logging

3. gem-log - default logger, maintained by Gem Authors - https://github.com/go-gem/log

Logger(https://godoc.org/github.com/go-gem/gem#Logger) includes four levels: debug, info, error and fatal, their APIs are
Debug and Debugf, Info and Infof, Error and Errorf, Fatal and Fatalf.

We take logrus as example to show that how to set and use logger.

	// set logrus logger as server's logger.
	srv.SetLogger(logrus.New())

	// we can use it in handler.
	router.GET("/logger", func(ctx *gem.Context) {
			ctx.Logger().Debug("debug")
			ctx.Logger().Info("info")
			ctx.Logger().Error("error")
	})

Static Files

example that serve static files:

	router.ServeFiles("/tmp/*filepath", http.Dir(os.TempDir()))

Note: the path(first parameter) must end with `*filepath`.

REST APIs

The router is friendly to REST APIs.

	// user list
	router.GET("/users", func(ctx *gem.Context) {
	    ctx.JSON(200, userlist)
	})

	// add user
	router.POST("/users", func(ctx *gem.Contexy) {
	    ctx.Request.ParseForm()
	    name := ctx.Request.FormValue("name")

	    // add user

	    ctx.JSON(200, msg)
	})

	// user profile.
	router.GET("/users/:name", func(ctx *gem.Context) {
	    // firstly, we need get the username from the URL query.
	    name, ok := ctx.UserValue("name").(string)
	    if !ok {
		ctx.JSON(404, userNotFound)
		return
	    }

	    // return user profile.
	    ctx.JSON(200, userProfileByName(name))
	})

	// update user profile
	router.PUT("/users/:name", func(ctx *gem.Context) {
	    // firstly, we need get the username from the URL query.
	    name, ok := ctx.UserValue("name").(string)
	    if !ok {
		ctx.JSON(404, userNotFound)
		return
	    }

	    // get nickname
	    ctx.Request.ParseForm()
	    nickname := ctx.Request.FormValue("nickname")

	    // update user nickname.

	    ctx.JSON(200, msg)
	})

	// delete user
	router.DELETE("/users/:name", func(ctx *gem.Context) {
	    // firstly, we need get the username from the URL query.
	    name, ok := ctx.UserValue("name").(string)
	    if !ok {
		ctx.JSON(404, userNotFound)
		return
	    }

	    // delete user.

	    ctx.JSON(200, msg)
	}

HTTP2 Server Push

see https://github.com/go-gem/examples/tree/master/http2

	router.GET("/", func(ctx *gem.Context) {
		if err := ctx.Push("/images/logo.png", nil); err != nil {
			ctx.Logger().Info(err)
		}

		ctx.HTML(200, `<html><head></head><body><img src="/images/logo.png"/></body></html>`)
	})
	router.ServeFiles("/images/*filepath", http.Dir(imagesDir))

Use Middleware

It is easy to implement a middleware, see [Middleware](https://godoc.org/github.com/go-gem/gem#Middleware) interface,
you just need to implement the `Wrap` function.

	type Middleware interface {
	    Wrap(next Handler) Handler
	}

For example, we defined a simple debug middleware:

	type Debug struct{}

	// Wrap implements the Middleware interface.
	func (d *Debug) Wrap(next gem.Handler) gem.Handler {
		// gem.HandlerFunc is adapter like http.HandlerFunc.
		return gem.HandlerFunc(func(ctx *gem.Context) {
			// print request info.
			log.Println(ctx.Request.URL, ctx.Request.Method)

			// call the next handler.
			next.Handle(ctx)
		})
	}

and then we should register it:

register the middleware for all handlers via Router.Use(https://godoc.org/github.com/go-gem/gem#Router.Use).

	router.Use(&Debug{})

we can also register the middleware for specific handler via HandlerOption(https://godoc.org/github.com/go-gem/gem#HandlerOption).

	router.GET("/specific", specificHandler, &gem.HandlerOption{Middlewares:[]gem.Middleware{&Debug{}}})

Gem also provides some frequently used middlewares, such as:

1. CSRF Middleware - Cross-Site Request Forgery protection - https://github.com/go-gem/middleware-csrf

2. CORS Middleware -  Cross-Origin Resource Sharing - https://github.com/go-gem/middleware-cors

3. AUTH Middleware - HTTP Basic and HTTP Digest authentication - https://github.com/go-gem/middleware-auth

4. JWT Middleware - JSON WEB TOKEN authentication - https://github.com/go-gem/middleware-jwt

5. Compress Middleware - Compress response body - https://github.com/go-gem/middleware-compress

6. Request Body Limit Middleware - limit request body maximum size - https://github.com/go-gem/middleware-body-limit


Share data between middlewares

Context provides two useful methods: `SetUserValue` and `UserValue` to share data between middlewares.

	// Store data into context in one middleware
	ctx.SetUserValue("name", "foo")

	// Get data from context in other middleware or hander
	ctx.UserValue("name")
*/
package gem
