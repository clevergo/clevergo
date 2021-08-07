# CleverGo
[![Build Status](https://img.shields.io/travis/clevergo/clevergo?style=flat-square)](https://travis-ci.com/clevergo/clevergo)
[![Coverage Status](https://img.shields.io/coveralls/github/clevergo/clevergo?style=flat-square)](https://coveralls.io/github/clevergo/clevergo)
[![Go Report Card](https://goreportcard.com/badge/github.com/clevergo/clevergo?style=flat-square)](https://goreportcard.com/report/github.com/clevergo/clevergo)
[![Go.Dev reference](https://img.shields.io/badge/go.dev-reference-blue?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/clevergo.tech/clevergo?tab=doc)
[![Release](https://img.shields.io/github/release/clevergo/clevergo.svg?style=flat-square)](https://github.com/clevergo/clevergo/releases)
[![Downloads](https://img.shields.io/endpoint?url=https://pkg.clevergo.tech/api/badges/downloads/total/clevergo.tech/clevergo&style=flat-square)](https://pkg.clevergo.tech/clevergo.tech/clevergo)
[![Chat](https://img.shields.io/badge/chat-telegram-blue?style=flat-square)](https://t.me/clevergotech)
[![Community](https://img.shields.io/badge/community-forum-blue?style=flat-square&color=orange)](https://forum.clevergo.tech)

CleverGo is a lightweight, feature rich and trie based high performance HTTP request router.

```shell
go get -u clevergo.tech/clevergo
```

- [English](https://clevergo.tech/en/)
- [简体中文](https://clevergo.tech/zh/)

[![Benchmark](https://clevergo.tech/img/benchmark.png)](https://clevergo.tech/docs/benchmark)

## Features

- **Full features of HTTP router**.
- **High Performance:** extremely fast, see [Benchmark](https://clevergo.tech/docs/benchmark).
- **Gradual learning curve:** you can learn the entire usages by going through the [documentation](#documentation) in half an hour.
- **[Reverse Route Generation](https://clevergo.tech/docs/routing/url-generation):** allow generating URLs by named route or matched route.
- **[Route Group](https://clevergo.tech/docs/routing/route-group):** as known as subrouter.
- **Friendly to APIs:** it is easy to design RESTful APIs and versioning your APIs by route group.
- **[Middleware](https://clevergo.tech/docs/middleware):** plug middleware in route group or particular route, supports global middleware as well. Compatible with most of third-party middleware.
- **[Logger](https://clevergo.tech/docs/logger):** a generic logger interface, supports [zap](https://github.com/uber-go/zap) and [logrus](http://github.com/sirupsen/logrus). Logger can be used in middleware or handler.
- ...

## Examples

Checkout [example](https://github.com/clevergo/examples) for details.

## Contribute

Contributions are welcome.

- Star it and spread the package.
- [File an issue](https://github.com/clevergo/clevergo/issues/new) to ask questions, request features or report bugs.
- Fork and make a pull request.
- Improve [documentations](https://github.com/clevergo/website).

## Credit

See [CREDIT.md](CREDIT.md).
