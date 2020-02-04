# CleverGo [![Build Status](https://travis-ci.org/clevergo/clevergo.svg?branch=master)](https://travis-ci.org/clevergo/clevergo) [![Coverage Status](https://coveralls.io/repos/github/clevergo/clevergo/badge.svg?branch=master)](https://coveralls.io/github/clevergo/clevergo?branch=master) [![Go Report Card](https://goreportcard.com/badge/github.com/clevergo/clevergo)](https://goreportcard.com/report/github.com/clevergo/clevergo) [![GoDoc](https://godoc.org/github.com/clevergo/clevergo?status.svg)](http://godoc.org/github.com/clevergo/clevergo) [![Release](https://img.shields.io/github/release/clevergo/clevergo.svg?style=flat-square)](https://github.com/clevergo/clevergo/releases)

CleverGo is an enhanced version of [julienschmidt/httprouter](https://github.com/julienschmidt/httprouter) which provides 
[extra useful features](#extra-features).

## Contents

- [Extra Features](#extra-features)
- [Usage](#usage)
- [Differences](#difference)
- [Contribute](#contribute)
- [FAQ](#faq)

## Extra Features

- **Named Routes:** allow the reverse route generation of URLs.
- **Save Matched Route:** allow to retrieve matched route in handler, it is useful to generate URLs of the current route.
- **Nestable Route Groups:** as known as subrouter.
- **Middleware:** just a function `func(http.Handler) http.Handler`, it can not only integrates third-party middleware
    easily, but also can be used in three scopes: root router, subrouter and route.

## Usage

All usage and examples can be found at [GoDoc](http://godoc.org/github.com/clevergo/clevergo):

- [Application](https://godoc.org/github.com/clevergo/clevergo#example-Application)
- [Routes group / subrouter](https://godoc.org/github.com/clevergo/clevergo#example-RouteGroup)
- [Named routes / reverse route generation](https://godoc.org/github.com/clevergo/clevergo#example-Router-URL)
- [Serve static files](https://godoc.org/github.com/clevergo/clevergo#example-Router-ServeFiles)
- [Middleware](#middleware)

## Middleware

There are a lot of third-party middlewares can be used out of box, such as [clevergo/middleware](https://github.com/clevergo/middleware) and [gorilla/handlers](https://github.com/gorilla/handlers).

## Differences

> You can skip this section if you have not use httprouter before.

The usage of this package is very similar to httprouter, but there are serveral important differences you should pay attention for.

- There is no **Handle** type anymore, you can registers `http.Handler` and `http.HandlerFunc` by 
    [Router.Handle](https://godoc.org/github.com/clevergo/clevergo#Router.Handle) and 
    [Router.HandleFunc](https://godoc.org/github.com/clevergo/clevergo#Router.HandleFunc) respectively. And [GetParams](https://godoc.org/github.com/clevergo/clevergo#GetParams) is the only way to retrieve
    `Params` in handler.
- **Params.ByName** was renamed to [Params.Get](https://godoc.org/github.com/clevergo/clevergo#Params.Get).
- **ParamsFromContext** was removed, use [GetParams](https://godoc.org/github.com/clevergo/clevergo#GetParams) instead.
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
- Breaking compatibility for introduing features, see [differences](#differences) above.

Anyway, httprouter is definitely an awesome package.
