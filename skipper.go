// Copyright 2020 CleverGo. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package clevergo

import "strings"

// Skipper is a function that indicates whether current request is skippable.
type Skipper func(c *Context) bool

// PathSkipper returns a skipper with the given patterns.
// Pattern has two forms, one is that contains a certain path, another contains a wildcard,
// both of them are case-insensitive.
//   Pattern     Path            Skippable
//   ""          "/"             false
//   "/"         "/"             true
//   "/"         "/login"        false
//   "/login"    "/login"        true
//   "/login"    "/Login"        true
//   "/login"    "/LOGIN"        true
//   "/guest*"   "/guest"        true
//   "/guest*"   "/guest/foo"    true
//   "/guest*"   "/guest/bar"    true
func PathSkipper(patterns ...string) Skipper {
	return func(c *Context) bool {
		for _, pattern := range patterns {
			if pattern == "" {
				continue
			}
			if pattern[len(pattern)-1] == '*' && len(c.Request.URL.Path) >= len(pattern)-1 {
				length := len(pattern) - 1
				if strings.EqualFold(c.Request.URL.Path[:length], pattern[:length]) {
					return true
				}
			}
			if strings.EqualFold(pattern, c.Request.URL.Path) {
				return true
			}
		}
		return false
	}
}
