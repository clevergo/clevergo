// Copyright 2020 CleverGo. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package clevergo

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const (
	headerContentType           = "Content-Type"
	headerContentTypeHTML       = "text/html; charset=utf-8"
	headerContentTypeJavaScript = "application/javascript; charset=utf-8"
	headerContentTypeJSON       = "application/json; charset=utf-8"
	headerContentTypeText       = "text/plain; charset=utf-8"
	headerContentTypeXML        = "application/xml; charset=utf-8"
)

var bufPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

func getBuffer() (buf *bytes.Buffer) {
	buf, _ = bufPool.Get().(*bytes.Buffer)
	buf.Reset()
	return
}

func putBuffer(buf *bytes.Buffer) {
	bufPool.Put(buf)
}

var contextPool = sync.Pool{
	New: func() interface{} {
		return &Context{}
	},
}

func getContext() *Context {
	ctx := contextPool.Get().(*Context)
	ctx.reset()
	return ctx
}

func putContext(ctx *Context) {
	contextPool.Put(ctx)
}

// Context contains incoming request, route, params and manages outgoing response.
type Context struct {
	router   *Router
	Params   Params
	Route    *Route
	Request  *http.Request
	Response http.ResponseWriter
	query    url.Values
}

func newContext(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		Request:  r,
		Response: w,
	}
}

func (ctx *Context) reset() {
	ctx.router = nil
	ctx.Params = nil
	ctx.Route = nil
	ctx.Response = nil
	ctx.Request = nil
	ctx.query = nil
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

// ServeFile is a shortcut of http.ServeFile.
func (ctx *Context) ServeFile(name string) {
	http.ServeFile(ctx.Response, ctx.Request, name)
}

// ServeContent is a shortcut of http.ServeContent.
func (ctx *Context) ServeContent(name string, modtime time.Time, content io.ReadSeeker) {
	http.ServeContent(ctx.Response, ctx.Request, name, modtime, content)
}

// SetContentType sets the content type header.
func (ctx *Context) SetContentType(v string) {
	ctx.Response.Header().Set("Content-Type", v)
}

// SetContentTypeHTML sets the content type as HTML.
func (ctx *Context) SetContentTypeHTML() {
	ctx.SetContentType(headerContentTypeHTML)
}

// SetContentTypeText sets the content type as text.
func (ctx *Context) SetContentTypeText() {
	ctx.SetContentType(headerContentTypeText)
}

// SetContentTypeJSON sets the content type as JSON.
func (ctx *Context) SetContentTypeJSON() {
	ctx.SetContentType(headerContentTypeJSON)
}

// SetContentTypeXML sets the content type as XML.
func (ctx *Context) SetContentTypeXML() {
	ctx.SetContentType(headerContentTypeXML)
}

// Cookie is a shortcut of http.Request.Cookie.
func (ctx *Context) Cookie(name string) (*http.Cookie, error) {
	return ctx.Request.Cookie(name)
}

// Cookies is a shortcut of http.Request.Cookies.
func (ctx *Context) Cookies() []*http.Cookie {
	return ctx.Request.Cookies()
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
	return ctx.Blob(code, headerContentTypeJSON, bs)
}

// JSONBlob sends blob JSON response with status code, it also sets
// Content-Type as "application/json".
func (ctx *Context) JSONBlob(code int, bs []byte) error {
	return ctx.Blob(code, headerContentTypeJSON, bs)
}

// JSONP is a shortcut of JSONPCallback with specified callback param name.
func (ctx *Context) JSONP(code int, data interface{}) error {
	return ctx.JSONPCallback(code, "callback", data)
}

// JSONPCallback sends JSONP response with status code, it also sets
// Content-Type as "application/javascript".
// If the callback is not present, returns JSON response instead.
func (ctx *Context) JSONPCallback(code int, callback string, data interface{}) error {
	fn := ctx.QueryParam(callback)
	if fn == "" {
		return ctx.JSON(code, data)
	}

	bs, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return ctx.Emit(code, headerContentTypeJavaScript, formatJSONP(fn, bs))
}

// JSONPBlob is a shortcut of JSONPCallbackBlob with specified callback param name.
func (ctx *Context) JSONPBlob(code int, bs []byte) error {
	return ctx.JSONPCallbackBlob(code, "callback", bs)
}

// JSONPCallbackBlob sends blob JSONP response with status code, it also sets
// Content-Type as "application/javascript".
// If the callback is not present, returns JSON response instead.
func (ctx *Context) JSONPCallbackBlob(code int, callback string, bs []byte) (err error) {
	fn := ctx.QueryParam(callback)
	if fn == "" {
		return ctx.JSONBlob(code, bs)
	}

	return ctx.Emit(code, headerContentTypeJavaScript, formatJSONP(fn, bs))
}

func formatJSONP(callback string, bs []byte) string {
	return callback + "(" + string(bs) + ")"
}

// String send string response with status code, it also sets
// Content-Type as "text/plain; charset=utf-8".
func (ctx *Context) String(code int, s string) error {
	return ctx.Emit(code, headerContentTypeText, s)
}

// XML sends XML response with status code, it also sets
// Content-Type as "application/xml".
func (ctx *Context) XML(code int, data interface{}) error {
	bs, err := xml.Marshal(data)
	if err != nil {
		return err
	}
	return ctx.Blob(code, headerContentTypeXML, bs)
}

// XMLBlob sends blob XML response with status code, it also sets
// Content-Type as "application/xml".
func (ctx *Context) XMLBlob(code int, bs []byte) error {
	return ctx.Blob(code, headerContentTypeXML, bs)
}

// HTML sends HTML response with status code, it also sets
// Content-Type as "text/html".
func (ctx *Context) HTML(code int, html string) error {
	return ctx.Emit(code, headerContentTypeHTML, html)
}

// HTMLBlob sends blob HTML response with status code, it also sets
// Content-Type as "text/html".
func (ctx *Context) HTMLBlob(code int, bs []byte) error {
	return ctx.Blob(code, headerContentTypeHTML, bs)
}

// Render renders a template with data, and sends HTML response with status code.
func (ctx *Context) Render(code int, name string, data interface{}) (err error) {
	if ctx.router.Renderer == nil {
		return ErrRendererNotRegister
	}

	buf := getBuffer()
	defer func() {
		putBuffer(buf)
	}()
	if err = ctx.router.Renderer.Render(buf, name, data, ctx); err != nil {
		return err
	}
	return ctx.Blob(code, headerContentTypeHTML, buf.Bytes())
}

// Emit sends a response with the given status code, content type and string body.
func (ctx *Context) Emit(code int, contentType string, body string) (err error) {
	ctx.SetContentType(contentType)
	ctx.Response.WriteHeader(code)
	_, err = io.WriteString(ctx.Response, body)
	return
}

// Blob sends a response with the given status code, content type and blob data.
func (ctx *Context) Blob(code int, contentType string, bs []byte) (err error) {
	ctx.SetContentType(contentType)
	ctx.Response.WriteHeader(code)
	_, err = ctx.Response.Write(bs)
	return
}

// FormValue is a shortcut of http.Request.FormValue.
func (ctx *Context) FormValue(key string) string {
	return ctx.Request.FormValue(key)
}

// PostFormValue is a shortcut of http.Request.PostFormValue.
func (ctx *Context) PostFormValue(key string) string {
	return ctx.Request.PostFormValue(key)
}

// QueryString returns the raw query of request URL.
func (ctx *Context) QueryString() string {
	return ctx.Request.URL.RawQuery
}

// QueryParam returns the param for the given key.
func (ctx *Context) QueryParam(key string) string {
	return ctx.QueryParams().Get(key)
}

// QueryParams returns request URL values.
func (ctx *Context) QueryParams() url.Values {
	if ctx.query == nil {
		ctx.query = ctx.Request.URL.Query()
	}
	return ctx.query
}

// DefaultQuery returns the param for the given key, returns the default value
// if the param is not present.
func (ctx *Context) DefaultQuery(key, defaultVlue string) string {
	if vs, ok := ctx.QueryParams()[key]; ok && len(vs) != 0 {
		return vs[0]
	}

	return defaultVlue
}

// RouteURL returns the URL of the naming route.
func (ctx *Context) RouteURL(name string, args ...string) (*url.URL, error) {
	return ctx.router.URL(name, args...)
}

// BasicAuth is a shortcut of http.Request.BasicAuth.
func (ctx *Context) BasicAuth() (username, password string, ok bool) {
	return ctx.Request.BasicAuth()
}

// SendFile sends a file to browser.
func (ctx *Context) SendFile(filename string, r io.Reader) (err error) {
	ctx.Response.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	_, err = io.Copy(ctx.Response, r)
	return
}
