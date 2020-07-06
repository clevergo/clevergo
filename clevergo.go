// Copyright 2020 CleverGo. All rights reserved.
// Use of this source code is governed by a MIT style license that can be found
// in the LICENSE file.

// Package clevergo is a trie based high performance HTTP request router.
package clevergo

import "clevergo.tech/log"

// Map is an alias of map[string]interface{}.
type Map map[string]interface{}

const serverName = "CleverGo"

var defaultLogger log.Logger = log.New()

// SetLogger sets default logger.
func SetLogger(logger log.Logger) {
	if logger == nil {
		panic("logger must not be empty")
	}
	defaultLogger = logger
}
