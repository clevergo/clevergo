// Copyright 2016 The Gem Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license
// that can be found in the LICENSE file.

package gem

import (
	"fmt"
	"strconv"
	"testing"
)

type testString struct {
	v    interface{}
	want string
	err  error
}

var (
	testStrings = []testString{
		testString{"foo", "foo", nil},
		testString{1, "1", nil},
		testString{[]byte("bar"), "bar", nil},
		testString{false, "", fmt.Errorf("unsupport to convert type %T to string", false)},
	}
)

func TestString(t *testing.T) {
	var got string
	var err error
	for _, str := range testStrings {
		got, err = String(str.v)
		if str.want != got {
			t.Errorf("expected %q, got %q", str.want, got)
		}

		if fmt.Sprintf("%s", str.err) != fmt.Sprintf("%s", err) {
			t.Errorf("expected %q, got %q", str.err, err)
		}
	}
}

type testInt struct {
	v    interface{}
	want int
	err  error
}

var (
	_, errEmptyStr2Int = strconv.Atoi("")

	testInts = []testInt{
		testInt{1, 1, nil},
		testInt{"2", 2, nil},
		testInt{"", 0, errEmptyStr2Int},
		testInt{false, 0, fmt.Errorf("unsupport to convert type %T to int", false)},
	}
)

func TestInt(t *testing.T) {
	var got int
	var err error
	for _, str := range testInts {
		got, err = Int(str.v)
		if str.want != got {
			t.Errorf("expected %q, got %q", str.want, got)
		}

		if fmt.Sprintf("%s", str.err) != fmt.Sprintf("%s", err) {
			t.Errorf("expected %q, got %q", str.err, err)
		}
	}
}
