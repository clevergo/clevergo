// Copyright 2020 CleverGo. All rights reserved.
// Use of this source code is governed by a MIT license that can be found
// in the LICENSE file.

package clevergo

import (
	"errors"
	"fmt"
	"net/http"

	"clevergo.tech/log"
)

// Error defines an HTTP response error.
type Error interface {
	error
	Status() int
}

// Errors
var (
	ErrNotFound         = StatusError{http.StatusNotFound, errors.New(http.StatusText(http.StatusNotFound))}
	ErrMethodNotAllowed = StatusError{http.StatusMethodNotAllowed, errors.New(http.StatusText(http.StatusMethodNotAllowed))}
)

// ErrorHandlerOption is a function that receives an error handler instance.
type ErrorHandlerOption func(*errorHandler)

// ErrorHandlerLogger is an option that sets error handler logger.
func ErrorHandlerLogger(logger log.Logger) ErrorHandlerOption {
	return func(h *errorHandler) {
		h.logger = logger
	}
}

type errorHandler struct {
	logger log.Logger
}

func (h *errorHandler) middleware(next Handle) Handle {
	return func(c *Context) (err error) {
		if err := next(c); err != nil {
			h.handleError(c, err)
		}
		return nil
	}
}

func (h *errorHandler) handleError(c *Context, err error) {
	h.logger.Error(err)
	switch e := err.(type) {
	case Error:
		c.Error(e.Status(), err.Error())
	default:
		c.Error(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}
}

// ErrorHandler returns a error handler middleware with the given options.
func ErrorHandler(opts ...ErrorHandlerOption) MiddlewareFunc {
	h := &errorHandler{
		logger: defaultLogger,
	}
	for _, opt := range opts {
		opt(h)
	}
	return h.middleware
}

// StatusError implements Error interface.
type StatusError struct {
	Code int
	Err  error
}

// NewError returns a status error with the given code and error.
func NewError(code int, err error) StatusError {
	return StatusError{code, err}
}

// Error implements error.Error.
func (e StatusError) Error() string {
	return e.Err.Error()
}

// Status implements Error.Status.
func (e StatusError) Status() int {
	return e.Code
}

// PanicError is an error that contains panic infomation.
type PanicError struct {
	// Context.
	Context *Context

	// Recovery data.
	Data interface{}

	// Debug stack.
	Stack []byte
}

// Error implements error interface.
func (e PanicError) Error() string {
	return fmt.Sprintf("Panic: %v\n%s\n", e.Data, e.Stack)
}
