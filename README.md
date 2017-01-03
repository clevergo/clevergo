# Gem Web Framework [![GoDoc](https://godoc.org/github.com/go-gem/gem?status.svg)](https://godoc.org/github.com/go-gem/gem) [![Build Status](https://travis-ci.org/go-gem/gem.svg?branch=master)](https://travis-ci.org/go-gem/gem) [![Go Report Card](https://goreportcard.com/badge/github.com/go-gem/gem)](https://goreportcard.com/report/github.com/go-gem/gem) [![Coverage Status](https://coveralls.io/repos/github/go-gem/gem/badge.svg?branch=master)](https://coveralls.io/github/go-gem/gem?branch=master) [![Join the chat at https://gitter.im/go-gem/gem](https://badges.gitter.im/go-gem/gem.svg)](https://gitter.im/go-gem/gem?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

Gem, a simple and fast web framework written in Go(golang), required `go1.8` or above.

The current version is `2.0.0-beta`, the old version `v1` is deprecated, see [changes](#changes). 
The APIs is currently unstable until the stable version `2.0.0` and `go1.8` being released.


## Features

- High performance
- Friendly to REST API
- Full test of all APIs [![Coverage Status](https://coveralls.io/repos/github/go-gem/gem/badge.svg?branch=master)](https://coveralls.io/github/go-gem/gem?branch=master)
- Pretty and fast router - the router is custom version of [httprouter](https://github.com/julienschmidt/httprouter)
- HTTP/2 support - HTTP/2 server push are supported since `go1.8`
- Leveled logging - included four levels `debug`, `info`, `error` and `fatal`, there are many third-party implements the Logger: 
    - [logrus](https://github.com/Sirupsen/logrus) - Structured, pluggable logging for Go
    - [go-logging](https://github.com/op/go-logging) - Golang logging library
    - [gem-log](https://github.com/go-gem/log) - default logger, maintained by Gem Authors
- [CSRF Middleware](https://github.com/go-gem/middleware-csrf) - Cross-Site Request Forgery protection
- [CORS Middleware](https://github.com/go-gem/middleware-cors) -  Cross-Origin Resource Sharing
- [AUTH Middleware](https://github.com/go-gem/middleware-auth) - HTTP Basic and HTTP Digest authentication
- [JWT Middleware](https://github.com/go-gem/middleware-jwt) - JSON WEB TOKEN authentication
- [Compress Middleware](https://github.com/go-gem/middleware-compress) - Compress response body
- [Request Body Limit Middleware](https://github.com/go-gem/middleware-body-limit) - limit request body maximum size
- Frozen APIs since the stable version `2.0.0` was released


## Getting Started

### Install

**This package requires `go1.8` or above.**
 
```
$ go get -u github.com/go-gem/gem
```

### Quick Start

```
package main

import (
    "log"

    "github.com/go-gem/gem"
)

func index(ctx *gem.Context) {
    ctx.HTML(200, "hello world")
}

func main() {
    // Create server.
    srv := gem.New(":8080")
    
    // Create router.
    router := gem.NewRouter()
    // Register handler
    router.GET("/", index)
    
    // Start server.
    log.Println(srv.ListenAndServe(router.Handler()))
}
```

### Logger

AFAIK, the following leveled logging packages are compatible with Gem web framework:

- [logrus](https://github.com/Sirupsen/logrus) - structured, pluggable logging for Go
- [go-logging](https://github.com/op/go-logging) - golang logging library
- [gem-log](https://github.com/go-gem/log) - default logger, maintained by Gem Authors

[Logger](https://godoc.org/github.com/go-gem/gem#Logger) includes four levels: `debug`, `info`, `error` and `fatal`, their APIs are 
`Debug` and `Debugf`, `Info` and `Infof`, `Error` and `Errorf`, `Fatal` and `Fatalf`.

We take `logrus` as example to show that how to set and use logger.

```
// set logrus logger as server's logger.
srv.SetLogger(logrus.New())

// we can use it in handler.
router.GET("/logger", func(ctx *gem.Context) {
		ctx.Logger().Debug("debug")
		ctx.Logger().Info("info")
		ctx.Logger().Error("error")
})
```

### Static Files

```
router.ServeFiles("/tmp/*filepath", http.Dir(os.TempDir()))
```

Note: the path(first parameter) must end with `*filepath`.

### REST APIs

The router is friendly to REST APIs, we take `users` handler as example to show that how to build a RESTful server.

```
// register user list handler.
router.GET("/users", func(ctx *gem.Context) {
    ctx.JSON(200, userlist)    
})

// register user add handler.
router.POST("/users", func(ctx *gem.Contexy) {
    ctx.Request.ParseForm()
    name := ctx.Request.FormValue("name")
    
    // add user
    
    ctx.JSON(200, msg)
})

// register user profile handler.
router.GET("/users/:name", func(ctx *gem.Context) {
    // firstly, we need get the username from the URL query.
    name, ok := ctx.UserValue("name").(string)
    if !ok {
        ctx.JSON(404, userNotFound)
        return
    }
    
    // return user profile.
    ctx.JSON(200, userProfileByName(name))
})

// register user profile update handler.
router.PUT("/users/:name", func(ctx *gem.Context) {
    // firstly, we need get the username from the URL query.
    name, ok := ctx.UserValue("name").(string)
    if !ok {
        ctx.JSON(404, userNotFound)
        return
    }
    
    // get nickname
    ctx.Request.ParseForm()
    nickname := ctx.Request.FormValue("nickname")
    
    // update user nickname.
    
    ctx.JSON(200, msg)
})

// register user delete handler.
router.DELETE("/users/:name", func(ctx *gem.Context) {
    // firstly, we need get the username from the URL query.
    name, ok := ctx.UserValue("name").(string)
    if !ok {
        ctx.JSON(404, userNotFound)
        return
    }
    
    // delete user.
    
    ctx.JSON(200, msg)
}
```

### HTTP/2 Server Push

See https://github.com/go-gem/examples/tree/master/http2.

```
router.GET("/", func(ctx *gem.Context) {
	if err := ctx.Push("/images/logo.png", nil); err != nil {
		ctx.Logger().Info(err)
	}

	ctx.HTML(200, `<html><head></head><body><img src="/images/logo.png"/></body></html>`)
})
router.ServeFiles("/images/*filepath", http.Dir(imagesDir))
```

### Use Middleware

It is easy to implement a middleware, see [Middleware](https://godoc.org/github.com/go-gem/gem#Middleware) interface,
you just need to implement the `Wrap` function.

```
type Middleware interface {
    Wrap(next Handler) Handler
}
```

For example, we defined a simple debug middleware:

```
type Debug struct{}

// Wrap implements the Middleware interface.
func (d *Debug) Wrap(next gem.Handler) gem.Handler {
    // gen.HandlerFunc is adapter like http.HandlerFunc.
	return gem.HandlerFunc(func(ctx *gem.Context) {
		// print request info.
		log.Println(ctx.Request.URL, ctx.Request.Method)

		// call the next handler.
		next.Handle(ctx)
	})
}
```

and then we should register it:

register the middleware for all handlers via [Router.Use](https://godoc.org/github.com/go-gem/gem#Router.Use).

```
router.Use(&Debug{})
```

we can also register the middleware for specific handler via [HandlerOption](https://godoc.org/github.com/go-gem/gem#HandlerOption). 

```
router.GET("/specific", specificHandler, &gem.HandlerOption{Middlewares:[]gem.Middleware{&Debug{}}})
```

Gem also provides some frequently used middlewares, such as: 

- [CSRF Middleware](https://github.com/go-gem/middleware-csrf) - Cross-Site Request Forgery protection

- [CORS Middleware](https://github.com/go-gem/middleware-cors) -  Cross-Origin Resource Sharing

- [AUTH Middleware](https://github.com/go-gem/middleware-auth) - HTTP Basic and HTTP Digest authentication

- [JWT Middleware](https://github.com/go-gem/middleware-jwt) - JSON WEB TOKEN authentication

- [Compress Middleware](https://github.com/go-gem/middleware-compress) - Compress response body

- [Request Body Limit Middleware](https://github.com/go-gem/middleware-body-limit) - limit request body maximum size

### Share data between middlewares

Context provides two useful methods: `SetUserValue` and `UserValue` to share data between middlewares.

```
// Store data into context in one middleware
ctx.SetUserValue("name", "foo")

// Get data from context in other middleware or hander
ctx.UserValue("name")
```

## Semantic Versioning

Gem follows [semantic versioning 2.0.0](http://semver.org/) managed through GitHub releases.


## Support Us

- :star: the project.
- Spread the word.
- [Contribute](#contribute) to the project.


## Contribute

- [Report issues](https://github.com/go-gem/gem/issues/new)
- Send PRs.
- Improve/fix documentation.

**We’re always looking for help, so if you would like to contribute, we’d love to have you!**


## Changes

The `v2` and `v1` are totally different:

- `v2` built on top of `net/http`.
 
- `v2` require `go1.8` or above.

- `v2` is compatible with `Windows`.


## FAQ

- Why choose `net/http` instead of `fasthttp`?

    1. `net/http` has much more third-party packages than `fasthttp`.
    
    2. `fasthttp` doesn't support `HTTP/2` yet.


## LICENSE

BSD 3-Clause License, see [LICENSE](LICENSE) and [AUTHORS](AUTHORS.md).

**Inspiration & Credits**

For respecting the third party packages, I added their author into [AUTHORS](AUTHORS.md), and listed those packages here.

- [**httprouter**](https://github.com/julienschmidt/httprouter) - [LICENSE](https://github.com/julienschmidt/httprouter/blob/master/LICENSE).
    Gem's router is a custom version of `httprouter`, thanks to `httprouter`.