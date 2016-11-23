// Copyright 2016 The Gem Authors. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package gem

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"sync"

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
func (ctx *Context) Logger() Logger {
	return ctx.server.logger
}

// SessionsStore returns server's sessions store.
func (ctx *Context) SessionsStore() sessions.Store {
	return ctx.server.sessionsStore
}

// newContext returns a Context instance.
//
// It will try to get Context instance from contextPool,
// returns a new Context instance when failure.
func acquireContext(server *Server, reqCtx *fasthttp.RequestCtx) *Context {
	ctx := contextPool.Get().(*Context)
	ctx.RequestCtx = reqCtx
	ctx.server = server
	return ctx
}

// releaseContext release context on request was finished,
// context will be put into context pool for reusing.
func releaseContext(ctx *Context) {
	ctx.server = nil
	ctx.RequestCtx = nil
	contextPool.Put(ctx)
}

// MethodString returns HTTP request method.
func (ctx *Context) MethodString() string {
	return bytes2String(ctx.RequestCtx.Request.Header.Method())
}

// HostString returns requested host.
func (ctx *Context) HostString() string {
	return bytes2String(ctx.RequestCtx.URI().Host())
}

// PathString returns URI path.
func (ctx *Context) PathString() string {
	return bytes2String(ctx.RequestCtx.Request.URI().Path())
}

// IsAjax returns whether this is an AJAX (XMLHttpRequest) request.
func (ctx *Context) IsAjax() bool {
	return bytes.Equal(ctx.RequestCtx.Request.Header.Peek("X-Requested-With"), bytesXMLHttpRequest)
}

// HTML responses HTML data and custom status code to client.
func (ctx *Context) HTML(code int, body string) {
	ctx.SetStatusCode(code)
	ctx.Response.Header.SetBytesKV(contentType, ContentTypeHTML)
	ctx.Response.SetBodyString(body)
}

// JSON responses JSON data and custom status code to client.
func (ctx *Context) JSON(code int, v interface{}) {
	ctx.Response.SetStatusCode(code)
	data, err := json.Marshal(v)
	if err != nil {
		ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
		return
	}
	ctx.Response.Header.SetBytesKV(contentType, ContentTypeJSON)
	ctx.Response.SetBody(data)
}

// JSONP responses JSONP data and custom status code to client.
func (ctx *Context) JSONP(code int, v interface{}, callback []byte) {
	ctx.Response.SetStatusCode(code)
	data, err := json.Marshal(v)
	if err != nil {
		ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
		return
	}
	ctx.Response.Header.SetBytesKV(contentType, ContentTypeJSONP)
	callback = append(callback, "("...)
	callback = append(callback, data...)
	callback = append(callback, ")"...)
	ctx.Response.SetBody(callback)
}

// XML responses XML data and custom status code to client.
func (ctx *Context) XML(code int, v interface{}, headers ...string) {
	ctx.Response.SetStatusCode(code)
	xmlBytes, err := xml.MarshalIndent(v, "", `   `)
	if err != nil {
		ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
		return
	}

	header := xml.Header
	if len(headers) > 0 {
		header = headers[0]
	}

	var bytes []byte
	bytes = append(bytes, header...)
	bytes = append(bytes, xmlBytes...)

	ctx.Response.Header.SetBytesKV(contentType, ContentTypeXML)
	ctx.Response.SetBody(bytes)
}
