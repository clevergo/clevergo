// Copyright 2016 The Gem Authors. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package gem

import (
	"net"
	"os"

	"github.com/go-gem/log"
	"github.com/go-gem/sessions"
	"github.com/valyala/fasthttp"
)

const (
	// Gem name
	name = "Gem"

	// Gem version
	version = "0.0.1"
)

var (
	defaultLogger = log.New(os.Stderr, log.LstdFlags, log.LevelAll)
)

// Name returns server name.
func Name() string {
	return name
}

// Version returns current version of Gem.
func Version() string {
	return version
}

// Middleware interface.
type Middleware interface {
	Handle(next Handler) Handler
}

// Server an extended edition of fasthttp.Server,
// see fasthttp.Server for details.
type Server struct {
	*fasthttp.Server

	// logger
	logger log.Logger

	// sessions store
	sessionsStore sessions.Store
}

// New returns a new Server instance.
func New() *Server {
	return &Server{
		Server: &fasthttp.Server{
			Name: name,
		},
		logger: defaultLogger,
	}
}

// SetLogger set logger.
func (s *Server) SetLogger(logger log.Logger) {
	s.logger = logger
}

// SetSessionStore set sessions store.
func (s *Server) SetSessionsStore(store sessions.Store) {
	s.sessionsStore = store
}

// Init for testing, should not invoke this method anyway.
func (s *Server) Init(handler HandlerFunc) {
	s.init(handler)
}

// init initialize server.
func (s *Server) init(handler HandlerFunc) {
	// Initialize fasthttp.Server's Handler.
	s.Server.Handler = func(ctx *fasthttp.RequestCtx) {
		c := acquireContext(s, ctx)
		defer c.close()
		handler(c)
	}
}

// ListenAndServe serves HTTP requests from the given TCP addr.
func (s *Server) ListenAndServe(addr string, handler HandlerFunc) error {
	s.init(handler)
	return s.Server.ListenAndServe(addr)
}

// ListenAndServeUNIX serves HTTP requests from the given UNIX addr.
//
// The function deletes existing file at addr before starting serving.
//
// The server sets the given file mode for the UNIX addr.
func (s *Server) ListenAndServeUNIX(addr string, mode os.FileMode, handler HandlerFunc) error {
	s.init(handler)
	return s.Server.ListenAndServeUNIX(addr, mode)
}

// ListenAndServeTLS serves HTTPS requests from the given TCP4 addr.
//
// certFile and keyFile are paths to TLS certificate and key files.
//
// Pass custom listener to Serve if you need listening on non-TCP4 media
// such as IPv6.
func (s *Server) ListenAndServeTLS(addr, certFile, keyFile string, handler HandlerFunc) error {
	s.init(handler)
	return s.Server.ListenAndServeTLS(addr, certFile, keyFile)
}

// ListenAndServeTLSEmbed serves HTTPS requests from the given TCP4 addr.
//
// certData and keyData must contain valid TLS certificate and key data.
//
// Pass custom listener to Serve if you need listening on arbitrary media
// such as IPv6.
func (s *Server) ListenAndServeTLSEmbed(addr string, certData, keyData []byte, handler HandlerFunc) error {
	s.init(handler)
	return s.Server.ListenAndServeTLSEmbed(addr, certData, keyData)
}

// Serve serves incoming connections from the given listener.
//
// Serve blocks until the given listener returns permanent error.
func (s *Server) Serve(ln net.Listener, handler HandlerFunc) error {
	s.init(handler)
	return s.Server.Serve(ln)
}

// ServeConn serves HTTP requests from the given connection.
//
// ServeConn returns nil if all requests from the c are successfully served.
// It returns non-nil error otherwise.
//
// Connection c must immediately propagate all the data passed to Write()
// to the client. Otherwise requests' processing may hang.
//
// ServeConn closes c before returning.
func (s *Server) ServeConn(c net.Conn, handler HandlerFunc) error {
	s.init(handler)
	return s.Server.ServeConn(c)
}

// ServeTLS serves HTTPS requests from the given net.Listener.
//
// certFile and keyFile are paths to TLS certificate and key files.
func (s *Server) ServeTLS(ln net.Listener, certFile, keyFile string, handler HandlerFunc) error {
	s.init(handler)
	return s.Server.ServeTLS(ln, certFile, keyFile)
}

// ServeTLSEmbed serves HTTPS requests from the given net.Listener.
//
// certData and keyData must contain valid TLS certificate and key data.
func (s *Server) ServeTLSEmbed(ln net.Listener, certData, keyData []byte, handler HandlerFunc) error {
	s.init(handler)
	return s.Server.ServeTLSEmbed(ln, certData, keyData)
}
