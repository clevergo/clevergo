Gem
===
![Gem logo](logo.png)

[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](https://godoc.org/github.com/go-gem/gem) 
[![Build Status](https://img.shields.io/travis/go-gem/gem.svg)](https://travis-ci.org/go-gem/gem) 
[![Go Report Card](https://img.shields.io/badge/go%20report-A+-brightgreen.svg?style=flat-square)](https://goreportcard.com/report/github.com/go-gem/gem) 
[![Coverage Status](https://img.shields.io/coveralls/go-gem/gem.svg)](https://coveralls.io/github/go-gem/gem?branch=master) 

Gem, a simple and fast web framework, it built top of [fasthttp](https://github.com/valyala/fasthttp).

currently, Gem API is **unstable** until the version v1.0.0 being released,
see [milestone](https://github.com/go-gem/gem/milestone/1) for more details.

The project inspired by third party packages, such as [fasthttp](https://github.com/valyala/fasthttp), [fasthttprouter](https://github.com/buaazp/fasthttprouter) and
[echo](https://github.com/labstack/echo), thier LICENSE can be found in LICENSE file.


### Install

```
go get github.com/go-gem/gem
```


### Example

See the [documentation](https://godoc.org/github.com/go-gem/gem) for more usages.

```
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
```

Run the code above, and then navigate to [127.0.0.1:8080](http://127.0.0.1:8080).

 
### Semantic Versioning

Gem follows [semantic versioning 2.0.0](http://semver.org/) managed through GitHub releases.


### Support Us

- :star: the project.

- Spread the word.

- [Contribute](#contribute) to the project.


### Contribute

- [Report issues](https://github.com/go-gem/gem/issues/new)

- Send PRs.

- Improve/fix documentation.


### Related Projects

1. [sessions](https://github.com/go-gem/sessions) Sessions manager for fasthttp.

2. [fasthttp](https://github.com/valyala/fasthttp) Fast HTTP package for Go.

3. [echo](https://github.com/labstack/echo) Fast and unfancy HTTP server framework.

4. [fasthttprouter](https://github.com/buaazp/fasthttprouter) A high performance fasthttp request router.


### About Name

The name means that this project aims to be Gem.

### LICENSE

MIT licensed. See the LICENSE file for details.