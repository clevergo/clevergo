// Copyright 2016 The Gem Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license
// that can be found in the LICENSE file.

package gem

import (
	"net/http"
)

const version = "2.0.0"

// Version return version number.
func Version() string {
	return version
}

// New return a Server instance by the given address.
func New(addr string) *Server {
	return &Server{
		Server: &http.Server{
			Addr: addr,
		},
		logger: defaultLogger,
	}
}

// Server contains *http.Server.
type Server struct {
	Server *http.Server
	logger Logger
}

// SetLogger set logger.
func (srv *Server) SetLogger(logger Logger) {
	srv.logger = logger
}

// ListenAndServe listens on the TCP network address srv.Addr and then
// calls Serve to handle requests on incoming connections.
func (srv *Server) ListenAndServe(handler Handler) error {
	srv.init(handler)

	return srv.Server.ListenAndServe()
}

// ListenAndServeTLS listens on the TCP network address srv.Addr and
// then calls Serve to handle requests on incoming TLS connections.
// Accepted connections are configured to enable TCP keep-alives.
func (srv *Server) ListenAndServeTLS(certFile, keyFile string, handler Handler) error {
	srv.init(handler)

	return srv.Server.ListenAndServeTLS(certFile, keyFile)
}

func (srv *Server) init(handler Handler) {
	srv.Server.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := newContext(srv, w, r)
		handler.Handle(ctx)
	})
}

// ListenAndServe listens on the TCP network address addr
// and then calls Serve with handler to handle requests
// on incoming connections.
func ListenAndServe(addr string, handler Handler) error {
	srv := New(addr)

	return srv.ListenAndServe(handler)
}

// ListenAndServeTLS acts identically to ListenAndServe, except that it
// expects HTTPS connections. Additionally, files containing a certificate and
// matching private key for the server must be provided. If the certificate
// is signed by a certificate authority, the certFile should be the concatenation
// of the server's certificate, any intermediates, and the CA's certificate.
func ListenAndServeTLS(addr, certFile, keyFile string, handler Handler) error {
	srv := New(addr)

	return srv.ListenAndServeTLS(certFile, keyFile, handler)
}
