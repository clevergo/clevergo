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
	"time"

	"github.com/go-gem/gem"
	"github.com/valyala/fasthttp"
)

// CSRF default configuration.
var (
	CSRFSafeMethods = []string{gem.StrMethodGet, gem.StrMethodHead, gem.StrMethodOptions}

	CSRFAcquireToken = func(ctx *gem.Context) (token []byte) {
		if token = ctx.RequestCtx.Request.Header.Peek("X-CSRF-Token"); len(token) == 0 {
			token = ctx.RequestCtx.FormValue("_csrf")
		}

		return
	}

	CSRFMaskLen = 8

	CSRFTokenLen = 32
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

	// AcquireToken acquire encoded token.
	AcquireToken func(ctx *gem.Context) []byte
}

// NewCSRF returns a CSRF instance with the default
// configuration.
func NewCSRF() *CSRF {
	return &CSRF{
		Skipper:      defaultSkipper,
		SafeMethods:  CSRFSafeMethods,
		AcquireToken: CSRFAcquireToken,
		MaskLen:      CSRFMaskLen,
		TokenLen:     CSRFTokenLen,
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

	safeMethods := make(map[string]bool, len(m.SafeMethods))
	for _, method := range m.SafeMethods {
		safeMethods[method] = true
	}

	return gem.HandlerFunc(func(ctx *gem.Context) {
		var trueToken []byte
		trueTokenStr := string(ctx.Request.Header.Cookie("_csrf"))
		if trueTokenStr != "" {
			if token, err := base64.StdEncoding.DecodeString(trueTokenStr); err == nil && len(token) >= m.TokenLen {
				trueToken = token[:m.TokenLen]
			}
		}

		if len(trueToken) == 0 {
			trueToken = randomBytes(m.TokenLen)
			cookie := &fasthttp.Cookie{}
			cookie.SetKey("_csrf")
			cookie.SetValue(base64.StdEncoding.EncodeToString(trueToken))
			cookie.SetExpire(time.Now().Add(time.Minute * 10))
			ctx.Response.Header.SetCookie(cookie)
		}

		encodedToken := generateCSRFToken(m.MaskLen, trueToken)
		ctx.SetUserValue("_csrf", encodedToken)

		method := ctx.MethodString()
		if _, safe := safeMethods[method]; safe {
			next.Handle(ctx)
			return
		}

		// Verify CSRF token
		if err := validateCSRF(m.MaskLen, m.AcquireToken(ctx), trueToken); err != nil {
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
