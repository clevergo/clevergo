// Copyright 2016 The Gem Authors. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package middleware

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/go-gem/gem"
	"github.com/valyala/fasthttp"
)

// JWT default configuration.
const (
	JWTFormKey    = "_jwt"
	JWTContextKey = "jwt_token"
	JWTClaimsKey  = "jwt_claims"
)

// JWT JSON WEB TOKEN middleware.
type JWT struct {
	// Skipper defines a function to skip middleware.
	Skipper Skipper

	// See jwt.SigningMethod
	SigningMethod jwt.SigningMethod

	// See jwt.Keyfunc
	KeyFunc jwt.Keyfunc

	// FormKey be used to acquire token from query string
	// or post form.
	FormKey string

	// ContextKey be used to ctx.SetUserValue(ContextKey,jwt.Token)
	ContextKey string

	// NewClaims returns a jwt.Claims instance,
	// And then use jwt.ParseWithClaims to parse token and claims.
	// If it is not set, use jwt.Parse instead.
	NewClaims func() jwt.Claims

	// ClaimsKey be used to ctx.SetUserValue(ClaimsKey, jwt.Claims)
	ClaimsKey string
}

// NewJWT returns a JWT instance with the given
// params and default configuration.
func NewJWT(signingMethod jwt.SigningMethod, keyFunc jwt.Keyfunc) *JWT {
	return &JWT{
		Skipper:       defaultSkipper,
		SigningMethod: signingMethod,
		KeyFunc:       keyFunc,
		FormKey:       JWTFormKey,
		ContextKey:    JWTContextKey,
		ClaimsKey:     JWTClaimsKey,
	}
}

// Handle implements Middleware's Handle function.
func (m *JWT) Handle(next gem.Handler) gem.Handler {
	if m.Skipper == nil {
		m.Skipper = defaultSkipper
	}

	return gem.HandlerFunc(func(ctx *gem.Context) {
		if m.Skipper(ctx) {
			next.Handle(ctx)
			return
		}

		var tokenStr string
		if tokenStr = AcquireJWTTokenFromHeader(ctx, gem.HeaderAuthorization); tokenStr == "" {
			tokenStr = AcquireJWTTokenFromForm(ctx, m.FormKey)
		}
		// Returns Bad Request status code if the token is empty.
		if tokenStr == "" {
			ctx.SetStatusCode(fasthttp.StatusBadRequest)
			ctx.SetBodyString(fasthttp.StatusMessage(fasthttp.StatusBadRequest))
			return
		}

		var err error
		var token *jwt.Token
		var claims jwt.Claims
		if m.NewClaims == nil {
			token, err = jwt.Parse(tokenStr, m.KeyFunc)
		} else {
			claims = m.NewClaims()
			token, err = jwt.ParseWithClaims(tokenStr, claims, m.KeyFunc)
			if err == nil {
				err = claims.Valid()
			}
		}

		if err != nil {
			ctx.Logger().Infoln(err)
			ctx.SetStatusCode(fasthttp.StatusUnauthorized)
			ctx.SetBodyString(fasthttp.StatusMessage(fasthttp.StatusUnauthorized))
			return
		}

		ctx.SetUserValue(m.ContextKey, token)
		ctx.SetUserValue(m.ClaimsKey, claims)

		next.Handle(ctx)
	})
}

var (
	bearerLen = len(gem.HeaderBearer)
)

// AcquireJWTTokenFromHeader acquire jwt token from the request
// header.
func AcquireJWTTokenFromHeader(ctx *gem.Context, key string) string {
	auth := gem.Bytes2String(ctx.RequestCtx.Request.Header.Peek(key))
	if len(auth) > bearerLen+1 && auth[:bearerLen] == gem.HeaderBearer {
		return auth[bearerLen+1:]
	}

	return ""
}

// AcquireJWTTokenFromForm acquire jwt token from the query string
// or post form.
func AcquireJWTTokenFromForm(ctx *gem.Context, key string) string {
	token := ctx.RequestCtx.FormValue(key)
	if len(token) == 0 {
		return ""
	}
	return string(token)
}
