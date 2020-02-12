// Copyright 2020 CleverGo. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package clevergo

import (
	"errors"
	"net/http"
)

// Error defines an HTTP response error.
type Error interface {
	error
	Status() int
}

// Errors
var (
	ErrNotFound         = StatusError{http.StatusNotFound, errors.New("404 page not found")}
	ErrMethodNotAllowed = StatusError{http.StatusMethodNotAllowed, errors.New(http.StatusText(http.StatusMethodNotAllowed))}
)

// ErrorHandler is a handler to handle error returns from handle.
type ErrorHandler interface {
	Handle(ctx *Context, err error)
}

// StatusError implements Error interface.
type StatusError struct {
	Code int
	Err  error
}

// Error implements error.Error.
func (se StatusError) Error() string {
	return se.Err.Error()
}

// Status implements Error.Status.
func (se StatusError) Status() int {
	return se.Code
}
