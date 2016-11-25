// Copyright 2016 The Gem Authors. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package middleware

import (
	"strconv"
	"strings"

	"github.com/go-gem/gem"
)

// CORS default configuration.
var (
	CORSAllowOrigins = []string{"*"}
	CORSAllowMethods = []string{
		gem.MethodGet, gem.MethodHead, gem.MethodPut,
		gem.MethodPatch, gem.MethodPost, gem.MethodDelete,
	}
)

// Cross-Origin Resource Sharing middleware.
type CORS struct {
	// Skipper defines a function to skip middleware.
	Skipper Skipper

	// Access-Control-Allow-Origin
	AllowOrigins []string

	// Access-Control-Allow-Methods
	AllowMethods []string

	// Access-Control-Allow-Headers
	AllowHeaders []string

	// Access-Control-Expose-Headers
	ExposeHeaders []string

	// Access-Control-Max-Age
	MaxAge int

	// Access-Control-Allow-Credentials
	AllowCredentials bool
}

// NewCORS returns CORS instance with the default configuration.
func NewCORS() *CORS {
	return &CORS{
		Skipper:       defaultSkipper,
		AllowOrigins:  CORSAllowOrigins,
		AllowMethods:  CORSAllowMethods,
		AllowHeaders:  []string{},
		ExposeHeaders: []string{},
	}
}

// Handle implements Middleware's Handle function.
func (m *CORS) Handle(next gem.Handler) gem.Handler {
	if m.Skipper == nil {
		m.Skipper = defaultSkipper
	}
	if len(m.AllowOrigins) == 0 {
		m.AllowOrigins = CORSAllowOrigins
	}
	if len(m.AllowMethods) == 0 {
		m.AllowMethods = CORSAllowMethods
	}

	allowMethods := strings.Join(m.AllowMethods, ",")
	allowHeaders := strings.Join(m.AllowHeaders, ",")
	exposeHeaders := strings.Join(m.ExposeHeaders, ",")
	maxAge := strconv.Itoa(m.MaxAge)

	return gem.HandlerFunc(func(ctx *gem.Context) {
		if m.Skipper(ctx) {
			next.Handle(ctx)
			return
		}

		next.Handle(ctx)

		origin := ctx.ReqHeader(gem.HeaderOrigin)

		allowedOrigin := ""
		for _, o := range m.AllowOrigins {
			if o == "*" || o == origin {
				allowedOrigin = o
				break
			}
		}

		// Simple request
		if !IsCORSSimpleRequest(ctx) {
			ctx.RequestCtx.Response.Header.Add(gem.HeaderVary, gem.HeaderOrigin)
			if origin == "" || allowedOrigin == "" {
				next.Handle(ctx)
				return
			}
			ctx.RequestCtx.Response.Header.Set(gem.HeaderAccessControlAllowOrigin, allowedOrigin)
			if m.AllowCredentials {
				ctx.RequestCtx.Response.Header.Set(gem.HeaderAccessControlAllowCredentials, "true")
			}
			if exposeHeaders != "" {
				ctx.RequestCtx.Response.Header.Set(gem.HeaderAccessControlExposeHeaders, exposeHeaders)
			}
			next.Handle(ctx)
			return
		}

		// Preflight request
		ctx.RequestCtx.Response.Header.Add(gem.HeaderVary, gem.HeaderOrigin)
		ctx.RequestCtx.Response.Header.Add(gem.HeaderVary, gem.HeaderAccessControlRequestMethod)
		ctx.RequestCtx.Response.Header.Add(gem.HeaderVary, gem.HeaderAccessControlRequestHeaders)
		if origin == "" || allowedOrigin == "" {
			next.Handle(ctx)
			return
		}
		ctx.RequestCtx.Response.Header.Set(gem.HeaderAccessControlAllowOrigin, allowedOrigin)
		ctx.RequestCtx.Response.Header.Set(gem.HeaderAccessControlAllowMethods, allowMethods)
		if m.AllowCredentials {
			ctx.RequestCtx.Response.Header.Set(gem.HeaderAccessControlAllowCredentials, "true")
		}
		if allowHeaders != "" {
			ctx.RequestCtx.Response.Header.Set(gem.HeaderAccessControlAllowHeaders, allowHeaders)
		} else {
			h := ctx.RequestCtx.Response.Header.Peek(gem.HeaderAccessControlRequestHeaders)
			if len(h) > 0 {
				ctx.RequestCtx.Response.Header.Set(gem.HeaderAccessControlAllowHeaders, string(h))
			}
		}
		if m.MaxAge > 0 {
			ctx.RequestCtx.Response.Header.Set(gem.HeaderAccessControlMaxAge, maxAge)
		}

		ctx.Response.ResetBody()
	})
}

const (
	contentTypeText          = "text/plain"
	contentTypeMultipartForm = "multipart/form-data"
	contentTypeForm          = "application/x-www-form-urlencoded"
)

// IsCORSSimpleRequest returns bool to indicate whether
// the current request is a simple request.
func IsCORSSimpleRequest(ctx *gem.Context) bool {
	// Simple request method should be HEAD, GET or POST.
	if !ctx.IsHead() && !ctx.IsGet() && !ctx.IsPost() {
		return false
	}

	// Content-Type should be one of the following:
	// 	application/x-www-form-urlencoded
	// 	multipart/form-data
	// 	text/plain
	ct := ctx.ReqHeader(gem.HeaderContentType)
	if !strings.Contains(ct, contentTypeText) &&
		!strings.Contains(ct, contentTypeMultipartForm) &&
		!strings.Contains(ct, contentTypeForm) {
		return false
	}

	return true
}
