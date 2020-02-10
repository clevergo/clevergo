// Copyright 2020 CleverGo. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package clevergo

import (
	"io"
	"net/http"
)

type Context struct {
	Params   Params
	Route    *Route
	Request  *http.Request
	Response http.ResponseWriter
}

func NewContext(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		Request:  r,
		Response: w,
	}
}

func (ctx *Context) Error(msg string, code int) {
	http.Error(ctx.Response, msg, code)
}

func (ctx *Context) NotFound() {
	http.NotFound(ctx.Response, ctx.Request)
}

func (ctx *Context) Redirect(url string, code int) {
	http.Redirect(ctx.Response, ctx.Request, url, code)
}

// SetContentType sets the content type header.
func (ctx *Context) SetContentType(v string) {
	ctx.Response.Header().Set("Content-Type", v)
}

// SetContentTypeHTML sets the content type as HTML.
func (ctx *Context) SetContentTypeHTML() {
	ctx.SetContentType("text/html; charset=utf-8")
}

// SetContentTypeText sets the content type as text.
func (ctx *Context) SetContentTypeText() {
	ctx.SetContentType("text/plain; charset=utf-8")
}

// SetContentTypeJSON sets the content type as JSON.
func (ctx *Context) SetContentTypeJSON() {
	ctx.SetContentType("application/json")
}

// SetContentTypeXML sets the content type as XML.
func (ctx *Context) SetContentTypeXML() {
	ctx.SetContentType("application/xml")
}

func (ctx *Context) Write(data []byte) (int, error) {
	return ctx.Response.Write(data)
}

func (ctx *Context) WriteString(data string) (int, error) {
	return io.WriteString(ctx.Response, data)
}
