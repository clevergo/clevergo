# CleverGo [简体中文](README-ZH.md)
[![Build Status](https://travis-ci.org/clevergo/clevergo.svg?branch=master)](https://travis-ci.org/clevergo/clevergo)
[![Financial Contributors on Open Collective](https://opencollective.com/clevergo/all/badge.svg?label=financial+contributors)](https://opencollective.com/clevergo) [![Coverage Status](https://coveralls.io/repos/github/clevergo/clevergo/badge.svg?branch=master)](https://coveralls.io/github/clevergo/clevergo?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/clevergo/clevergo)](https://goreportcard.com/report/github.com/clevergo/clevergo)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue)](https://godoc.org/github.com/clevergo/clevergo)
[![Release](https://img.shields.io/github/release/clevergo/clevergo.svg?style=flat-square)](https://github.com/clevergo/clevergo/releases)

CleverGo is a lightweight, feature-rich and trie based high performance HTTP request router.

## Contents

- [Benchmark](#benchmark)
- [Features](#features)
- [Installation](#installation)
- [Examples](#examples)
- [Contribute](#contribute)

## Benchmark

- Date: 2020/03/13
- CPU: 4 Core
- RAM: 8G 
- Go: 1.14
- [Detail](BENCHMARK.md)

**Lower is better!**

[![Benchmark](https://razonyang.com/wp-content/uploads/2020/03/benchmark.png)](BENCHMARK.md)

## Features

- **High Performance:** see [Benchmark](#benchmark) shown above.
- **[Reverse Route Generation](#reverse-route-generation):** there are two ways to generate URL by a route: named route and matched route.
- **Route Group:** as known as subrouter, see [route group](#route-group).
- **Friendly to APIs:** it is easy to design [RESTful APIs](#restful-apis) and versioning your APIs by [route group](#route-group).
- **Middleware:** allow to plug middleware in route group or particular route, supports global middleware as well, see [middleware](#middleware) exmaple.
- **[Error Handler](#error-handler)** allow to custom error response, for example, display an error page.

## Installation

```shell
GO111MODULE=on go get github.com/clevergo/clevergo
```

or install particular [releases](https://github.com/clevergo/clevergo/releases):

```shell
GO111MODULE=on go get github.com/clevergo/clevergo@version
```

> Replace `version` to particular version.

## Examples

```go
package main

import (
	"fmt"
	"net/http"

	"github.com/clevergo/clevergo"
)

func home(ctx *clevergo.Context) error {
	ctx.WriteString("hello world")
	return nil
}

func hello(ctx *clevergo.Context) error {
	ctx.WriteString(fmt.Sprintf("hello %s", ctx.Params.String("name")))
	return nil
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
func (ctx *clevergo.Context) error {
	name := ctx.Params.String("name")
	page, err := ctx.Params.Int("page")
	num, err := ctx.Params.Int64("num")
	amount, err := ctx.Params.Uint64("amount")
	enable, err := ctx.Params.Bool("enable")
	price, err := ctx.Params.Float64("price")
	return err
}
```

`Router.UseRawPath` allows to match parameter that contains escaped slash `%2f`:

```go
router.UseRawPath = true
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
queryPost := func (ctx *clevergo.Context) error {
	// generates URL by matched route.
	url, _ := ctx.Route.URL("year", "2020", "month", "02", "slug", "hello world")
	return nil
}

router.Get("/posts/:year/:month/:slug", queryPost, router.RouteName("post"))

// generates URL by named route.
url, _ := router.URL("post", "year", "2020", "month", "02", "slug", "hello world")
```

### Error Handler

Error handler allow to custom error response.

```go
type MyErrorHandler struct {
	Tmpl *template.Template
}

func (meh MyErrorHandler) Handle(ctx *clevergo.Context, err error) {
	// display an error page.
	if err := meh.Tmpl.Execute(ctx.Response, err); err != nil {
		ctx.Error(err.Error(), http.StatusInternalServerError)
	}
}

router.ErrorHandler = MyErrorHandler{
	Tmpl: template.Must(template.New("error").Parse(`<html><body><h1>{{ .Error }}</h1></body></html>`)),
}
```

### Middleware

Middleware is a function `func(next Handle) Handle`.

**Built-in middlewares:**

- [Recovery](https://pkg.go.dev/github.com/clevergo/clevergo?tab=doc#Recovery)

**Example:**

```go
// global middlewares.
serverHeader := func(next clevergo.Handle) clevergo.Handle {
	func(ctx *clevergo.Context) error {
		// writes server header.
		ctx.Response.Header().Set("Server", "CleverGo")
		return next(ctx)
	}
}
router.Use(
	clevergo.Recovery(true),
	serverHeader,
	// ...
)

authenticator := func(next clevergo.Handle) clevergo.Handle {
	func(ctx *clevergo.Context) error {
		// authenticate returns an user instance and a boolean value indicates whether the provided credential is valid.
		if user, ok := authenticate(ctx); !ok {
			// returns an error if failed, in order to stop subsequent middlewares and handle.
			// you can also write response here, and return nil.
			return clevergo.NewError(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		}

		// share data between middlewares and handle.
		ctx.WithValue("user", user)
		return next(ctx)
	}
}

auth := func(ctx *clevergo.Context) error {
	ctx.WriteString(fmt.Sprintf("hello %v", ctx.Value("user")))
	return nil
}

router.Get("/auth", auth, RouteMiddleware(
	// middleware for current route.
	authenticator,
))

// use third-party global middleware, takes gorilla compress middleware as exmaple.
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

### Determine Request Method

```go
func (ctx *clevergo.Context) error {
	// equals to ctx.IsMethod(http.MethodGet).
	if ctx.IsGet() {

	}
	// other shortcuts:
	//ctx.IsDelete()
	//ctx.IsPatch()
	//ctx.IsPost()
	//ctx.IsPut()
	//ctx.IsOptions()
	return nil
}
```

## Contribute

- Give it a :star: and spread the package.
- [File an issue](https://github.com/clevergo/clevergo/issues/new) for features or bugs.
- Fork and make a pull request.

## Credit

- [julienschmidt/httprouter](https://github.com/julienschmidt/httprouter)

## Contributors

### Code Contributors

This project exists thanks to all the people who contribute. [[Contribute](CONTRIBUTING.md)].
<a href="https://github.com/clevergo/clevergo/graphs/contributors"><img src="https://opencollective.com/clevergo/contributors.svg?width=890&button=false" /></a>

### Financial Contributors

Become a financial contributor and help us sustain our community. [[Contribute](https://opencollective.com/clevergo/contribute)]

#### Individuals

<a href="https://opencollective.com/clevergo"><img src="https://opencollective.com/clevergo/individuals.svg?width=890"></a>

#### Organizations

Support this project with your organization. Your logo will show up here with a link to your website. [[Contribute](https://opencollective.com/clevergo/contribute)]
