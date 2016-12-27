# Gem Web Framework [![GoDoc](https://godoc.org/github.com/go-gem/gem?status.svg)](https://godoc.org/github.com/go-gem/gem) [![Build Status](https://travis-ci.org/go-gem/gem.svg?branch=master)](https://travis-ci.org/go-gem/gem) [![Go Report Card](https://goreportcard.com/badge/github.com/go-gem/gem)](https://goreportcard.com/report/github.com/go-gem/gem) [![Coverage Status](https://coveralls.io/repos/github/go-gem/gem/badge.svg?branch=master)](https://coveralls.io/github/go-gem/gem?branch=master) [![Join the chat at https://gitter.im/go-gem/gem](https://badges.gitter.im/go-gem/gem.svg)](https://gitter.im/go-gem/gem?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

Gem, a simple and fast web framework written in Go(golang).

The current version is `2.0.0.alpha`, it locate at branch `master`, the old version is located at `v1`(deprecated), see [changes](#changes). 
The APIs is currently unstable until the stable version `2.0.0` being released.


## Install

```
go get github.com/go-gem/gem
```


## Features

- High performance
- Friendly to REST API
- HTTP/2 support
- Leveled logging
- Middlewares


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