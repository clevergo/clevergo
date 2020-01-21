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
			if err == nil {
				t.Error("expected an error, got nil")
			}
			continue
		}

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if url.String() != test.exepctedURL {
			t.Errorf("expected url %q, got %q", test.exepctedURL, url)
		}
	}

}

func TestRouteGroupAPI(t *testing.T) {
	var get, head, options, post, put, patch, delete, handler, handlerFunc bool

	httpHandler := handlerStruct{&handler}

	router := NewRouter()
	api := router.Group("/api")
	api.Get("/GET", func(w http.ResponseWriter, r *http.Request) {
		get = true
	})
	api.Head("/GET", func(w http.ResponseWriter, r *http.Request) {
		head = true
	})
	api.Options("/GET", func(w http.ResponseWriter, r *http.Request) {
		options = true
	})
	api.Post("/POST", func(w http.ResponseWriter, r *http.Request) {
		post = true
	})
	api.Put("/PUT", func(w http.ResponseWriter, r *http.Request) {
		put = true
	})
	api.Patch("/PATCH", func(w http.ResponseWriter, r *http.Request) {
		patch = true
	})
	api.Delete("/DELETE", func(w http.ResponseWriter, r *http.Request) {
		delete = true
	})
	api.Handle(http.MethodGet, "/Handler", httpHandler)
	api.HandleFunc(http.MethodGet, "/HandlerFunc", func(w http.ResponseWriter, r *http.Request) {
		handlerFunc = true
	})

	w := new(mockResponseWriter)

	r, _ := http.NewRequest(http.MethodGet, "/api/GET", nil)
	router.ServeHTTP(w, r)
	if !get {
		t.Error("routing GET failed")
	}

	r, _ = http.NewRequest(http.MethodHead, "/api/GET", nil)
	router.ServeHTTP(w, r)
	if !head {
		t.Error("routing HEAD failed")
	}

	r, _ = http.NewRequest(http.MethodOptions, "/api/GET", nil)
	router.ServeHTTP(w, r)
	if !options {
		t.Error("routing OPTIONS failed")
	}

	r, _ = http.NewRequest(http.MethodPost, "/api/POST", nil)
	router.ServeHTTP(w, r)
	if !post {
		t.Error("routing POST failed")
	}

	r, _ = http.NewRequest(http.MethodPut, "/api/PUT", nil)
	router.ServeHTTP(w, r)
	if !put {
		t.Error("routing PUT failed")
	}

	r, _ = http.NewRequest(http.MethodPatch, "/api/PATCH", nil)
	router.ServeHTTP(w, r)
	if !patch {
		t.Error("routing PATCH failed")
	}

	r, _ = http.NewRequest(http.MethodDelete, "/api/DELETE", nil)
	router.ServeHTTP(w, r)
	if !delete {
		t.Error("routing DELETE failed")
	}

	r, _ = http.NewRequest(http.MethodGet, "/api/Handler", nil)
	router.ServeHTTP(w, r)
	if !handler {
		t.Error("routing Handler failed")
	}

	r, _ = http.NewRequest(http.MethodGet, "/api/HandlerFunc", nil)
	router.ServeHTTP(w, r)
	if !handlerFunc {
		t.Error("routing HandlerFunc failed")
	}
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
	if w.Body.String() != "hello" {
		t.Errorf("expected body %q, got %q", "hello", w.Body)
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/middleware", nil)
	router.ServeHTTP(w, req)
	if w.Body.String() != "m1 m2 hello" {
		t.Errorf("expected body %q, got %q", "m1 m2 hello", w.Body)
	}
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
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if url.String() != "/api/v1/" {
		t.Errorf("expected url %q got %q", "/api/v1/", url)
	}

	url, err = router.URL("/api/v2/home")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if url.String() != "/api/v2/" {
		t.Errorf("expected url %q got %q", "/api/v2/", url)
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/", nil)
	router.ServeHTTP(w, req)
	if w.Body.String() != "m1 hello" {
		t.Errorf("expected body %q, got %q", "m1 hello", w.Body)
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v2/", nil)
	router.ServeHTTP(w, req)
	if w.Body.String() != "m2 hello" {
		t.Errorf("expected body %q, got %q", "m2 hello", w.Body)
	}
}

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

	router := NewRouter()
	for _, test := range tests {
		if test.shouldPanic {
			recv := catchPanic(func() {
				newRouteGroup(router, test.path)
			})
			if recv == nil {
				t.Error("expected a panic")
			}
			continue
		}

		route := newRouteGroup(router, test.path)
		if test.expectedPath != route.path {
			t.Errorf("expected path %q, got %q", test.expectedPath, route.path)
		}
	}
}

func ExampleRouteGroup() {
	router := NewRouter()
	api := router.Group("/api")

	v1 := api.Group("/v1")
	v1.Get("/users/:name", func(w http.ResponseWriter, r *http.Request) {
		params := GetParams(r)
		fmt.Printf("v1 user: %s\n", params.Get("name"))
	})

	v2 := api.Group("/v2")
	v2.Get("/users/:name", func(w http.ResponseWriter, r *http.Request) {
		params := GetParams(r)
		fmt.Printf("v2 user: %s\n", params.Get("name"))
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/foo", nil)
	router.ServeHTTP(nil, req)

	req = httptest.NewRequest(http.MethodGet, "/api/v2/users/bar", nil)
	router.ServeHTTP(nil, req)

	// Output:
	// v1 user: foo
	// v2 user: bar
}

func ExampleGetRoute() {
	router := NewRouter()
	router.Get("/posts/:page", func(_ http.ResponseWriter, r *http.Request) {
		ps := GetParams(r)
		page, _ := strconv.Atoi(ps.Get("page"))
		route := GetRoute(r)
		prev, _ := route.URL("page", strconv.Itoa(page-1))
		next, _ := route.URL("page", strconv.Itoa(page+1))
		fmt.Printf("prev page url: %s\n", prev)
		fmt.Printf("next page url: %s\n", next)
	})

	req := httptest.NewRequest(http.MethodGet, "/posts/3", nil)
	router.ServeHTTP(nil, req)

	// Output:
	// prev page url: /posts/2
	// next page url: /posts/4
}
