// Copyright 2016 The Gem Authors. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package gem

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"sync"

	"github.com/go-gem/log"
	"github.com/go-gem/sessions"
	"github.com/valyala/fasthttp"
)

var (
	contentType = []byte("Content-Type")

	// ContentTypeDefault default Content-Type
	ContentTypeDefault = []byte("text/plain; charset=utf-8")

	// ContentTypeHTML HTML Content-Type
	ContentTypeHTML = []byte("text/html; charset=utf-8")

	// ContentTypeJSON JSON Content-Type
	ContentTypeJSON = []byte("application/json; charset=utf-8")

	// ContentTypeJSONP JSONP Content-Type
	ContentTypeJSONP = []byte("application/javascript; charset=utf-8")

	// ContentTypeXML XML Content-Type
	ContentTypeXML = []byte("application/xml; charset=utf-8")
)

var (
	bytesXMLHttpRequest = []byte("XMLHttpRequest")
)

// contextPool Contexts's pool for reusing.
var contextPool = &sync.Pool{
	New: func() interface{} {
		return &Context{}
	},
}

// Context context contains request and response.
//
// It is forbidden copying Context instances.
type Context struct {
	*fasthttp.RequestCtx
	server *Server
}

// Logger returns server's logger.
func (c *Context) Logger() log.Logger {
	return c.server.logger
}

// SessionStore returns server's sessions store.
func (c *Context) SessionStore() sessions.Store {
	return c.server.sessionStore
}

// newContext returns a Context instance.
//
// It will try to get Context instance from contextPool,
// returns a new Context instance when failure.
func acquireContext(server *Server, ctx *fasthttp.RequestCtx) *Context {
	c := contextPool.Get().(*Context)
	c.RequestCtx = ctx
	c.server = server
	return c
}

// close close the current context on request was finished,
// context will be put into context pool for reusing.
func (c *Context) close() {
	c.server = nil
	c.RequestCtx = nil
	contextPool.Put(c)
}

// MethodString returns HTTP request method.
func (c *Context) MethodString() string {
	return bytes2String(c.RequestCtx.Request.Header.Method())
}

// HostString returns Host header value.
func (c *Context) HostString() string {
	return bytes2String(c.RequestCtx.Request.Header.Host())
}

// PathString returns URI path.
func (c *Context) PathString() string {
	return bytes2String(c.RequestCtx.Request.URI().Path())
}

// IsAjax returns whether this is an AJAX (XMLHttpRequest) request.
func (c *Context) IsAjax() bool {
	return bytes.Equal(c.RequestCtx.Request.Header.Peek("X-Requested-With"), bytesXMLHttpRequest)
}

// HTML responses HTML data and custom status code to client.
func (c *Context) HTML(code int, body string) {
	c.SetStatusCode(code)
	c.Response.Header.SetBytesKV(contentType, ContentTypeHTML)
	c.Response.SetBodyString(body)
}

// JSON responses JSON data and custom status code to client.
func (c *Context) JSON(code int, v interface{}) {
	c.Response.SetStatusCode(code)
	data, err := json.Marshal(v)
	if err != nil {
		c.Logger().Fatalf("JSON: %q\n.", err)
		return
	}
	c.Response.Header.SetBytesKV(contentType, ContentTypeJSON)
	c.Response.SetBody(data)
}

// JSONP responses JSONP data and custom status code to client.
func (c *Context) JSONP(code int, v interface{}, callback []byte) {
	c.Response.SetStatusCode(code)
	data, err := json.Marshal(v)
	if err != nil {
		c.Logger().Fatalf("JSONP: %q\n.", err)
		return
	}
	c.Response.Header.SetBytesKV(contentType, ContentTypeJSONP)
	callback = append(callback, "("...)
	callback = append(callback, data...)
	callback = append(callback, ")"...)
	c.Response.SetBody(callback)
}

// XML responses XML data and custom status code to client.
func (c *Context) XML(code int, v interface{}, headers ...string) {
	c.Response.SetStatusCode(code)
	xmlBytes, err := xml.MarshalIndent(v, "", `   `)
	if err != nil {
		c.Logger().Fatalf("XML: %q\n.", err)
		return
	}

	header := xml.Header
	if len(headers) > 0 {
		header = headers[0]
	}

	var bytes []byte
	bytes = append(bytes, header...)
	bytes = append(bytes, xmlBytes...)

	c.Response.Header.SetBytesKV(contentType, ContentTypeXML)
	c.Response.SetBody(bytes)
}
