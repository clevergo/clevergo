Gem
===
![Gem logo](logo.png)

[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](https://godoc.org/github.com/go-gem/gem) 
[![Build Status](https://img.shields.io/travis/go-gem/gem.svg)](https://travis-ci.org/go-gem/gem) 
[![Go Report Card](https://img.shields.io/badge/go%20report-A+-brightgreen.svg?style=flat-square)](https://goreportcard.com/report/github.com/go-gem/gem) 
[![Coverage Status](https://img.shields.io/coveralls/go-gem/gem.svg)](https://coveralls.io/github/go-gem/gem?branch=master) 
[![license](https://img.shields.io/github/license/go-gem/gem.svg?style=flat-square)](LICENSE)

Gem, a simple and fast web framework, it built top of [fasthttp](https://github.com/valyala/fasthttp).

The project inspired by third party packages, such as [fasthttp](https://github.com/valyala/fasthttp), [fasthttprouter](https://github.com/buaazp/fasthttprouter) and
[echo](https://github.com/labstack/echo), thier LICENSE can be found in LICENSE file.

### Install
```
go get github.com/go-gem/gem
```


### Example

```
package main

import (
	"github.com/go-gem/gem"
	"github.com/labstack/gommon/log"
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


### Related Projects

1. [sessions](https://github.com/go-gem/sessions) Sessions manager for fasthttp.

2. [fasthttp](https://github.com/valyala/fasthttp) Fast HTTP package for Go.

3. [echo](https://github.com/labstack/echo) Fast and unfancy HTTP server framework.

4. [fasthttprouter](https://github.com/buaazp/fasthttprouter) A high performance fasthttp request router.

### LICENSE

MIT licensed. See the LICENSE file for details.