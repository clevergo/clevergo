// Copyright 2016 The Gem Authors. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package middleware

import (
	"github.com/go-gem/gem"
	"github.com/valyala/fasthttp"
)

// BodyLimit request body limit middleware.
type BodyLimit struct {
	Skipper Skipper

	// Maximum allowed size for a request body,
	// it's unit is byte.
	Limit int
}

// NewBodyLimit returns BodyLimit instance by the
// given limit.
func NewBodyLimit(limit int) *BodyLimit {
	return &BodyLimit{
		Limit: limit,
	}
}

// Handle implements Middleware's Handle function.
func (bl *BodyLimit) Handle(next gem.Handler) gem.Handler {
	if bl.Skipper == nil {
		bl.Skipper = defaultSkipper
	}

	return gem.HandlerFunc(func(ctx *gem.Context) {
		if bl.Skipper(ctx) {
			next.Handle(ctx)
			return
		}

		if ctx.Request.Header.ContentLength() > bl.Limit || len(ctx.RequestCtx.Request.Body()) > bl.Limit {
			ctx.RequestCtx.SetStatusCode(fasthttp.StatusRequestEntityTooLarge)
			ctx.RequestCtx.SetBodyString(fasthttp.StatusMessage(fasthttp.StatusRequestEntityTooLarge))
			return
		}

		next.Handle(ctx)
	})
}
