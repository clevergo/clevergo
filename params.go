// Copyright 2013 Julien Schmidt. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// at https://github.com/julienschmidt/httprouter/blob/master/LICENSE.

package clevergo

import "strconv"

// Param is a single URL parameter, consisting of a key and a value.
type Param struct {
	Key   string
	Value string
}

// Params is a Param-slice, as returned by the router.
// The slice is ordered, the first URL parameter is also the first slice value.
// It is therefore safe to read values by the index.
type Params []Param

// String returns the value of the first Param which key matches the given name.
// If no matching Param is found, an empty string is returned.
func (ps Params) String(name string) string {
	for _, p := range ps {
		if p.Key == name {
			return p.Value
		}
	}
	return ""
}

// Bool returns the boolean value of the given name.
func (ps Params) Bool(name string) (bool, error) {
	return strconv.ParseBool(ps.String(name))
}

// Float64 returns the float64 value of the given name.
func (ps Params) Float64(name string) (float64, error) {
	return strconv.ParseFloat(ps.String(name), 64)
}

// Int returns the int value of the given name.
func (ps Params) Int(name string) (int, error) {
	return strconv.Atoi(ps.String(name))
}

// Int64 returns the int64 value of the given name.
func (ps Params) Int64(name string) (int64, error) {
	return strconv.ParseInt(ps.String(name), 10, 64)
}

// Uint64 returns the uint64 value of the given name.
func (ps Params) Uint64(name string) (uint64, error) {
	return strconv.ParseUint(ps.String(name), 10, 64)
}
