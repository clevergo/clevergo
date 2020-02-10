// Copyright 2020 CleverGo. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package clevergo

import (
	"log"
	"net"
	"net/http"
)

// Application application is a wrapper of Router and http.Server.
type Application struct {
	*Router
	*http.Server

	middlewares []MiddlewareFunc
	onCleanup   []func()
}

// New returns an application.
func New(addr string) *Application {
	return &Application{
		Server: &http.Server{
			Addr: addr,
		},
		Router: NewRouter(),
	}
}

// Use registers middlewares.
func (app *Application) Use(middlewares ...MiddlewareFunc) {
	app.middlewares = append(app.middlewares, middlewares...)
}

func (app *Application) prepare() {
	//app.Server.Handler = Chain(app.Router, app.middlewares...)
	app.Server.Handler = app.Router
}

// ListenAndServe overrides http.Server.ListenAndServe with extra preparations.
func (app *Application) ListenAndServe() error {
	app.prepare()
	return app.Server.ListenAndServe()
}

// ListenAndServeTLS overrides http.Server.ListenAndServeTLS with extra preparations.
func (app *Application) ListenAndServeTLS(certFile, keyFile string) error {
	app.prepare()
	return app.Server.ListenAndServeTLS(certFile, keyFile)
}

// ListenAndServeUnix listens on the Unix socket app.Server.Addr
// and then calls Serve to handle requests on incoming connections.
func (app *Application) ListenAndServeUnix() error {
	l, err := net.Listen("unix", app.Addr)
	if err != nil {
		return err
	}
	return app.Serve(l)
}

// Serve overrides http.Server.Serve with extra preparations.
func (app *Application) Serve(l net.Listener) error {
	app.prepare()
	return app.Server.Serve(l)
}

// ServeTLS overrides http.Server.ServeTLS with extra preparations.
func (app *Application) ServeTLS(l net.Listener, certFile, keyFile string) error {
	app.prepare()
	return app.Server.ServeTLS(l, certFile, keyFile)
}

// RegisterOnCleanup registers a function to call on CleanUp.
func (app *Application) RegisterOnCleanup(f func()) {
	cleanup := func() {
		defer app.cleanupRecover()
		f()
	}
	app.onCleanup = append(app.onCleanup, cleanup)
}

func (app *Application) cleanupRecover() {
	if r := recover(); r != nil {
		log.Println("recovered in cleanup", r)
	}
}

// Cleanup calls clean up functions before closing server.
func (app *Application) Cleanup() {
	for _, f := range app.onCleanup {
		f()
	}
}
