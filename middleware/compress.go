// Copyright 2016 The Gem Authors. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package middleware

import (
	"github.com/go-gem/gem"
	"github.com/valyala/fasthttp"
)

// Supported compression levels.
const (
	CompressNoCompression      = fasthttp.CompressNoCompression
	CompressBestSpeed          = fasthttp.CompressBestSpeed
	CompressBestCompression    = fasthttp.CompressBestCompression
	CompressDefaultCompression = fasthttp.CompressDefaultCompression
)

// Compress compress middleware.
type Compress struct {
	level int
}

// NewCompress returns a Compress middleware instance.
//
// Level is the desired compression level:
//     * CompressNoCompression
//     * CompressBestSpeed
//     * CompressBestCompression
//     * CompressDefaultCompression
func NewCompress(level int) *Compress {
	return &Compress{
		level: level,
	}
}

// Handle implements Middleware's Handle function.
func (m *Compress) Handle(next gem.Handler) gem.Handler {
	return gem.HandlerFunc(func(ctx *gem.Context) {
		defer fasthttp.CompressHandlerLevel(
			func(ctx *fasthttp.RequestCtx) {},
			m.level,
		)(ctx.RequestCtx)

		next.Handle(ctx)
	})
}
