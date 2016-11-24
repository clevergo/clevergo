// Copyright 2016 The Gem Authors. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package middleware

import (
	"encoding/base64"

	"github.com/go-gem/gem"
	"github.com/valyala/fasthttp"
)

// BasicAuth Basic Auth middleware.
type BasicAuth struct {
	Skipper   Skipper
	Validator func(username, password string) bool
}

// NewBasicAuth returns BasicAuth instance by the
// given validator function
func NewBasicAuth(validator func(username, password string) bool) *BasicAuth {
	return &BasicAuth{
		Validator: validator,
	}
}

const (
	basic = "Basic"
)

// Handle implements Middleware's Handle function.
func (m *BasicAuth) Handle(next gem.Handler) gem.Handler {
	if m.Skipper == nil {
		m.Skipper = defaultSkipper
	}

	return gem.HandlerFunc(func(ctx *gem.Context) {
		if m.Skipper(ctx) {
			next.Handle(ctx)
			return
		}

		auth := string(ctx.RequestCtx.Request.Header.PeekBytes(gem.HeaderAuthorization))
		l := len(basic)

		if len(auth) > l+1 && auth[:l] == basic {
			b, err := base64.StdEncoding.DecodeString(auth[l+1:])
			if err != nil {
				ctx.Logger().Errorf("Basic Auth error: %s\n", err)
				return
			}
			cred := string(b)
			for i := 0; i < len(cred); i++ {
				if cred[i] == ':' {
					// Verify credentials
					if m.Validator(cred[:i], cred[i+1:]) {
						next.Handle(ctx)
						return
					}
				}
			}
		}

		// Return 401 message.
		ctx.Response.SetStatusCode(fasthttp.StatusUnauthorized)
		ctx.Response.SetBodyString(fasthttp.StatusMessage(fasthttp.StatusUnauthorized))
	})
}
