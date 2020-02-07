# CleverGo [![Build Status](https://travis-ci.org/clevergo/clevergo.svg?branch=master)](https://travis-ci.org/clevergo/clevergo) [![Coverage Status](https://coveralls.io/repos/github/clevergo/clevergo/badge.svg?branch=master)](https://coveralls.io/github/clevergo/clevergo?branch=master) [![Go Report Card](https://goreportcard.com/badge/github.com/clevergo/clevergo)](https://goreportcard.com/report/github.com/clevergo/clevergo) [![GoDoc](https://img.shields.io/badge/godoc-reference-blue)](https://pkg.go.dev/github.com/clevergo/clevergo) [![Release](https://img.shields.io/github/release/clevergo/clevergo.svg?style=flat-square)](https://github.com/clevergo/clevergo/releases)

CleverGo is an enhanced version of [julienschmidt/httprouter](https://github.com/julienschmidt/httprouter) which provides 
[extra useful features](#extra-features).

## Contents

- [Motivation](#motivation)
- [Features](#features)
- [Usage](#usage)
- [Differences](#difference)
- [Contribute](#contribute)
- [FAQ](#faq)

## Motivation

CleverGo **isn't an another web framework**, it aims to be a **lightweight**, **feature-rich** and **high performance** HTTP router which can be intergrates with any third-party packages(such as HTTP middlewares) easily.

## Features

**Only explicit matches:** With other routers, like [`http.ServeMux`](https://golang.org/pkg/net/http/#ServeMux), a requested URL path could match multiple patterns. Therefore they have some awkward pattern priority rules, like *longest match* or *first registered, first matched*. By design of this router, a request can only match exactly one or no route. As a result, there are also no unintended matches, which makes it great for SEO and improves the user experience.

**Stop caring about trailing slashes:** Choose the URL style you like, the router automatically redirects the client if a trailing slash is missing or if there is one extra. Of course it only does so, if the new path has a handler. If you don't like it, you can [turn off this behavior](https://pkg.go.dev/github.com/clevergo/clevergo#Router.RedirectTrailingSlash).

**Path auto-correction:** Besides detecting the missing or additional trailing slash at no extra cost, the router can also fix wrong cases and remove superfluous path elements (like `../` or `//`). Is [CAPTAIN CAPS LOCK](http://www.urbandictionary.com/define.php?term=Captain+Caps+Lock) one of your users? HttpRouter can help him by making a case-insensitive look-up and redirecting him to the correct URL.

**Parameters in your routing pattern:** Stop parsing the requested URL path, just give the path segment a name and the router delivers the dynamic value to you. Because of the design of the router, path parameters are very cheap.

**Zero Garbage:** The matching and dispatching process generates zero bytes of garbage. The only heap allocations that are made are building the slice of the key-value pairs for path parameters, and building new context and request objects (the latter only in the standard `Handler`/`HandlerFunc` API). In the 3-argument API, if the request path contains no parameters not a single heap allocation is necessary.

**High Performance**

**Perfect for APIs:** The router design encourages to build sensible, hierarchical RESTful APIs. Moreover it has built-in native support for [OPTIONS requests](http://zacstewart.com/2012/04/14/http-options-method.html) and `405 Method Not Allowed` replies.

Of course you can also set **custom [`NotFound`](https://pkg.go.dev/github.com/clevergo/clevergo#Router.NotFound) and  [`MethodNotAllowed`](https://pkg.go.dev/github.com/clevergo/clevergo#Router.MethodNotAllowed) handlers** and [**serve static files**](https://pkg.go.dev/github.com/clevergo/clevergo#Router.ServeFiles).

### Extra Features

- **Named Routes:** allow the reverse route generation of URLs.
- **Save Matched Route:** allow to retrieve matched route in handler, it is useful to generate URLs of the current route.
- **Nestable Route Groups:** as known as subrouter.
- **Middleware:** just a function `func(http.Handler) http.Handler`, it can not only integrates third-party middleware
    easily, but also can be used in three scopes: root router, subrouter and route.

## Usage

All usage and examples can be found at [GoDoc](https://pkg.go.dev/github.com/clevergo/clevergo):

- [Application](https://pkg.go.dev/github.com/clevergo/clevergo#example-Application)
- [Routes group / subrouter](https://pkg.go.dev/github.com/clevergo/clevergo#example-RouteGroup)
- [Named routes / reverse route generation](https://pkg.go.dev/github.com/clevergo/clevergo#example-Router.URL)
- [Serve static files](https://pkg.go.dev/github.com/clevergo/clevergo#example-Router.ServeFiles)
- [Middleware](#middleware)

## Middleware

There are a lot of third-party middlewares can be used out of box, such as:

- [clevergo/middleware](https://github.com/clevergo/middleware): a collection of HTTP middleware, adapter for gorilla handlers(compress and logging).
- [gorilla/handlers](https://github.com/gorilla/handlers): a collection of useful middleware for Go HTTP services & web applications.
- [goji/httpauth](https://github.com/goji/httpauth): basic auth middleware.
- List other middlewares here by PR.

### Chaining

[Chain](https://pkg.go.dev/github.com/clevergo/clevergo#Chain) allow to attach any middlewares on a `http.Handler`.

## Differences

> You can skip this section if you have not use httprouter before.

The usage of this package is very similar to httprouter, but there are serveral important differences you should pay attention for.

- There is no **Handle** type anymore, you can registers `http.Handler` and `http.HandlerFunc` by 
    [Router.Handle](https://pkg.go.dev/github.com/clevergo/clevergo#Router.Handle) and 
    [Router.HandleFunc](https://pkg.go.dev/github.com/clevergo/clevergo#Router.HandleFunc) respectively. And [GetParams](https://pkg.go.dev/github.com/clevergo/clevergo#GetParams) is the only way to retrieve
    `Params` in handler.
- **Params.ByName** was renamed to [Params.Get](https://pkg.go.dev/github.com/clevergo/clevergo#Params.Get), and added some useful functions for converting value type:
    - [Params.Int](https://pkg.go.dev/github.com/clevergo/clevergo#Params.Int)
    - [Params.Int64](https://pkg.go.dev/github.com/clevergo/clevergo#Params.Int64)
    - [Params.Uint64](https://pkg.go.dev/github.com/clevergo/clevergo#Params.Uint64)
    - [Params.Float64](https://pkg.go.dev/github.com/clevergo/clevergo#Params.Float64)
    - [Params.Bool](https://pkg.go.dev/github.com/clevergo/clevergo#Params.Bool)
- **ParamsFromContext** was removed, use [GetParams](https://pkg.go.dev/github.com/clevergo/clevergo#GetParams) instead.
- **Router.PanicHandler** was removed, it is more reasonable to use RecoveryMiddleware in the top level instead.
- `Router` methods `GET`, `POST`, `PUT`, `DELETE`, `PATCH`, `HEAD`, `OPTIONS` were renamed to `Get`, `Post`,
    `Put`, `Delete`, `Patch`, `Head`, `Options` respectively.

## Contribute

- Give it a :star: and spread the package.
- [File an issue](https://github.com/clevergo/clevergo/issues/new) for features or bugs.
- Fork and make a pull request.

## FAQ

### Why not contribute to the original repository?

There are multiple reasons:

- Slow maintenance, such as the PR of [subrouter](https://github.com/julienschmidt/httprouter/pull/89) is still unaccepted.
- Missing some [useful features](#extra-features) I mentioned above.
- Breaking compatibility for introducing features, see [differences](#differences) above.

Anyway, httprouter is definitely an awesome package.
