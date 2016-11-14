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

// Gzip gzip compress middleware.
type Gzip struct {
	level int
}

// NewGzip returns a Gzip middleware instance.
// See NewGzipWithSkipper.
//
// Level is the desired compression level:
//     * CompressNoCompression
//     * CompressBestSpeed
//     * CompressBestCompression
//     * CompressDefaultCompression
func NewGzip(level int) *Gzip {
	return &Gzip{
		level: level,
	}
}

// Handle implements Middleware's Handle function.
func (g *Gzip) Handle(next gem.Handler) gem.Handler {
	return gem.HandlerFunc(func(c *gem.Context) {
		defer fasthttp.CompressHandlerLevel(
			func(ctx *fasthttp.RequestCtx) {},
			g.level,
		)(c.RequestCtx)
		next.Handle(c)
	})
}
