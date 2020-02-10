// Copyright 2020 CleverGo. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package clevergo

import "net/http"

type ErrorHandler interface {
	Handle(*Context, error)
}

type Error interface {
	error
	Status() int
}

type UserError struct {
	err    error
	status int
}

func (ue UserError) Error() string {
	return ue.err.Error()
}

func (ue UserError) Status() int {
	return ue.status
}

type PanicError struct {
}

func (pc PanicError) Status() int {
	return http.StatusInternalServerError
}

func (pc PanicError) Error() string {
	return ""
}

type NotFoundError struct {
}

func (nfe NotFoundError) Status() int {
	return http.StatusNotFound
}

func (nfe NotFoundError) Error() string {
	return "404 page not found"
}
