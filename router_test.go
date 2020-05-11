// Copyright 2013 Julien Schmidt. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package clevergo

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockResponseWriter struct{}

func (m *mockResponseWriter) Header() (h http.Header) {
	return http.Header{}
}

func (m *mockResponseWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (m *mockResponseWriter) WriteString(s string) (n int, err error) {
	return len(s), nil
}

func (m *mockResponseWriter) WriteHeader(int) {}

func TestRouter(t *testing.T) {
	router := NewRouter()

	routed := false
	router.Handle(http.MethodGet, "/user/:name", func(ctx *Context) error {
		routed = true
		expected := Params{Param{"name", "gopher"}}
		assert.Equal(t, expected, ctx.Params)
		return nil
	})

	w := new(mockResponseWriter)

	req, _ := http.NewRequest(http.MethodGet, "/user/gopher", nil)
	router.ServeHTTP(w, req)
	assert.True(t, routed)
}

type handlerStruct struct {
	handled *bool
}

func (h handlerStruct) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	*h.handled = true
}

func TestRouterAPI(t *testing.T) {
	var get, head, options, post, put, patch, delete, handler, handlerFunc bool

	httpHandler := handlerStruct{&handler}

	router := NewRouter()
	router.Get("/GET", func(ctx *Context) error {
		get = true
		return nil
	})
	router.Head("/GET", func(ctx *Context) error {
		head = true
		return nil
	})
	router.Options("/GET", func(ctx *Context) error {
		options = true
		return nil
	})
	router.Post("/POST", func(ctx *Context) error {
		post = true
		return nil
	})
	router.Put("/PUT", func(ctx *Context) error {
		put = true
		return nil
	})
	router.Patch("/PATCH", func(ctx *Context) error {
		patch = true
		return nil
	})
	router.Delete("/DELETE", func(ctx *Context) error {
		delete = true
		return nil
	})
	router.Handler(http.MethodGet, "/Handler", httpHandler)
	router.HandlerFunc(http.MethodGet, "/HandlerFunc", func(w http.ResponseWriter, r *http.Request) {
		handlerFunc = true
	})

	w := new(mockResponseWriter)

	r, _ := http.NewRequest(http.MethodGet, "/GET", nil)
	router.ServeHTTP(w, r)
	assert.True(t, get, "routing GET failed")

	r, _ = http.NewRequest(http.MethodHead, "/GET", nil)
	router.ServeHTTP(w, r)
	assert.True(t, head, "routing HEAD failed")

	r, _ = http.NewRequest(http.MethodOptions, "/GET", nil)
	router.ServeHTTP(w, r)
	assert.True(t, options, "routing OPTIONS failed")

	r, _ = http.NewRequest(http.MethodPost, "/POST", nil)
	router.ServeHTTP(w, r)
	assert.True(t, post, "routing POST failed")

	r, _ = http.NewRequest(http.MethodPut, "/PUT", nil)
	router.ServeHTTP(w, r)
	assert.True(t, put, "routing PUT failed")

	r, _ = http.NewRequest(http.MethodPatch, "/PATCH", nil)
	router.ServeHTTP(w, r)
	assert.True(t, patch, "routing PATCH failed")

	r, _ = http.NewRequest(http.MethodDelete, "/DELETE", nil)
	router.ServeHTTP(w, r)
	assert.True(t, delete, "routing DELETE failed")

	r, _ = http.NewRequest(http.MethodGet, "/Handler", nil)
	router.ServeHTTP(w, r)
	assert.True(t, handler, "routing Handler failed")

	r, _ = http.NewRequest(http.MethodGet, "/HandlerFunc", nil)
	router.ServeHTTP(w, r)
	assert.True(t, handlerFunc, "routing HandlerFunc failed")
}

func TestRouterAny(t *testing.T) {
	router := NewRouter()
	handle := func(ctx *Context) error {
		ctx.WriteString(ctx.Request.Method)
		return nil
	}
	nameOpt := RouteName("ping")
	router.Any("/ping", handle, nameOpt)
	group := router.Group("/foo")
	group.Any("/ping", handle, nameOpt)
	paths := []string{"/ping", "/foo/ping"}
	for _, method := range requestMethods {
		for _, path := range paths {
			w := httptest.NewRecorder()
			router.ServeHTTP(w, httptest.NewRequest(method, path, nil))
			assert.Equal(t, method, w.Body.String())
		}
	}
	url, err := router.URL("ping")
	assert.Nil(t, err)
	assert.Equal(t, "/ping", url.String())
}

func TestRouterInvalidInput(t *testing.T) {
	router := NewRouter()

	handle := func(ctx *Context) error {
		return nil
	}

	recv := catchPanic(func() {
		router.Handle("", "/", handle)
	})
	assert.NotNil(t, recv, "registering empty method did not panic")

	recv = catchPanic(func() {
		router.Get("", handle)
	})
	assert.NotNil(t, recv, "registering empty path did not panic")

	recv = catchPanic(func() {
		router.Get("noSlashRoot", handle)
	})
	assert.NotNil(t, recv, "registering path not beginning with '/' did not panic")

	recv = catchPanic(func() {
		router.Get("/", nil)
	})
	assert.NotNil(t, recv, "registering nil handler did not panic")
}

func TestRouterChaining(t *testing.T) {
	router1 := NewRouter()
	router2 := NewRouter()
	router1.NotFound = router2

	fooHit := false
	router1.Post("/foo", func(ctx *Context) error {
		fooHit = true
		ctx.Response.WriteHeader(http.StatusOK)
		return nil
	})

	barHit := false
	router2.Post("/bar", func(ctx *Context) error {
		barHit = true
		ctx.Response.WriteHeader(http.StatusOK)
		return nil
	})

	r, _ := http.NewRequest(http.MethodPost, "/foo", nil)
	w := httptest.NewRecorder()
	router1.ServeHTTP(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, fooHit)

	r, _ = http.NewRequest(http.MethodPost, "/bar", nil)
	w = httptest.NewRecorder()
	router1.ServeHTTP(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, barHit)

	r, _ = http.NewRequest(http.MethodPost, "/qax", nil)
	w = httptest.NewRecorder()
	router1.ServeHTTP(w, r)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func BenchmarkAllowed(b *testing.B) {
	handlerFunc := func(ctx *Context) error {
		return nil
	}

	router := NewRouter()
	router.Post("/path", handlerFunc)
	router.Get("/path", handlerFunc)

	b.Run("Global", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = router.allowed("*", http.MethodOptions)
		}
	})
	b.Run("Path", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = router.allowed("/path", http.MethodOptions)
		}
	})
}

func TestRouterOPTIONS(t *testing.T) {
	handlerFunc := func(ctx *Context) error {
		return nil
	}

	router := NewRouter()
	router.Post("/path", handlerFunc)

	// test not allowed
	// * (server)
	r, _ := http.NewRequest(http.MethodOptions, "*", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "OPTIONS, POST", w.Header().Get("Allow"))

	// path
	r, _ = http.NewRequest(http.MethodOptions, "/path", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "OPTIONS, POST", w.Header().Get("Allow"))

	r, _ = http.NewRequest(http.MethodOptions, "/doesnotexist", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	assert.Equal(t, http.StatusNotFound, w.Code)

	// add another method
	router.Get("/path", handlerFunc)

	// set a global OPTIONS handler
	router.GlobalOPTIONS = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Adjust status code to 204
		w.WriteHeader(http.StatusNoContent)
	})

	// test again
	// * (server)
	r, _ = http.NewRequest(http.MethodOptions, "*", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "GET, OPTIONS, POST", w.Header().Get("Allow"))

	// path
	r, _ = http.NewRequest(http.MethodOptions, "/path", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "GET, OPTIONS, POST", w.Header().Get("Allow"))

	// custom handler
	var custom bool
	router.Options("/path", func(ctx *Context) error {
		custom = true
		return nil
	})

	// test again
	// * (server)
	r, _ = http.NewRequest(http.MethodOptions, "*", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "GET, OPTIONS, POST", w.Header().Get("Allow"))
	assert.False(t, custom, "custom handler called on *")

	// path
	r, _ = http.NewRequest(http.MethodOptions, "/path", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, custom, "custom handler not called*")
}

func TestRouterNotAllowed(t *testing.T) {
	handlerFunc := func(ctx *Context) error {
		return nil
	}

	router := NewRouter()
	router.Post("/path", handlerFunc)

	// test not allowed
	r, _ := http.NewRequest(http.MethodGet, "/path", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	assert.Equal(t, "OPTIONS, POST", w.Header().Get("Allow"))

	// add another method
	router.Delete("/path", handlerFunc)
	router.Options("/path", handlerFunc) // must be ignored

	// test again
	r, _ = http.NewRequest(http.MethodGet, "/path", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	assert.Equal(t, "DELETE, OPTIONS, POST", w.Header().Get("Allow"))

	// test custom handler
	w = httptest.NewRecorder()
	responseText := "custom method"
	router.MethodNotAllowed = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		w.Write([]byte(responseText))
	})
	router.ServeHTTP(w, r)
	assert.Equal(t, responseText, w.Body.String())
	assert.Equal(t, http.StatusTeapot, w.Code)
	assert.Equal(t, "DELETE, OPTIONS, POST", w.Header().Get("Allow"))
}

func TestRouterNotFound(t *testing.T) {
	handlerFunc := func(ctx *Context) error {
		return nil
	}

	router := NewRouter()
	router.Get("/path", handlerFunc)
	router.Get("/dir/", handlerFunc)
	router.Get("/", handlerFunc)

	testRoutes := []struct {
		route    string
		code     int
		location string
	}{
		{"/path/", http.StatusMovedPermanently, "/path"},   // TSR -/
		{"/dir", http.StatusMovedPermanently, "/dir/"},     // TSR +/
		{"", http.StatusMovedPermanently, "/"},             // TSR +/
		{"/PATH", http.StatusMovedPermanently, "/path"},    // Fixed Case
		{"/DIR/", http.StatusMovedPermanently, "/dir/"},    // Fixed Case
		{"/PATH/", http.StatusMovedPermanently, "/path"},   // Fixed Case -/
		{"/DIR", http.StatusMovedPermanently, "/dir/"},     // Fixed Case +/
		{"/../path", http.StatusMovedPermanently, "/path"}, // CleanPath
		{"/nope", http.StatusNotFound, ""},                 // NotFound
	}
	for _, tr := range testRoutes {
		r, _ := http.NewRequest(http.MethodGet, tr.route, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)
		assert.Equal(t, tr.code, w.Code)
		if w.Code != http.StatusNotFound {
			assert.Equal(t, tr.location, fmt.Sprint(w.Header().Get("Location")))
		}
	}

	// Test custom not found handler
	var notFound bool
	router.NotFound = http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusNotFound)
		notFound = true
	})
	r, _ := http.NewRequest(http.MethodGet, "/nope", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)
	assert.True(t, notFound)
	assert.Equal(t, http.StatusNotFound, w.Code)

	// Test other method than GET (want 308 instead of 301)
	router.Patch("/path", handlerFunc)
	r, _ = http.NewRequest(http.MethodPatch, "/path/", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	assert.Equal(t, http.StatusPermanentRedirect, w.Code)
	assert.Equal(t, "map[Location:[/path]]", fmt.Sprint(w.Header()))

	// Test special case where no node for the prefix "/" exists
	router = NewRouter()
	router.Get("/a", handlerFunc)
	r, _ = http.NewRequest(http.MethodGet, "/", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestRouterLookup(t *testing.T) {
	routed := false
	wantHandle := func(ctx *Context) error {
		routed = true
		return nil
	}
	wantParams := Params{Param{"name", "gopher"}}

	router := NewRouter()

	// try empty router first
	route, _, tsr := router.Lookup(http.MethodGet, "/nope")
	assert.Nil(t, route, "Got route for unregistered pattern: %v", route)
	assert.False(t, tsr, "Got wrong TSR recommendation!")

	// insert route and try again
	router.Get("/user/:name", wantHandle)
	route, params, _ := router.Lookup(http.MethodGet, "/user/gopher")
	assert.NotNil(t, route, "Got no route!")
	route.handle(newContext(nil, nil))
	assert.True(t, routed)
	assert.Equal(t, wantParams, params)

	routed = false

	// route without param
	router.Get("/user", wantHandle)
	route, params, _ = router.Lookup(http.MethodGet, "/user")
	assert.NotNil(t, route, "Got no route!")
	route.handle(newContext(nil, nil))
	assert.True(t, routed)
	assert.Len(t, params, 0)

	route, _, tsr = router.Lookup(http.MethodGet, "/user/gopher/")
	assert.Nil(t, route, "Got route for unregistered pattern: %v", route)
	assert.True(t, tsr, "Got no TSR recommendation!")

	route, _, tsr = router.Lookup(http.MethodGet, "/nope")
	assert.Nilf(t, route, "Got route for unregistered pattern: %v", route)
	assert.False(t, tsr, "Got wrong TSR recommendation!")
}

func TestRouterParamsFromContext(t *testing.T) {
	routed := false

	wantParams := Params{Param{"name", "gopher"}}
	handlerFunc := func(ctx *Context) error {
		assert.Equal(t, wantParams, ctx.Params)
		routed = true
		return nil
	}

	handlerFuncNil := func(ctx *Context) error {
		assert.Len(t, ctx.Params, 0)
		routed = true
		return nil
	}
	router := NewRouter()
	router.Handle(http.MethodGet, "/user", handlerFuncNil)
	router.Handle(http.MethodGet, "/user/:name", handlerFunc)

	w := new(mockResponseWriter)
	r, _ := http.NewRequest(http.MethodGet, "/user/gopher", nil)
	router.ServeHTTP(w, r)
	assert.True(t, routed)

	routed = false
	r, _ = http.NewRequest(http.MethodGet, "/user", nil)
	router.ServeHTTP(w, r)
	assert.True(t, routed)
}

func TestRouterMatchedRoutePath(t *testing.T) {
	route1 := "/user/:name"
	routed1 := false
	handle1 := func(ctx *Context) error {
		assert.Equal(t, route1, ctx.Route.path)
		routed1 = true
		return nil
	}

	route2 := "/user/:name/details"
	routed2 := false
	handle2 := func(ctx *Context) error {
		assert.Equal(t, route2, ctx.Route.path)
		routed2 = true
		return nil
	}

	route3 := "/"
	routed3 := false
	handle3 := func(ctx *Context) error {
		assert.Equal(t, route3, ctx.Route.path)
		routed3 = true
		return nil
	}

	router := NewRouter()
	router.Handle(http.MethodGet, route1, handle1)
	router.Handle(http.MethodGet, route2, handle2)
	router.Handle(http.MethodGet, route3, handle3)

	w := new(mockResponseWriter)
	r, _ := http.NewRequest(http.MethodGet, "/user/gopher", nil)
	router.ServeHTTP(w, r)
	assert.True(t, routed1)

	w = new(mockResponseWriter)
	r, _ = http.NewRequest(http.MethodGet, "/user/gopher/details", nil)
	router.ServeHTTP(w, r)
	assert.True(t, routed2)

	w = new(mockResponseWriter)
	r, _ = http.NewRequest(http.MethodGet, "/", nil)
	router.ServeHTTP(w, r)
	assert.True(t, routed3)
}

type mockFileSystem struct {
	opened bool
}

func (mfs *mockFileSystem) Open(name string) (http.File, error) {
	mfs.opened = true
	return nil, errors.New("this is just a mock")
}

func TestRouterServeFiles(t *testing.T) {
	router := NewRouter()
	mfs := &mockFileSystem{}

	recv := catchPanic(func() {
		router.ServeFiles("/noFilepath", mfs)
	})
	assert.NotNil(t, recv, "registering path not ending with '*filepath' did not panic")

	router.ServeFiles("/*filepath", mfs)
	w := new(mockResponseWriter)
	r, _ := http.NewRequest(http.MethodGet, "/favicon.ico", nil)
	router.ServeHTTP(w, r)
	assert.True(t, mfs.opened, "serving file failed")
}

func TestRouterNamedRoute(t *testing.T) {
	tests := []struct {
		path        string
		name        string
		args        []string
		expectedURL string
	}{
		{"/", "home", nil, "/"},
		{"/users/:id", "user", []string{"id", "foo"}, "/users/foo"},
	}
	router := NewRouter()
	for _, test := range tests {
		router.Handle(http.MethodGet, test.path, func(ctx *Context) error {

			return nil
		}, RouteName(test.name))
		url, err := router.URL(test.name, test.args...)
		assert.Nil(t, err)
		assert.Equal(t, test.expectedURL, url.String())
	}

	// unregistered name.
	_, err := router.URL("unregistered")
	assert.NotNil(t, err)

	// registers same route name.
	recv := catchPanic(func() {
		router.Handle(http.MethodGet, "/same", func(ctx *Context) error {
			return nil
		}, RouteName("home"))
	})
	assert.NotNil(t, recv)
}

func ExampleRouter_URL() {
	router := NewRouter()
	router.Get("/hello/:name", func(ctx *Context) error {
		return nil
	}, RouteName("hello"))
	// nested routes group
	api := router.Group("/api")

	v1 := api.Group("/v1")
	// the group path will become the prefix of route name.
	v1.Get("/users/:name", func(ctx *Context) error {
		return nil
	}, RouteName("user"))

	// specified the name of the route group.
	v2 := api.Group("/v2", RouteGroupName("/apiV2"))
	v2.Get("/users/:name", func(ctx *Context) error {
		return nil
	}, RouteName("user"))

	routes := []struct {
		name string
		args []string
	}{
		{"hello", []string{"name", "foo"}},
		{"hello", []string{"name", "bar"}},
		{"/api/v1/user", []string{"name", "foo"}},
		{"/api/v1/user", []string{"name", "bar"}},
		{"/apiV2/user", []string{"name", "foo"}},
		{"/apiV2/user", []string{"name", "bar"}},
	}

	for _, route := range routes {
		url, _ := router.URL(route.name, route.args...)
		fmt.Println(url)
	}

	// Output:
	// /hello/foo
	// /hello/bar
	// /api/v1/users/foo
	// /api/v1/users/bar
	// /api/v2/users/foo
	// /api/v2/users/bar
}

func ExampleRouter_ServeFiles() {
	router := NewRouter()

	router.ServeFiles("/static/*filepath", http.Dir("/path/to/static"))

	// sometimes, it is useful to treat http.FileServer as NotFoundHandler,
	// such as "/favicon.ico".
	router.NotFound = http.FileServer(http.Dir("public"))
}

type testErrorHandler struct {
	status int
}

func (eh testErrorHandler) Handle(ctx *Context, err error) {
	ctx.Error(eh.status, err.Error())
}

func TestRouter_ErrorHandler(t *testing.T) {
	router := NewRouter()
	router.ErrorHandler = &testErrorHandler{http.StatusInternalServerError}
	router.Get("/error/:msg", func(ctx *Context) error {
		return errors.New(ctx.Params.String("msg"))
	})

	msgs := []string{"foo", "bar"}
	for _, msg := range msgs {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/error/"+msg, nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, fmt.Sprintln(msg), w.Body.String())
	}
}

func TestRouter_HandleError(t *testing.T) {
	router := NewRouter()
	tests := []struct {
		err  error
		body string
		code int
	}{
		{errors.New("foo"), "foo", http.StatusInternalServerError},
		{ErrNotFound, ErrNotFound.Error(), ErrNotFound.Code},
		{ErrMethodNotAllowed, ErrMethodNotAllowed.Error(), ErrMethodNotAllowed.Code},
	}
	for _, test := range tests {
		w := httptest.NewRecorder()
		ctx := newContext(w, nil)
		router.HandleError(ctx, test.err)
		assert.Equal(t, test.code, w.Code)
		assert.Equal(t, fmt.Sprintln(test.body), w.Body.String())
	}
}

func TestRouterUseRawPath(t *testing.T) {
	router := NewRouter()
	router.UseRawPath = true
	handled := false
	handle := func(ctx *Context) error {
		expected := Params{Param{"name", "foo/bar"}}
		assert.Equal(t, expected, ctx.Params)
		handled = true
		return nil
	}
	router.Get("/hello/:name", handle)
	req := httptest.NewRequest(http.MethodGet, "/hello/foo%2fbar", nil)
	router.ServeHTTP(nil, req)
	assert.True(t, handled, "raw path routing failed")
}

func TestRouterUseRawPathMixed(t *testing.T) {
	router := NewRouter()
	router.UseRawPath = true
	handled := false
	handle := func(ctx *Context) error {
		expected := Params{Param{"date", "2020/03/23"}, Param{"slug", "hello world"}}
		assert.Equal(t, expected, ctx.Params)
		handled = true
		return nil
	}
	router.Get("/post/:date/:slug", handle)
	req := httptest.NewRequest(http.MethodGet, "/post/2020%2f03%2f23/hello%20world", nil)
	router.ServeHTTP(nil, req)
	assert.True(t, handled, "raw path routing failed")
}

func TestRouterUseRawPathCatchAll(t *testing.T) {
	router := NewRouter()
	router.UseRawPath = true
	handled := false
	handle := func(ctx *Context) error {
		expected := Params{Param{"slug", "/2020/03/23-hello world"}}
		assert.Equal(t, expected, ctx.Params)
		handled = true
		return nil
	}
	router.Get("/post/*slug", handle)
	req := httptest.NewRequest(http.MethodGet, "/post/2020%2f03%2f23-hello%20world", nil)
	router.ServeHTTP(nil, req)
	assert.True(t, handled, "raw path routing failed")
}

func TestRouterUse(t *testing.T) {
	router := NewRouter()
	router.Use(
		echoMiddleware("m1"),
		echoMiddleware("m2"),
	)
	router.Get("/", echoHandler("foobar"))
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, "m1 m2 foobar", w.Body.String())

	router.Use(terminatedMiddleware())
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, "m1 m2 terminated", w.Body.String())
}
