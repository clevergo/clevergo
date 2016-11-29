// Copyright 2016 The Gem Authors. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package middleware

import (
	"encoding/base64"
	"errors"

	"github.com/go-gem/gem"
	"github.com/valyala/fasthttp"
)

// BasicAuth default configuration
var (
	BasicAuthOnValid   = func(ctx *gem.Context, _ string) {}
	BasicAuthOnInvalid = func(ctx *gem.Context, _ error) {
		// Sends 401 Unauthorized response.
		ctx.Response.SetStatusCode(fasthttp.StatusUnauthorized)
	}
)

// BasicAuth Basic Auth middleware
type BasicAuth struct {
	// Skipper defines a function to skip middleware.
	Skipper Skipper

	// Validator is a function to validate BasicAuth credentials.
	// Required.
	Validator func(username, password string) bool

	// OnValid will be invoked on when the Validator return true.
	//
	// It is easy to share the username with the other middlewares
	// by using ctx.SetUserValue.
	//
	// Optional.
	OnValid func(ctx *gem.Context, username string)

	// OnInvalid will be invoked on when error was occurred, such
	// as empty authorization, invalid authorization, incorrect
	// username or password.
	//
	// If you are care about the error message, you can
	// print it by using ctx.Logger().
	//
	// Optional.
	OnInvalid func(ctx *gem.Context, err error)
}

// NewBasicAuth returns BasicAuth instance by the
// given validator function.
func NewBasicAuth(validator func(username, password string) bool) *BasicAuth {
	return &BasicAuth{
		Skipper:   defaultSkipper,
		Validator: validator,
		OnValid:   BasicAuthOnValid,
		OnInvalid: BasicAuthOnInvalid,
	}
}

// BasicAuth errors
var (
	BasicAuthErrEmptyAuthorization = errors.New("empty authorization")
)

// Handle implements Middleware's Handle function.
func (m *BasicAuth) Handle(next gem.Handler) gem.Handler {
	if m.Skipper == nil {
		m.Skipper = defaultSkipper
	}
	basicLen := len(gem.HeaderBasic)

	return gem.HandlerFunc(func(ctx *gem.Context) {
		if m.Skipper(ctx) {
			next.Handle(ctx)
			return
		}

		auth := gem.Bytes2String(ctx.RequestCtx.Request.Header.Peek(gem.HeaderAuthorization))

		if auth == "" {
			m.OnInvalid(ctx, BasicAuthErrEmptyAuthorization)
			return
		}

		if len(auth) > basicLen+1 && auth[:basicLen] == gem.HeaderBasic {
			b, err := base64.StdEncoding.DecodeString(auth[basicLen+1:])
			if err != nil {
				m.OnInvalid(ctx, err)
				return
			}
			cred := string(b)
			for i := 0; i < len(cred); i++ {
				if cred[i] == ':' {
					// Verify credentials
					username := cred[:i]
					psw := cred[i+1:]
					if m.Validator(username, psw) {
						m.OnValid(ctx, username)
						next.Handle(ctx)
						return
					}
				}
			}
		}

		m.OnInvalid(ctx, nil)
	})
}
