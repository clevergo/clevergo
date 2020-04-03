# CleverGo [English](README.md)
[![Build Status](https://travis-ci.org/clevergo/clevergo.svg?branch=master)](https://travis-ci.org/clevergo/clevergo)
[![Coverage Status](https://coveralls.io/repos/github/clevergo/clevergo/badge.svg?branch=master)](https://coveralls.io/github/clevergo/clevergo?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/clevergo/clevergo)](https://goreportcard.com/report/github.com/clevergo/clevergo)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue)](https://godoc.org/github.com/clevergo/clevergo)
[![Release](https://img.shields.io/github/release/clevergo/clevergo.svg?style=flat-square)](https://github.com/clevergo/clevergo/releases)

CleverGo 是一个轻量级、功能丰富和高性能的 HTTP 路由。

## 目录

- [基准测试](#基准测试)
- [功能特性](#功能特性)
- [安装](#安装)
- [举个栗子](#举个栗子)
- [贡献](#贡献)

## 基准测试

日期: 2020/02/11

- Date: 2020/03/13
- CPU: 4 Core
- RAM: 8G 
- Go: 1.14
- [详情](BENCHMARK.md)

**值越小性能越好**

[![Benchmark](https://i.imgur.com/Eato1pO.png)](BENCHMARK.md)

## 功能特性

- **高性能：** 参见[基准测试](#基准测试)。
- **[反向路由生成](#反向路由生成):** 可以通过**命名路由**和**匹配路由**生成 URL。
- **路由组:** 亦称子路由, 参看[路由组](#路由组)。
- **对 APIs 友好:** 很容易设计 [RESTful APIs](#restful-apis) 和通过[路由组](#路由组)进行 APIs 版本化。
- **中间件:** 可以在路由组或特定路由插入中间件，也可以使用全局中间件, 请参看[中间件](#中间件)例子。
- **[错误处理器](#错误处理器)** 可以自定义错误响应，比如显示一个错误页面。

## 安装

```shell
GO111MODULE=on go get github.com/clevergo/clevergo
```

## 举个栗子

```go
package main

import (
	"fmt"
	"net/http"

	"github.com/clevergo/clevergo"
)

func home(ctx *clevergo.Context) error {
	return ctx.String(http.StatusOK, "hello world")
}

func hello(ctx *clevergo.Context) error {
	return ctx.HTML(http.StatusOk, fmt.Sprintf("hello %s", ctx.Params.String("name")))
}

func main() {
	router := clevergo.NewRouter()
	router.Get("/", home)
	router.Get("/hello/:name", hello)
	http.ListenAndServe(":8080", router)
}
```

### 响应

```go
func text(ctx *clevergo.Context) error {
	return ctx.String(http.StatusOk, "hello world")
}

func html(ctx *clevergo.Context) error {
	return ctx.HTML(http.StatusOk, "<html><body>hello world</body></html>")
}

func json(ctx *clevergo.Context) error {
	// any type of data.
	data := map[string]interface{}{
		"message": "hello world",
	}
	return ctx.JSON(http.StatusOk, data)
}

func jsonp(ctx *clevergo.Context) error {
	// any type of data.
	data := map[string]interface{}{
		"message": "hello world",
	}
	// equals to ctx.JSONPCallback(http.StatusOk, "callback", data)
	return ctx.JSONP(http.StatusOk, data)
}

func xml(ctx *clevergo.Context) error {
	// any type of data.
	data := map[string]interface{}{
		"message": "hello world",
	}
	return ctx.XML(http.StatusOk, data)
}
```

### 参数

可以通过多种方式获取各种类型的参数值。

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

`Router.UseRawPath` 允许匹配带有空格转义符 `%2f` 的参数:

```go
router.UseRawPath = true
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
queryPost := func (ctx *clevergo.Context) error {
	// 通过匹配路由生成 URL
	url, _ := ctx.Route.URL("year", "2020", "month", "02", "slug", "hello world")
	return nil
}

router.Get("/posts/:year/:month/:slug", queryPost, router.RouteName("post"))

// 通过命名路由生成 URL
url, _ := router.URL("post", "year", "2020", "month", "02", "slug", "hello world")
```

### 错误处理器

错误处理器可以自定义错误响应。

```go
type MyErrorHandler struct {
	Tmpl *template.Template
}

func (meh MyErrorHandler) Handle(ctx *clevergo.Context, err error) {
	// 显示一个错误页面。
	if err := meh.Tmpl.Execute(ctx.Response, err); err != nil {
		ctx.Error(err.Error(), http.StatusInternalServerError)
	}
}

router.ErrorHandler = MyErrorHandler{
	Tmpl: template.Must(template.New("error").Parse(`<html><body><h1>{{ .Error }}</h1></body></html>`)),
}
```

### 中间件

中间件是一个 `func(next Handle) Handle` 函数。[WrapHH](https://pkg.go.dev/github.com/clevergo/clevergo?tab=doc#WrapHH) 是一个将 `func(http.Handler) http.Handler` 转化成中间件的适配器。

**内置中间件：**

- [Recovery](https://pkg.go.dev/github.com/clevergo/clevergo?tab=doc#Recovery)

**例子：**

```go
// 全局中间件.
serverHeader := func(next clevergo.Handle) clevergo.Handle {
	return func(ctx *clevergo.Context) error {
		ctx.Response.Header().Set("Server", "CleverGo")
		return next(ctx)
	}
}
router.Use(
	clevergo.Recovery(true),
	serverHeader,

	// third-party func(http.Handler) http.Handler middlewares
	clevergo.WrapHH(gziphandler.GzipHandler) // https://github.com/nytimes/gziphandler

	// ...
)

authenticator := func(next clevergo.Handle) clevergo.Handle {
	return func(ctx *clevergo.Context) error {
		// authenticate 返回一个 user 和一个布尔值表示提供的凭证是否有效。
		if user, ok := authenticate(ctx); !ok {
			// 返回一个错误，以终止后续的中间件和 Handle。
			// 也可以在这发送响应，并返回 nil。
			return clevergo.NewError(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		}

		// 在中间件之间共享数据。
		ctx.WithValue("user", user)
		return next(ctx)
	}
}

auth := func(ctx *clevergo.Context) error {
	ctx.WriteString(fmt.Sprintf("hello %v", ctx.Value("user")))
	return nil
}

router.Get("/auth", auth, RouteMiddleware(
	// 中间件，只在当前路由生效。
	authenticator,
))

http.ListenAndServe(":8080", router)
```

中间件也可以在[路由组](#路由组)中使用。

### 路由组

```go
router := clevergo.NewRouter()

api := router.Group("/api", clevergo.RouteGroupMiddleware(
	// APIs 的中间件，如：CORS、身份验证、授权验证等。
	clevergo.WrapHH(cors.Default().Handler), // https://github.com/rs/cors
))

apiV1 := api.Group("/v1", clevergo.RouteGroupMiddleware(
    // middlewares for v1's APIs
))

apiV2 := api.Group("/v2", clevergo.RouteGroupMiddleware(
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

### 判断请求方法

```go
func (ctx *clevergo.Context) error {
	// 等同于 ctx.IsMethod(http.MethodGet).
	if ctx.IsGet() {

	}
	// 其他方法:
	//ctx.IsDelete()
	//ctx.IsPatch()
	//ctx.IsPost()
	//ctx.IsPut()
	//ctx.IsOptions()
	//ctx.IsAJAX()
	return nil
}
```

## 贡献

- 给颗 :star:。
- [提交问题](https://github.com/clevergo/clevergo/issues/new) 以提问、请求新特性或者反馈 Bug。
- Fork 和提交 PR。
