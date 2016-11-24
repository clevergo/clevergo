// Copyright 2016 The Gem Authors. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package middleware

import (
	"errors"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-gem/gem"
	"github.com/valyala/fasthttp"
)

const bearer = "Bearer"

// JWT default configuration.
var (
	JWTAcquireToken = func(ctx *gem.Context) (token string, err error) {
		if token, err = AcquireJWTTokenFromHeader(ctx, gem.StrHeaderAuthorization); err != nil {
			token, err = AcquireJWTTokenFromForm(ctx, "jwt_token")
		}
		return
	}

	JWTOnValid = func(ctx *gem.Context, token *jwt.Token, claims jwt.Claims) {
		ctx.SetUserValue("jwt_token", token)
		ctx.SetUserValue("jwt_claims", claims)
	}

	JWTOnInvalid = func(ctx *gem.Context, err error) {
		ctx.Logger().Errorf("JWT middleware error: %s\n", err)
	}
)

// JWT JSON WEB TOKEN middleware.
type JWT struct {
	// Skipper defines a function to skip middleware.
	Skipper Skipper

	// AcquireToken provides API to customize a function that
	// get jwt token.
	AcquireToken func(*gem.Context) (string, error)

	// OnValid will be invoked if the token is valid.
	// It is useful to store the token by ctx.SetUserValue()
	OnValid func(*gem.Context, *jwt.Token, jwt.Claims)

	// OnInvalid will be invoked if the token is invalid.
	OnInvalid func(*gem.Context, error)

	// See jwt.SigningMethod
	SigningMethod jwt.SigningMethod

	// See jwt.Keyfunc
	KeyFunc jwt.Keyfunc

	// NewClaims returns a jwt.Claims instance,
	// And then use jwt.ParseWithClaims to parse token and claims.
	// If it is not set, use jwt.Parse instead.
	NewClaims func() jwt.Claims
}

// NewJWT returns a JWT instance with the given
// params and default configuration.
func NewJWT(signingMethod jwt.SigningMethod, keyFunc jwt.Keyfunc) *JWT {
	return &JWT{
		Skipper:       defaultSkipper,
		AcquireToken:  JWTAcquireToken,
		OnValid:       JWTOnValid,
		OnInvalid:     JWTOnInvalid,
		SigningMethod: signingMethod,
		KeyFunc:       keyFunc,
	}
}

// Handle implements Middleware's Handle function.
func (m *JWT) Handle(next gem.Handler) gem.Handler {
	if m.Skipper == nil {
		m.Skipper = defaultSkipper
	}

	return gem.HandlerFunc(func(ctx *gem.Context) {
		tokenStr, err := m.AcquireToken(ctx)
		// Returns Bad Request status code if the token is empty.
		if err != nil || tokenStr == "" {
			ctx.SetStatusCode(fasthttp.StatusBadRequest)
			ctx.SetBodyString(fasthttp.StatusMessage(fasthttp.StatusBadRequest))
			return
		}

		var token *jwt.Token
		var claims jwt.Claims
		if m.NewClaims == nil {
			token, err = jwt.Parse(tokenStr, m.KeyFunc)
		} else {
			claims = m.NewClaims()
			token, err = jwt.ParseWithClaims(tokenStr, claims, m.KeyFunc)

			if err = claims.Valid(); err != nil {
				m.OnInvalid(ctx, err)
				return
			}
		}

		if err != nil {
			m.OnInvalid(ctx, err)
			return
		}

		m.OnValid(ctx, token, claims)

		next.Handle(ctx)
	})
}

// JWT Error
var (
	ErrEmptyJWTInHeader = errors.New("empty jwt in request header")
	ErrEmptyJWTInForm   = errors.New("empty jwt in query string or post form")
)

// AcquireJWTTokenFromHeader acquire jwt token from the request
// header.
func AcquireJWTTokenFromHeader(ctx *gem.Context, key string) (string, error) {
	auth := string(ctx.RequestCtx.Request.Header.Peek(key))
	l := len(bearer)
	if len(auth) > l+1 && auth[:l] == bearer {
		return auth[l+1:], nil
	}

	return "", ErrEmptyJWTInHeader
}

// AcquireJWTTokenFromForm acquire jwt token from the query string
// or post form.
func AcquireJWTTokenFromForm(ctx *gem.Context, key string) (string, error) {
	token := ctx.RequestCtx.FormValue(key)
	if len(token) == 0 {
		return "", ErrEmptyJWTInForm
	}
	return string(token), nil
}
