// Copyright 2020 CleverGo. All rights reserved.
// Use of this source code is governed by a MIT style license that can be found
// in the LICENSE file.

package clevergo

import (
	"bytes"
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
		logger: logger,
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
		resp := newBufferedResponse(c.Response)
		defer func(w http.ResponseWriter) {
			if err := resp.emit(); err != nil {
				c.Logger().Errorf("clevergo: logging middleware failed to send buffered response: %s", err.Error())
			}
			l.print(c.Request, resp)
			c.Response = w
		}(c.Response)
		c.Response = resp
		return next(c)
	}
}

func (l *logging) print(req *http.Request, resp *bufferedResponse) {
	l.logger.Infof("| %d | %-10s | %s %s %s", resp.statusCode, resp.duration(), req.Method, req.RequestURI, req.Proto)
}

type bufferedResponse struct {
	http.ResponseWriter
	wroteHeader bool
	statusCode  int
	buf         bytes.Buffer
	start       time.Time
}

func newBufferedResponse(w http.ResponseWriter) *bufferedResponse {
	return &bufferedResponse{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
		start:          time.Now(),
	}
}

func (resp *bufferedResponse) WriteHeader(statusCode int) {
	if !resp.wroteHeader {
		resp.wroteHeader = true
		resp.statusCode = statusCode
		resp.ResponseWriter.WriteHeader(statusCode)
	}
}

func (resp *bufferedResponse) Write(p []byte) (int, error) {
	return resp.buf.Write(p)
}

func (resp *bufferedResponse) WriteString(s string) (int, error) {
	return resp.buf.WriteString(s)
}

func (resp *bufferedResponse) duration() time.Duration {
	return time.Since(resp.start)
}

func (resp *bufferedResponse) emit() error {
	_, err := resp.ResponseWriter.Write(resp.buf.Bytes())
	return err
}
