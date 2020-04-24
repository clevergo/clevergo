// Copyright 2020 CleverGo. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

// Package clevergo is a trie based high performance HTTP request router.
package clevergo

// Map is an alias of map[string]interface{}.
type Map map[string]interface{}

// Validatable indicates whether a value can be validated.
type Validatable interface {
	Validate() error
}
