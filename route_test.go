// Copyright 2020 CleverGo. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package clevergo

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRouteGroupURL(t *testing.T) {
	router := NewRouter()
	router.Handle(http.MethodGet, "/", echoHandler(""), RouteName("home"))
	router.Handle(http.MethodGet, "/users/:id", echoHandler(""), RouteName("user"))
	router.Handle(http.MethodGet, "/posts/:year/:month/:title", echoHandler(""), RouteName("post"))
	router.Handle(http.MethodGet, "/static/*filepath", echoHandler(""), RouteName("static"))

	var tests = []struct {
		name        string
		args        []string
		exepctedURL string
		shouldError bool
	}{
		{"home", nil, "/", false},
		{"home", []string{"keyWithoutValue"}, "", true},

		{"user", nil, "", true},
		{"user", nil, "", true},
		{"user", []string{"id", "foo"}, "/users/foo", false},
		{"user", []string{"id", "bar"}, "/users/bar", false},

		{"post", nil, "", true},
		{"post", []string{"year", "2020"}, "", true},
		{"post", []string{"month", "01"}, "", true},
		{"post", []string{"title", "foo"}, "", true},
		{"post", []string{"year", "2020", "month", "01"}, "/posts/2020/01/foo", true},
		{"post", []string{"month", "01", "title", "foo"}, "/posts/2020/01/foo", true},
		{"post", []string{"year", "2020", "title", "foo"}, "/posts/2020/01/foo", true},
		{"post", []string{"year", "2020", "month", "01", "title", "foo"}, "/posts/2020/01/foo", false},
		{"post", []string{"year", "2020", "month", "02", "title", "bar"}, "/posts/2020/02/bar", false},

		{"static", nil, "/static/", false},
		{"static", []string{"filepath", "js/app.js"}, "/static/js/app.js", false},
		{"static", []string{"filepath", "css/app.css"}, "/static/css/app.css", false},
	}
	for _, test := range tests {
		url, err := router.URL(test.name, test.args...)
		if test.shouldError {
			assert.NotNil(t, err)
			continue
		}
		assert.Nil(t, err)
		assert.Equal(t, test.exepctedURL, url.String())
	}

}

func TestRouteGroupAPI(t *testing.T) {
	var get, head, options, post, put, patch, delete, handler, handlerFunc bool

	httpHandler := handlerStruct{&handler}

	router := NewRouter()
	api := router.Group("/api")
	api.Get("/GET", func(ctx *Context) error {
		get = true
		return nil
	})
	api.Head("/GET", func(ctx *Context) error {
		head = true
		return nil
	})
	api.Options("/GET", func(ctx *Context) error {
		options = true
		return nil
	})
	api.Post("/POST", func(ctx *Context) error {
		post = true
		return nil
	})
	api.Put("/PUT", func(ctx *Context) error {
		put = true
		return nil
	})
	api.Patch("/PATCH", func(ctx *Context) error {
		patch = true
		return nil
	})
	api.Delete("/DELETE", func(ctx *Context) error {
		delete = true
		return nil
	})
	api.Handler(http.MethodGet, "/Handler", httpHandler)
	api.HandlerFunc(http.MethodGet, "/HandlerFunc", func(w http.ResponseWriter, r *http.Request) {
		handlerFunc = true
	})

	w := new(mockResponseWriter)

	r, _ := http.NewRequest(http.MethodGet, "/api/GET", nil)
	router.ServeHTTP(w, r)
	assert.True(t, get, "routing GET failed")

	r, _ = http.NewRequest(http.MethodHead, "/api/GET", nil)
	router.ServeHTTP(w, r)
	assert.True(t, head, "routing HEAD failed")

	r, _ = http.NewRequest(http.MethodOptions, "/api/GET", nil)
	router.ServeHTTP(w, r)
	assert.True(t, options, "routing GEOPTIONST failed")

	r, _ = http.NewRequest(http.MethodPost, "/api/POST", nil)
	router.ServeHTTP(w, r)
	assert.True(t, post, "routing POST failed")

	r, _ = http.NewRequest(http.MethodPut, "/api/PUT", nil)
	router.ServeHTTP(w, r)
	assert.True(t, put, "routing PUT failed")

	r, _ = http.NewRequest(http.MethodPatch, "/api/PATCH", nil)
	router.ServeHTTP(w, r)
	assert.True(t, patch, "routing PATCH failed")

	r, _ = http.NewRequest(http.MethodDelete, "/api/DELETE", nil)
	router.ServeHTTP(w, r)
	assert.True(t, delete, "routing DELETE failed")

	r, _ = http.NewRequest(http.MethodGet, "/api/Handler", nil)
	router.ServeHTTP(w, r)
	assert.True(t, handler, "routing Handler failed")

	r, _ = http.NewRequest(http.MethodGet, "/api/HandlerFunc", nil)
	router.ServeHTTP(w, r)
	assert.True(t, handlerFunc, "routing HandlerFunc failed")
}

func TestRouteMiddleware(t *testing.T) {
	m1 := echoMiddleware("m1")
	m2 := echoMiddleware("m2")
	handler := echoHandler("hello")

	router := NewRouter()
	router.Handle(http.MethodGet, "/", handler)
	router.Handle(http.MethodGet, "/middleware", handler, RouteMiddleware(m1, m2))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, "hello", w.Body.String())

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/middleware", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, "m1 m2 hello", w.Body.String())
}

func TestNestedRouteGroup(t *testing.T) {
	m1 := echoMiddleware("m1")
	m2 := echoMiddleware("m2")
	handler := echoHandler("hello")

	router := NewRouter()
	api := router.Group("/api")
	v1 := api.Group("/v1", RouteGroupMiddleware(m1))
	v2 := api.Group("/v2", RouteGroupMiddleware(m2))

	v1.Handle(http.MethodGet, "/", handler, RouteName("home"))
	v2.Handle(http.MethodGet, "/", handler, RouteName("home"))

	url, err := router.URL("/api/v1/home")
	assert.Nil(t, err)
	assert.Equal(t, "/api/v1/", url.String())

	url, err = router.URL("/api/v2/home")
	assert.Nil(t, err)
	assert.Equal(t, "/api/v2/", url.String())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, "m1 hello", w.Body.String())

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v2/", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, "m2 hello", w.Body.String())
}

func ExampleRoute() {
	router := NewRouter()
	router.Get("/posts/:page", func(ctx *Context) error {
		page, _ := ctx.Params.Int("page")
		route := ctx.Route
		prev, _ := route.URL("page", strconv.Itoa(page-1))
		next, _ := route.URL("page", strconv.Itoa(page+1))
		fmt.Printf("prev page url: %s\n", prev)
		fmt.Printf("next page url: %s\n", next)
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/posts/3", nil)
	router.ServeHTTP(nil, req)

	// Output:
	// prev page url: /posts/2
	// next page url: /posts/4
}
