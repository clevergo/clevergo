# CleverGo 
[![Build Status](https://travis-ci.org/clevergo/clevergo.svg?branch=master)](https://travis-ci.org/clevergo/clevergo)
[![Coverage Status](https://coveralls.io/repos/github/clevergo/clevergo/badge.svg?branch=master)](https://coveralls.io/github/clevergo/clevergo?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/clevergo/clevergo)](https://goreportcard.com/report/github.com/clevergo/clevergo)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue)](https://pkg.go.dev/github.com/clevergo/clevergo)

[简体中文](README-ZH.md)

CleverGo is a lightweight, feature-rich and trie based high performance HTTP request router.

## Contents

- [Benchmark](#benchmark)
- [Features](#features)
- [Examples](#examples)
- [Contribute](#contribute)

## Benchmark

Date: 2020/02/11

**Lower is better!**

[![Benchmark](https://i.imgur.com/n8q1CCd.png)](https://github.com/razonyang/go-http-routing-benchmark)

## Features

- **High Performance:** see [Benchmark](#benchmark) shown above.
- **[Reverse Route Generation](#reverse-route-generation):** there are two ways to generate URL by a route: named route and matched route.
- **Route Group:** as known as subrouter, see [route group](#route-group).
- **Friendly to APIs:** it is easy to design [RESTful APIs](#restful-apis) and versioning your APIs by [route group](#route-group).
- **Middleware:** allow to plug middleware in route group or particular route, supports global middleware as well, see [middleware](#middleware) exmaple.

## Examples

```go
package main

import (
	"fmt"
	"net/http"

	"github.com/clevergo/clevergo"
)

func home(ctx *clevergo.Context) {
	ctx.WriteString("hello world")
}

func hello(ctx *clevergo.Context) {
	ctx.WriteString(fmt.Sprintf("hello %s", ctx.Params.String("name")))
}

func main() {
	router := clevergo.NewRouter()
	router.Get("/", home)
	router.Get("/hello/:name", hello)
	http.ListenAndServe(":8080", router)
}
```

### Params

There are some useful functions to retrieve the parameter value.

```go
func (ctx *clevergo.Context) {
	name := ctx.Params.String("name")
	page, err := ctx.Params.Int("page")
	num, err := ctx.Params.Int64("num")
	amount, err := ctx.Params.Uint64("amount")
	enable, err := ctx.Params.Bool("enable")
	price, err := ctx.Params.Float64("price")
}
```

### Static Files

```go
router.ServeFiles("/static/*filepath", http.Dir("/path/to/static"))

// sometimes, it is useful to treat http.FileServer as NotFoundHandler,
// such as "/favicon.ico".
router.NotFound = http.FileServer(http.Dir("public"))
```

### Reverse Route Generation

```go
queryPost := func (ctx *clevergo.Context) {
	// generates URL by matched route.
	url, _ := ctx.Route.URL("year", "2020", "month", "02", "slug", "hello world")
}

router.Get("/posts/:year/:month/:slug", queryPost, router.RouteName("post"))

// generates URL by named route.
url, _ := router.URL("post", "year", "2020", "month", "02", "slug", "hello world")
```

### Middleware

Middleware is a function defined as `func (clevergo.Handle) clevergo.Handle`.

```go
authenticator := func (handle clevergo.Handle) clevergo.Handle {
    return func(ctx *clevergo.Context) {
	// authenticate, terminate request if failed.
		
	// share data between middlewares and handle.
        ctx.WithValue("user", "foo")
        handle(ctx)
    }
}

router.Get("/auth", func(ctx *clevergo.Context) {
    ctx.WriteString(fmt.Sprintf("hello %s", ctx.Value("user")))
}, RouteMiddleware(
	// middleware for current route.
	authenticator,
))

// global middleware, takes gorilla compress middleware as exmaple.
http.ListenAndServe(":8080", handlers.CompressHandler(router))
```

Middleware also can be used in route group, see [Route Group](#route-group) for details.

### Route Group

```go
router := clevergo.NewRouter()

api := router.Group("/api", clevergo.RouteGroupMiddleware(
    // middlewares for APIs, such as CORS, authenticator, authorization
))

apiV1 := api.Group("/v1", clevergo.RouteGroupMiddleware(
    // middlewares for v1's APIs
))

apiV2 := api.Group("v2", clevergo.RouteGroupMiddleware(
    // middlewares for v2's APIs
))
```

### RESTful APIs

```go
router.Get("/users", queryUsers)
router.Post("/users", createUser)
router.Get("/users/:id", queryUser)
router.Put("/users/:id", updateUser)
router.Delete("/users/:id", deleteUser)
```

See [Route Group](#route-group) for versioning your APIs.

## Contribute

- Give it a :star: and spread the package.
- [File an issue](https://github.com/clevergo/clevergo/issues/new) for features or bugs.
- Fork and make a pull request.

## Credit

- [julienschmidt/httprouter](https://github.com/julienschmidt/httprouter)
