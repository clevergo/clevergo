// Copyright 2020 CleverGo. All rights reserved.
// Use of this source code is governed by a MIT style license that can be found
// in the LICENSE file.

package clevergo

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPathSkipper(t *testing.T) {
	cases := []struct {
		pattern string
		cases   []struct {
			target   string
			expected bool
		}
	}{
		{
			pattern: "",
			cases: []struct {
				target   string
				expected bool
			}{
				{"/", false},
				{"/login", false},
			},
		},
		{
			pattern: "/login",
			cases: []struct {
				target   string
				expected bool
			}{
				{"/", false},
				{"/login", true},
				{"/Login", true},
				{"/LOGIN", true},
			},
		},
		{
			pattern: "/guest*",
			cases: []struct {
				target   string
				expected bool
			}{
				{"/", false},
				{"/login", false},
				{"/guest", true},
				{"/Guest", true},
				{"/GUEST", true},
				{"/guest/bar", true},
				{"/guest/foo", true},
				{"/GUEST/foo", true},
			},
		},
	}
	for _, test := range cases {
		skipper := PathSkipper(test.pattern)
		for _, testCase := range test.cases {
			c := newContext(nil, httptest.NewRequest(http.MethodGet, testCase.target, nil))
			assert.Equal(t, testCase.expected, skipper(c), fmt.Sprintf("pattern: %q, target: %q", test.pattern, testCase.target))
		}
	}
}
