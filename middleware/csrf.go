// Copyright 2016 The Gem Authors. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package middleware

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"strings"

	"github.com/go-gem/gem"
	"github.com/go-gem/sessions"
	"github.com/valyala/fasthttp"
)

// CSRF default configuration.
var (
	CSRFSafeMethods = []string{gem.MethodGet, gem.MethodHead, gem.MethodOptions}

	CSRFMaskLen = 8

	CSRFTokenLen = 32

	CSRFCookieKey = "_csrf"

	CSRFCookieOptions = &sessions.Options{
		MaxAge:   60 * 60, // one hour.
		HttpOnly: true,
	}

	CSRFFormKey = "_csrf"

	CSRFHeaderKey = "X-CSRF-Token"

	CSRFContextKey = "csrf_token"
)

// CSRF Cross-site request forgery protection middleware.
type CSRF struct {
	// Skipper defines a function to skip middleware.
	Skipper Skipper

	// SafeMethods
	// See https://www.w3.org/Protocols/rfc2616/rfc2616-sec9.html#sec9.1.1
	SafeMethods []string

	// MaskLen mask length.
	MaskLen int

	// TokenLen the length of true token.
	TokenLen int

	// ContextKey be used to ctx.SetUserValue(ContextKey, encodedToken)
	ContextKey string

	// CookieKey be used to acquire true token from cookie.
	CookieKey string

	//
	CookieOptions *sessions.Options

	// FormKey be used to acquire encoded token from query string
	// or post form.
	FormKey string

	// HeaderKey be used to acquire encoded token from header.
	HeaderKey string
}

// NewCSRF returns a CSRF instance with the default
// configuration.
func NewCSRF() *CSRF {
	return &CSRF{
		Skipper:       defaultSkipper,
		SafeMethods:   CSRFSafeMethods,
		MaskLen:       CSRFMaskLen,
		TokenLen:      CSRFTokenLen,
		ContextKey:    CSRFContextKey,
		CookieKey:     CSRFCookieKey,
		CookieOptions: CSRFCookieOptions,
		FormKey:       CSRFFormKey,
		HeaderKey:     CSRFHeaderKey,
	}
}

// Handle implements Middleware's Handle function.
func (m *CSRF) Handle(next gem.Handler) gem.Handler {
	if m.Skipper == nil {
		m.Skipper = defaultSkipper
	}
	if m.MaskLen <= 0 {
		m.MaskLen = CSRFMaskLen
	}
	if m.TokenLen <= 0 {
		m.TokenLen = CSRFTokenLen
	}
	if m.CookieKey == "" {
		m.CookieKey = CSRFCookieKey
	}
	if m.CookieOptions == nil {
		m.CookieOptions = CSRFCookieOptions
	}
	if m.ContextKey == "" {
		m.ContextKey = CSRFContextKey
	}
	if m.FormKey == "" {
		m.FormKey = CSRFFormKey
	}
	if m.HeaderKey == "" {
		m.HeaderKey = CSRFHeaderKey
	}

	return gem.HandlerFunc(func(ctx *gem.Context) {
		var trueToken []byte
		trueTokenStr := string(ctx.RequestCtx.Request.Header.Cookie(m.CookieKey))
		if trueTokenStr != "" {
			if token, err := base64.StdEncoding.DecodeString(trueTokenStr); err == nil && len(token) >= m.TokenLen {
				trueToken = token[:m.TokenLen]
			}
		}

		if len(trueToken) == 0 {
			trueToken = randomBytes(m.TokenLen)
			cookie := sessions.NewCookie(m.CookieKey, base64.StdEncoding.EncodeToString(trueToken), m.CookieOptions)
			ctx.Response.Header.SetCookie(cookie)
		}

		// Always generate en encoded token.
		encodedToken := generateCSRFToken(m.MaskLen, trueToken)
		ctx.SetUserValue(m.ContextKey, encodedToken)

		if m.Skipper(ctx) {
			next.Handle(ctx)
			return
		}

		method := gem.Bytes2String(ctx.RequestCtx.Request.Header.Method())
		for _, v := range m.SafeMethods {
			if v == method {
				next.Handle(ctx)
				return
			}
		}

		// acquire csrf token from header, query string or post form.
		token := ctx.RequestCtx.Request.Header.Peek(m.HeaderKey)
		if len(token) == 0 {
			token = ctx.RequestCtx.FormValue(m.FormKey)
		}

		// verify CSRF token
		if err := validateCSRF(m.MaskLen, token, trueToken); err != nil {
			ctx.SetStatusCode(fasthttp.StatusBadRequest)
			ctx.SetBodyString("Unable to verify your data submission.")
			return
		}

		next.Handle(ctx)
	})
}

var errInvalidCRSF = errors.New("The CSRF token is invalid.")

func generateCSRFToken(maskLen int, token []byte) string {
	// Generate mask bytes.
	mask := randomBytes(maskLen)

	// XOR
	tokenBytes := xorCsrfToken(token, mask)

	// Base64 encoding.
	tokenStr := base64.StdEncoding.EncodeToString(append(mask, tokenBytes...))

	return strings.Replace(tokenStr, "+", ".", -1)
}

func randomBytes(length int) []byte {
	if length < 0 {
		return nil
	}

	b := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return nil
	}

	return b
}

func validateCSRF(maskLen int, encodeToken, trueToken []byte) error {
	if len(encodeToken) <= maskLen {
		return errInvalidCRSF
	}

	// Restore the original base64 encoding string.
	encodeToken = bytes.Replace(encodeToken, []byte{'.'}, []byte{'+'}, -1)

	// Base64 decoding.
	decodeBytes := make([]byte, len(encodeToken))
	_, err := base64.StdEncoding.Decode(decodeBytes, encodeToken)
	if err != nil {
		return err
	}

	// Check length.
	trueTokenLen := len(trueToken)
	if len(decodeBytes) < (maskLen + trueTokenLen) {
		return errInvalidCRSF
	}

	// Get mask by maskLen.
	mask := decodeBytes[:maskLen]
	tokenBytes := make([]byte, trueTokenLen)
	copy(tokenBytes, decodeBytes[maskLen:maskLen+trueTokenLen])

	// XOR
	xorToken := xorCsrfToken(mask, trueToken)

	if bytes.Equal(xorToken, tokenBytes) {
		return nil
	}

	return errInvalidCRSF
}

func xorCsrfToken(token1, token2 []byte) []byte {
	len1 := len(token1)
	len2 := len(token2)
	if len1 > len2 {
		for i := 0; i < len1-len2; i++ {
			token2 = append(token2, token2[i%len2])
		}
	} else {
		for i := 0; i < len2-len1; i++ {
			if len1 == 0 {
				token1 = append(token1, ' ')
			} else {
				token1 = append(token1, token1[i%len1])
			}
		}
	}
	token := []byte{}
	for i := 0; i < len(token1); i++ {
		token = append(token, token1[i]^token2[i])
	}
	return token
}
