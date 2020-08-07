// Copyright 2020 CleverGo. All rights reserved.
// Use of this source code is governed by a MIT style license that can be found
// in the LICENSE file.

// Package clevergo is a trie based high performance HTTP request router.
package clevergo

import (
	stdlog "log"
	"os"

	"clevergo.tech/log"
)

// Map is an alias of map[string]interface{}.
type Map map[string]interface{}

const serverName = "CleverGo"

var logger log.Logger = log.New(os.Stderr, "", stdlog.LstdFlags)
