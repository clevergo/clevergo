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

func TestNewRouteGroup(t *testing.T) {
	tests := []struct {
		path         string
		expectedPath string
		shouldPanic  bool
	}{
		{"without-prefix-slash", "", true},
		{"/", "/", false},
		{"//", "/", false},
		{"/users", "/users", false},
		{"/users/", "/users", false},
	}

	app := New()
	for _, test := range tests {
		if test.shouldPanic {
			recv := catchPanic(func() {
				newRouteGroup(app, test.path)
			})
			assert.NotNil(t, recv)
			continue
		}

		route := newRouteGroup(app, test.path)
		assert.Equal(t, test.expectedPath, route.path)
		assert.Equal(t, test.expectedPath, route.name)
	}
}

func ExampleRouteGroup() {
	app := New()
	api := app.Group("/api", RouteGroupMiddleware(echoMiddleware("api")))

	v1 := api.Group("/v1", RouteGroupMiddleware(
		echoMiddleware("v1"),
		echoMiddleware("authenticate"),
	))
	v1.Get("/users/:name", func(c *Context) error {
		c.WriteString(fmt.Sprintf("user: %s", c.Params.String("name")))
		return nil
	}, RouteMiddleware(
		echoMiddleware("fizz1"),
		echoMiddleware("fizz2"),
	))

	v2 := api.Group("/v2", RouteGroupMiddleware(
		echoMiddleware("v2"),
		echoMiddleware("authenticate"),
	))
	v2.Get("/users/:name", func(c *Context) error {
		c.WriteString(fmt.Sprintf("user: %s", c.Params.String("name")))
		return nil
	}, RouteMiddleware(
		echoMiddleware("buzz1"),
		echoMiddleware("buzz2"),
	))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/foo", nil)
	app.ServeHTTP(w, req)
	fmt.Println(w.Body.String())

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v2/users/bar", nil)
	app.ServeHTTP(w, req)
	fmt.Println(w.Body.String())

	// Output:
	// api v1 authenticate fizz1 fizz2 user: foo
	// api v2 authenticate buzz1 buzz2 user: bar
}

func TestRouteGroupName(t *testing.T) {
	for _, name := range []string{"foo", "bar"} {
		g := &RouteGroup{}
		RouteGroupName(name)(g)
		assert.Equal(t, name, g.name)
	}
}
