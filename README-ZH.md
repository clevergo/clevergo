# CleverGo 
[![Build Status](https://travis-ci.org/clevergo/clevergo.svg?branch=master)](https://travis-ci.org/clevergo/clevergo)
[![Coverage Status](https://coveralls.io/repos/github/clevergo/clevergo/badge.svg?branch=master)](https://coveralls.io/github/clevergo/clevergo?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/clevergo/clevergo)](https://goreportcard.com/report/github.com/clevergo/clevergo)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue)](https://pkg.go.dev/github.com/clevergo/clevergo)

[English](README.md)

CleverGo 是一个轻量级、功能丰富和高性能的 HTTP 路由。

## 目录

- [基准测试](#基准测试)
- [功能特性](#功能特性)
- [举个栗子](#举个栗子)
- [贡献](#贡献)

## 基准测试

日期: 2020/02/11

**越小性能越好**

[![Benchmark](https://i.imgur.com/n8q1CCd.png)](https://github.com/razonyang/go-http-routing-benchmark)

## 功能特性

- **高性能：** 参见[基准测试](#基准测试)。
- **[反向路由生成](#反向路由生成):** 可以通过**命名路由**和**匹配路由**生成 URL。
- **路由组:** 亦称子路由, 参看[路由组](#路由组)。
- **对 APIs 友好:** 很容易设计 [RESTful APIs](#restful-apis) 和通过[路由组](#路由组)进行 APIs 版本化。
- **中间件:** 可以在路由组或特定路由插入中间件，也可以使用全局中间件, 请参看[中间件](#中间件)例子。

## 举个栗子

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

### 参数

可以通过多种方式获取各种类型的参数值。

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

### 静态文件

```go
router.ServeFiles("/static/*filepath", http.Dir("/path/to/static"))

// 有时候，将 http.FileServer 作为 NotFound 处理器很有用处。
// 比如 "/favicon.ico"。
router.NotFound = http.FileServer(http.Dir("public"))
```

### 反向路由生成

```go
queryPost := func (ctx *clevergo.Context) {
	// 通过匹配路由生成 URL
	url, _ := ctx.Route.URL("year", "2020", "month", "02", "slug", "hello world")
}

router.Get("/posts/:year/:month/:slug", queryPost, router.RouteName("post"))

// 通过命名路由生成 URL
url, _ := router.URL("post", "year", "2020", "month", "02", "slug", "hello world")
```

### 中间件

中间件是一个定义为 `func (clevergo.Handle) clevergo.Handle` 的函数。

```go
authenticator := func (handle clevergo.Handle) clevergo.Handle {
    return func(ctx *clevergo.Context) {
		// 身份验证, 验证失败则终止请求。
		
		// 在中间件间共享数据。
        ctx.WithValue("user", "foo")
        handle(ctx)
    }
}

router.Get("/auth", func(ctx *clevergo.Context) {
    ctx.WriteString(fmt.Sprintf("hello %s", ctx.Value("user")))
}, RouteMiddleware(
	// 中间件，只在当前路由生效。
	authenticator,
))

// 全局路由，以 gorilla compress 中间件为例。
http.ListenAndServe(":8080", handlers.CompressHandler(router))
```

中间件也可以在[路由组](#路由组)中使用。

### 路由组

```go
router := clevergo.NewRouter()

api := router.Group("/api", clevergo.RouteGroupMiddleware(
    // APIs 的中间件，如：CORS、身份验证、授权验证等。
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

可以通过[路由组](#路由组)对你的 APIs 进行版本化。

## 贡献

- 给颗 :star:。
- [提交问题](https://github.com/clevergo/clevergo/issues/new) 以报告 Bug 或者请求新特性。
- Fork 和提交 PR。
