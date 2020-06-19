// Copyright 2020 CleverGo. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package clevergo

import (
	"net/http"
	"time"

	"clevergo.tech/log"
)

// LoggingOption is a function that receives a logging instance.
type LoggingOption func(*logging)

// LoggingLogger is an option that sets logging logger.
func LoggingLogger(logger log.Logger) LoggingOption {
	return func(l *logging) {
		l.logger = logger
	}
}

// Logging returns a logging middleware with the given options.
func Logging(opts ...LoggingOption) MiddlewareFunc {
	l := &logging{
		logger: defaultLogger,
	}
	for _, opt := range opts {
		opt(l)
	}
	return l.middleware
}

type logging struct {
	logger log.Logger
}

func (l *logging) middleware(next Handle) Handle {
	return func(c *Context) error {
		resp := newLoggingResponse(c.Response)
		defer func(w http.ResponseWriter) {
			l.print(c.Request, resp)
			c.Response = w
		}(c.Response)
		c.Response = resp
		return next(c)
	}
}

func (l *logging) print(req *http.Request, resp *loggingResponse) {
	l.logger.Infof("| %d | %-10s | %s %s %s", resp.code, resp.duration(), req.Method, req.RequestURI, req.Proto)
}

type loggingResponse struct {
	http.ResponseWriter
	code  int
	start time.Time
}

func newLoggingResponse(w http.ResponseWriter) *loggingResponse {
	return &loggingResponse{
		ResponseWriter: w,
		code:           http.StatusOK,
		start:          time.Now(),
	}
}

func (resp *loggingResponse) WriteHeader(code int) {
	resp.code = code
	resp.ResponseWriter.WriteHeader(code)
}

func (resp *loggingResponse) duration() time.Duration {
	return time.Since(resp.start)
}
