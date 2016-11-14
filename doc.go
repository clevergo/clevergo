// Copyright 2016 The Gem Authors. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

/*
Package gem, a simple and fast web framework,
for building web or restful application.

Install
	go get github.com/go-gem/gem

Example
	package main
	import (
		"log"

		"github.com/go-gem/gem"
		"github.com/valyala/fasthttp"
	)

	func main() {
		server := gem.New()

		router := gem.NewRouter()

		router.GET("/", func(c *gem.Context) {
			c.HTML(fasthttp.StatusOK, "Hello world.")
		})

		log.Fatal(server.ListenAndServe(":8080", router.Handler))
	}
*/
package gem
