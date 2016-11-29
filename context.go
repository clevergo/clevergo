// Copyright 2016 The Gem Authors. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package gem

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"strconv"
	"sync"

	"github.com/go-gem/sessions"
	"github.com/valyala/fasthttp"
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

// acquireContext returns a Context instance.
func acquireContext(srv *Server, reqCtx *fasthttp.RequestCtx) *Context {
	ctx := contextPool.Get().(*Context)
	ctx.RequestCtx = reqCtx
	ctx.server = srv
	return ctx
}

// releaseContext release context on request was finished,
// context will be put into context pool for reusing.
func releaseContext(ctx *Context) {
	ctx.server = nil
	ctx.RequestCtx = nil
	contextPool.Put(ctx)
}

// Param returns a string value from router params
// by the given key.
func (ctx *Context) Param(key string) string {
	v, ok := ctx.UserValue(key).(string)
	if ok {
		return v
	}

	return ""
}

// ParamInt returns an integer value from router params
// by the given key.
func (ctx *Context) ParamInt(key string) int {
	if v, err := strconv.Atoi(ctx.Param(key)); err == nil {
		return v
	}

	return 0
}

// IsAjax returns bool to indicate whether the current request
// is an AJAX (XMLHttpRequest) request.
func (ctx *Context) IsAjax() bool {
	return bytes.Equal(ctx.RequestCtx.Request.Header.Peek(HeaderXRequestedWith), HeaderXMLHttpRequestBytes)
}

// HTML responses HTML data and custom status code to client.
func (ctx *Context) HTML(code int, body string) {
	ctx.RequestCtx.Response.Header.SetStatusCode(code)
	ctx.RequestCtx.Response.Header.SetContentType(HeaderContentTypeHTML)
	ctx.Response.SetBodyString(body)
}

// JSON responses JSON data and custom status code to client.
func (ctx *Context) JSON(code int, v interface{}) {
	ctx.RequestCtx.Response.Header.SetStatusCode(code)
	data, err := json.Marshal(v)
	if err != nil {
		ctx.Logger().Errorf("JSON error: %s\n", err)
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}
	ctx.RequestCtx.Response.Header.SetContentType(HeaderContentTypeJSON)
	ctx.RequestCtx.Response.SetBody(data)
}

// JSONP responses JSONP data and custom status code to client.
func (ctx *Context) JSONP(code int, v interface{}, callback []byte) {
	ctx.RequestCtx.Response.Header.SetStatusCode(code)
	data, err := json.Marshal(v)
	if err != nil {
		ctx.Logger().Errorf("JSON error: %s\n", err)
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
	}
	ctx.RequestCtx.Response.Header.SetContentType(HeaderContentTypeJSONP)
	callback = append(callback, "("...)
	callback = append(callback, data...)
	callback = append(callback, ")"...)
	ctx.RequestCtx.Response.SetBody(callback)
}

// XML responses XML data and custom status code to client.
func (ctx *Context) XML(code int, v interface{}, headers ...string) {
	ctx.RequestCtx.Response.Header.SetStatusCode(code)
	xmlBytes, err := xml.Marshal(v)
	if err != nil {
		ctx.Logger().Errorf("XML error: %s\n", err)
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}

	header := xml.Header
	if len(headers) > 0 {
		header = headers[0]
	}

	var bytes []byte
	bytes = append(bytes, header...)
	bytes = append(bytes, xmlBytes...)

	ctx.RequestCtx.Response.Header.SetContentType(HeaderContentTypeXML)
	ctx.RequestCtx.Response.SetBody(bytes)
}
