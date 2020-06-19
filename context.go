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

	"clevergo.tech/log"
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

func getContext(app *Application, w http.ResponseWriter, r *http.Request) *Context {
	c := contextPool.Get().(*Context)
	c.reset()
	c.app = app
	c.Response = w
	c.Request = r
	if cap(c.Params) < int(app.maxParams) {
		c.Params = make(Params, 0, app.maxParams)
	}
	return c
}

func putContext(c *Context) {
	contextPool.Put(c)
}

// Context contains incoming request, route, params and manages outgoing response.
type Context struct {
	app      *Application
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

func (c *Context) reset() {
	c.Params = c.Params[0:0]
	c.Route = nil
	c.query = nil
}

// Error is a shortcut of http.Error.
func (c *Context) Error(code int, msg string) error {
	http.Error(c.Response, msg, code)
	return nil
}

// NotFound is a shortcut of http.NotFound.
func (c *Context) NotFound() error {
	http.NotFound(c.Response, c.Request)
	return nil
}

// Redirect is a shortcut of http.Redirect.
func (c *Context) Redirect(code int, url string) error {
	http.Redirect(c.Response, c.Request, url, code)
	return nil
}

// ServeFile is a shortcut of http.ServeFile.
func (c *Context) ServeFile(name string) error {
	http.ServeFile(c.Response, c.Request, name)
	return nil
}

// ServeContent is a shortcut of http.ServeContent.
func (c *Context) ServeContent(name string, modtime time.Time, content io.ReadSeeker) error {
	http.ServeContent(c.Response, c.Request, name, modtime, content)
	return nil
}

// SetContentType sets the content type header.
func (c *Context) SetContentType(v string) {
	c.SetHeader(headerContentType, v)
}

// SetContentTypeHTML sets the content type as HTML.
func (c *Context) SetContentTypeHTML() {
	c.SetContentType(headerContentTypeHTML)
}

// SetContentTypeText sets the content type as text.
func (c *Context) SetContentTypeText() {
	c.SetContentType(headerContentTypeText)
}

// SetContentTypeJSON sets the content type as JSON.
func (c *Context) SetContentTypeJSON() {
	c.SetContentType(headerContentTypeJSON)
}

// SetContentTypeXML sets the content type as XML.
func (c *Context) SetContentTypeXML() {
	c.SetContentType(headerContentTypeXML)
}

// Cookie is a shortcut of http.Request.Cookie.
func (c *Context) Cookie(name string) (*http.Cookie, error) {
	return c.Request.Cookie(name)
}

// Cookies is a shortcut of http.Request.Cookies.
func (c *Context) Cookies() []*http.Cookie {
	return c.Request.Cookies()
}

// SetCookie is a shortcut of http.SetCookie.
func (c *Context) SetCookie(cookie *http.Cookie) {
	http.SetCookie(c.Response, cookie)
}

// Write is a shortcut of http.ResponseWriter.Write.
func (c *Context) Write(data []byte) (int, error) {
	return c.Response.Write(data)
}

// WriteString writes the string data to response.
func (c *Context) WriteString(data string) (int, error) {
	return io.WriteString(c.Response, data)
}

// WriteHeader is a shortcut of http.ResponseWriter.WriteHeader.
func (c *Context) WriteHeader(code int) {
	c.Response.WriteHeader(code)
}

// WithValue stores the given value under the given key.
func (c *Context) WithValue(key, val interface{}) {
	c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), key, val))
}

// Value returns the value of the given key.
func (c *Context) Value(key interface{}) interface{} {
	return c.Request.Context().Value(key)
}

// IsMethod returns a boolean value indicates whether the request method is the given method.
func (c *Context) IsMethod(method string) bool {
	return c.Request.Method == method
}

// IsAJAX indicates whether it is an AJAX (XMLHttpRequest) request.
func (c *Context) IsAJAX() bool {
	return c.Request.Header.Get("X-Requested-With") == "XMLHttpRequest"
}

// IsDelete returns a boolean value indicates whether the request method is DELETE.
func (c *Context) IsDelete() bool {
	return c.IsMethod(http.MethodDelete)
}

// IsGet returns a boolean value indicates whether the request method is GET.
func (c *Context) IsGet() bool {
	return c.IsMethod(http.MethodGet)
}

// IsOptions returns a boolean value indicates whether the request method is OPTIONS.
func (c *Context) IsOptions() bool {
	return c.IsMethod(http.MethodOptions)
}

// IsPatch returns a boolean value indicates whether the request method is PATCH.
func (c *Context) IsPatch() bool {
	return c.IsMethod(http.MethodPatch)
}

// IsPost returns a boolean value indicates whether the request method is POST.
func (c *Context) IsPost() bool {
	return c.IsMethod(http.MethodPost)
}

// IsPut returns a boolean value indicates whether the request method is PUT.
func (c *Context) IsPut() bool {
	return c.IsMethod(http.MethodPut)
}

// GetHeader is a shortcut of http.Request.Header.Get.
func (c *Context) GetHeader(name string) string {
	return c.Request.Header.Get(name)
}

// JSON sends JSON response with status code, it also sets
// Content-Type as "application/json".
func (c *Context) JSON(code int, data interface{}) error {
	bs, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return c.Blob(code, headerContentTypeJSON, bs)
}

// JSONBlob sends blob JSON response with status code, it also sets
// Content-Type as "application/json".
func (c *Context) JSONBlob(code int, bs []byte) error {
	return c.Blob(code, headerContentTypeJSON, bs)
}

// JSONP is a shortcut of JSONPCallback with specified callback param name.
func (c *Context) JSONP(code int, data interface{}) error {
	return c.JSONPCallback(code, "callback", data)
}

// JSONPCallback sends JSONP response with status code, it also sets
// Content-Type as "application/javascript".
// If the callback is not present, returns JSON response instead.
func (c *Context) JSONPCallback(code int, callback string, data interface{}) error {
	fn := c.QueryParam(callback)
	if fn == "" {
		return c.JSON(code, data)
	}

	bs, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return c.Emit(code, headerContentTypeJavaScript, formatJSONP(fn, bs))
}

// JSONPBlob is a shortcut of JSONPCallbackBlob with specified callback param name.
func (c *Context) JSONPBlob(code int, bs []byte) error {
	return c.JSONPCallbackBlob(code, "callback", bs)
}

// JSONPCallbackBlob sends blob JSONP response with status code, it also sets
// Content-Type as "application/javascript".
// If the callback is not present, returns JSON response instead.
func (c *Context) JSONPCallbackBlob(code int, callback string, bs []byte) (err error) {
	fn := c.QueryParam(callback)
	if fn == "" {
		return c.JSONBlob(code, bs)
	}

	return c.Emit(code, headerContentTypeJavaScript, formatJSONP(fn, bs))
}

func formatJSONP(callback string, bs []byte) string {
	return callback + "(" + string(bs) + ")"
}

// String send string response with status code, it also sets
// Content-Type as "text/plain; charset=utf-8".
func (c *Context) String(code int, s string) error {
	return c.Emit(code, headerContentTypeText, s)
}

// Stringf formats according to a format specifier and sends the resulting string
// with the status code, it also sets Content-Type as "text/plain; charset=utf-8".
func (c *Context) Stringf(code int, format string, a ...interface{}) error {
	return c.String(code, fmt.Sprintf(format, a...))
}

// XML sends XML response with status code, it also sets
// Content-Type as "application/xml".
func (c *Context) XML(code int, data interface{}) error {
	bs, err := xml.Marshal(data)
	if err != nil {
		return err
	}
	return c.Blob(code, headerContentTypeXML, bs)
}

// XMLBlob sends blob XML response with status code, it also sets
// Content-Type as "application/xml".
func (c *Context) XMLBlob(code int, bs []byte) error {
	return c.Blob(code, headerContentTypeXML, bs)
}

// HTML sends HTML response with status code, it also sets
// Content-Type as "text/html".
func (c *Context) HTML(code int, html string) error {
	return c.Emit(code, headerContentTypeHTML, html)
}

// HTMLBlob sends blob HTML response with status code, it also sets
// Content-Type as "text/html".
func (c *Context) HTMLBlob(code int, bs []byte) error {
	return c.Blob(code, headerContentTypeHTML, bs)
}

// Render renders a template with data, and sends HTML response with status code.
func (c *Context) Render(code int, name string, data interface{}) (err error) {
	if c.app.Renderer == nil {
		return ErrRendererNotRegister
	}

	buf := getBuffer()
	defer func() {
		putBuffer(buf)
	}()
	if err = c.app.Renderer.Render(buf, name, data, c); err != nil {
		return err
	}
	return c.Blob(code, headerContentTypeHTML, buf.Bytes())
}

// Emit sends a response with the given status code, content type and string body.
func (c *Context) Emit(code int, contentType string, body string) (err error) {
	c.SetContentType(contentType)
	c.Response.WriteHeader(code)
	_, err = io.WriteString(c.Response, body)
	return
}

// Blob sends a response with the given status code, content type and blob data.
func (c *Context) Blob(code int, contentType string, bs []byte) (err error) {
	c.SetContentType(contentType)
	c.Response.WriteHeader(code)
	_, err = c.Response.Write(bs)
	return
}

// Context returns the context of request.
func (c *Context) Context() context.Context {
	return c.Request.Context()
}

// FormValue is a shortcut of http.Request.FormValue.
func (c *Context) FormValue(key string) string {
	return c.Request.FormValue(key)
}

// PostFormValue is a shortcut of http.Request.PostFormValue.
func (c *Context) PostFormValue(key string) string {
	return c.Request.PostFormValue(key)
}

// Host returns http.Request.Host.
func (c *Context) Host() string {
	return c.Request.Host
}

// QueryString returns the raw query of request URL.
func (c *Context) QueryString() string {
	return c.Request.URL.RawQuery
}

// QueryParam returns the param for the given key.
func (c *Context) QueryParam(key string) string {
	return c.QueryParams().Get(key)
}

// QueryParams returns request URL values.
func (c *Context) QueryParams() url.Values {
	if c.query == nil {
		c.query = c.Request.URL.Query()
	}
	return c.query
}

// DefaultQuery returns the param for the given key, returns the default value
// if the param is not present.
func (c *Context) DefaultQuery(key, defaultVlue string) string {
	if vs, ok := c.QueryParams()[key]; ok && len(vs) != 0 {
		return vs[0]
	}

	return defaultVlue
}

// RouteURL returns the URL of the naming route.
func (c *Context) RouteURL(name string, args ...string) (*url.URL, error) {
	return c.app.RouteURL(name, args...)
}

// BasicAuth is a shortcut of http.Request.BasicAuth.
func (c *Context) BasicAuth() (username, password string, ok bool) {
	return c.Request.BasicAuth()
}

// SendFile sends a file to browser.
func (c *Context) SendFile(filename string, r io.Reader) (err error) {
	c.Response.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	_, err = io.Copy(c.Response, r)
	return
}

// Decode decodes request's input, stores it in the value pointed to by v.
func (c *Context) Decode(v interface{}) (err error) {
	if c.app.Decoder == nil {
		return ErrDecoderNotRegister
	}
	return c.app.Decoder.Decode(c.Request, v)
}

// SetHeader is a shortcut of http.ResponseWriter.Header().Set.
func (c *Context) SetHeader(key, value string) {
	c.Response.Header().Set(key, value)
}

// Logger returns the logger of application.
func (c *Context) Logger() log.Logger {
	return c.app.Logger
}
