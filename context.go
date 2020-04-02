// Copyright 2020 CleverGo. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package clevergo

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
)

// Context contains incoming request, route, params and manages outgoing response.
type Context struct {
	Params   Params
	Route    *Route
	Request  *http.Request
	Response http.ResponseWriter
}

func newContext(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		Request:  r,
		Response: w,
	}
}

func (ctx *Context) reset() {
	ctx.Params = nil
	ctx.Route = nil
	ctx.Response = nil
	ctx.Request = nil
}

// Error is a shortcut of http.Error.
func (ctx *Context) Error(msg string, code int) {
	http.Error(ctx.Response, msg, code)
}

// NotFound is a shortcut of http.NotFound.
func (ctx *Context) NotFound() {
	http.NotFound(ctx.Response, ctx.Request)
}

// Redirect is a shortcut of http.Redirect.
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

// SetCookie is a shortcut of http.SetCookie.
func (ctx *Context) SetCookie(cookie *http.Cookie) {
	http.SetCookie(ctx.Response, cookie)
}

// Write is a shortcut of http.ResponseWriter.Write.
func (ctx *Context) Write(data []byte) (int, error) {
	return ctx.Response.Write(data)
}

// WriteString writes the string data to response.
func (ctx *Context) WriteString(data string) (int, error) {
	return io.WriteString(ctx.Response, data)
}

// WriteHeader is a shortcut of http.ResponseWriter.WriteHeader.
func (ctx *Context) WriteHeader(code int) {
	ctx.Response.WriteHeader(code)
}

// WithValue stores the given value under the given key.
func (ctx *Context) WithValue(key, val interface{}) {
	ctx.Request = ctx.Request.WithContext(context.WithValue(ctx.Request.Context(), key, val))
}

// Value returns the value of the given key.
func (ctx *Context) Value(key interface{}) interface{} {
	return ctx.Request.Context().Value(key)
}

// IsMethod returns a boolean value indicates whether the request method is the given method.
func (ctx *Context) IsMethod(method string) bool {
	return ctx.Request.Method == method
}

// IsAJAX indicates whether it is an AJAX (XMLHttpRequest) request.
func (ctx *Context) IsAJAX() bool {
	return ctx.Request.Header.Get("X-Requested-With") == "XMLHttpRequest"
}

// IsDelete returns a boolean value indicates whether the request method is DELETE.
func (ctx *Context) IsDelete() bool {
	return ctx.IsMethod(http.MethodDelete)
}

// IsGet returns a boolean value indicates whether the request method is GET.
func (ctx *Context) IsGet() bool {
	return ctx.IsMethod(http.MethodGet)
}

// IsOptions returns a boolean value indicates whether the request method is OPTIONS.
func (ctx *Context) IsOptions() bool {
	return ctx.IsMethod(http.MethodOptions)
}

// IsPatch returns a boolean value indicates whether the request method is PATCH.
func (ctx *Context) IsPatch() bool {
	return ctx.IsMethod(http.MethodPatch)
}

// IsPost returns a boolean value indicates whether the request method is POST.
func (ctx *Context) IsPost() bool {
	return ctx.IsMethod(http.MethodPost)
}

// IsPut returns a boolean value indicates whether the request method is PUT.
func (ctx *Context) IsPut() bool {
	return ctx.IsMethod(http.MethodPut)
}

// GetHeader is a shortcut of http.Request.Header.Get.
func (ctx *Context) GetHeader(name string) string {
	return ctx.Request.Header.Get(name)
}

// JSON sends JSON response with status code, it also sets
// Content-Type as "application/json".
func (ctx *Context) JSON(code int, data interface{}) error {
	bs, err := json.Marshal(data)
	if err != nil {
		return err
	}

	ctx.SetContentTypeJSON()
	ctx.Response.WriteHeader(code)
	_, err = ctx.Response.Write(bs)
	return err
}

// String send string response with status code, it also sets
// Content-Type as "text/plain; charset=utf-8".
func (ctx *Context) String(code int, s string) error {
	ctx.SetContentTypeText()
	ctx.Response.WriteHeader(code)
	_, err := ctx.WriteString(s)
	return err
}

// XML sends XML response with status code, it also sets
// Content-Type as "application/xml".
func (ctx *Context) XML(code int, data interface{}) error {
	bs, err := xml.Marshal(data)
	if err != nil {
		return err
	}

	ctx.SetContentTypeXML()
	ctx.Response.WriteHeader(code)
	_, err = ctx.Response.Write(bs)
	return err
}
