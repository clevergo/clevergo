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

func TestApplication(t *testing.T) {
	app := New()

	routed := false
	app.Handle(http.MethodGet, "/user/:name", func(ctx *Context) error {
		routed = true
		expected := Params{Param{"name", "gopher"}}
		assert.Equal(t, expected, ctx.Params)
		return nil
	})

	w := new(mockResponseWriter)

	req, _ := http.NewRequest(http.MethodGet, "/user/gopher", nil)
	app.ServeHTTP(w, req)
	assert.True(t, routed)
}

type handlerStruct struct {
	handled *bool
}

func (h handlerStruct) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	*h.handled = true
}

func TestApplicationAPI(t *testing.T) {
	var get, head, options, post, put, patch, delete, handler, handlerFunc bool

	httpHandler := handlerStruct{&handler}

	app := New()
	app.Get("/GET", func(ctx *Context) error {
		get = true
		return nil
	})
	app.Head("/GET", func(ctx *Context) error {
		head = true
		return nil
	})
	app.Options("/GET", func(ctx *Context) error {
		options = true
		return nil
	})
	app.Post("/POST", func(ctx *Context) error {
		post = true
		return nil
	})
	app.Put("/PUT", func(ctx *Context) error {
		put = true
		return nil
	})
	app.Patch("/PATCH", func(ctx *Context) error {
		patch = true
		return nil
	})
	app.Delete("/DELETE", func(ctx *Context) error {
		delete = true
		return nil
	})
	app.Handler(http.MethodGet, "/Handler", httpHandler)
	app.HandlerFunc(http.MethodGet, "/HandlerFunc", func(w http.ResponseWriter, r *http.Request) {
		handlerFunc = true
	})

	w := new(mockResponseWriter)

	r, _ := http.NewRequest(http.MethodGet, "/GET", nil)
	app.ServeHTTP(w, r)
	assert.True(t, get, "routing GET failed")

	r, _ = http.NewRequest(http.MethodHead, "/GET", nil)
	app.ServeHTTP(w, r)
	assert.True(t, head, "routing HEAD failed")

	r, _ = http.NewRequest(http.MethodOptions, "/GET", nil)
	app.ServeHTTP(w, r)
	assert.True(t, options, "routing OPTIONS failed")

	r, _ = http.NewRequest(http.MethodPost, "/POST", nil)
	app.ServeHTTP(w, r)
	assert.True(t, post, "routing POST failed")

	r, _ = http.NewRequest(http.MethodPut, "/PUT", nil)
	app.ServeHTTP(w, r)
	assert.True(t, put, "routing PUT failed")

	r, _ = http.NewRequest(http.MethodPatch, "/PATCH", nil)
	app.ServeHTTP(w, r)
	assert.True(t, patch, "routing PATCH failed")

	r, _ = http.NewRequest(http.MethodDelete, "/DELETE", nil)
	app.ServeHTTP(w, r)
	assert.True(t, delete, "routing DELETE failed")

	r, _ = http.NewRequest(http.MethodGet, "/Handler", nil)
	app.ServeHTTP(w, r)
	assert.True(t, handler, "routing Handler failed")

	r, _ = http.NewRequest(http.MethodGet, "/HandlerFunc", nil)
	app.ServeHTTP(w, r)
	assert.True(t, handlerFunc, "routing HandlerFunc failed")
}

func TestApplicationAny(t *testing.T) {
	app := New()
	handle := func(ctx *Context) error {
		ctx.WriteString(ctx.Request.Method)
		return nil
	}
	nameOpt := RouteName("ping")
	app.Any("/ping", handle, nameOpt)
	group := app.Group("/foo")
	group.Any("/ping", handle, nameOpt)
	paths := []string{"/ping", "/foo/ping"}
	for _, method := range requestMethods {
		for _, path := range paths {
			w := httptest.NewRecorder()
			app.ServeHTTP(w, httptest.NewRequest(method, path, nil))
			assert.Equal(t, method, w.Body.String())
		}
	}
	url, err := app.RouteURL("ping")
	assert.Nil(t, err)
	assert.Equal(t, "/ping", url.String())
}

func TestApplicationInvalidInput(t *testing.T) {
	app := New()

	handle := func(ctx *Context) error {
		return nil
	}

	recv := catchPanic(func() {
		app.Handle("", "/", handle)
	})
	assert.NotNil(t, recv, "registering empty method did not panic")

	recv = catchPanic(func() {
		app.Get("", handle)
	})
	assert.NotNil(t, recv, "registering empty path did not panic")

	recv = catchPanic(func() {
		app.Get("noSlashRoot", handle)
	})
	assert.NotNil(t, recv, "registering path not beginning with '/' did not panic")

	recv = catchPanic(func() {
		app.Get("/", nil)
	})
	assert.NotNil(t, recv, "registering nil handler did not panic")
}

func TestApplicationChaining(t *testing.T) {
	app1 := New()
	app2 := New()
	app1.NotFound = app2

	fooHit := false
	app1.Post("/foo", func(ctx *Context) error {
		fooHit = true
		ctx.Response.WriteHeader(http.StatusOK)
		return nil
	})

	barHit := false
	app2.Post("/bar", func(ctx *Context) error {
		barHit = true
		ctx.Response.WriteHeader(http.StatusOK)
		return nil
	})

	r, _ := http.NewRequest(http.MethodPost, "/foo", nil)
	w := httptest.NewRecorder()
	app1.ServeHTTP(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, fooHit)

	r, _ = http.NewRequest(http.MethodPost, "/bar", nil)
	w = httptest.NewRecorder()
	app1.ServeHTTP(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, barHit)

	r, _ = http.NewRequest(http.MethodPost, "/qax", nil)
	w = httptest.NewRecorder()
	app1.ServeHTTP(w, r)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func BenchmarkAllowed(b *testing.B) {
	handlerFunc := func(ctx *Context) error {
		return nil
	}

	app := New()
	app.Post("/path", handlerFunc)
	app.Get("/path", handlerFunc)

	b.Run("Global", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = app.allowed("*", http.MethodOptions)
		}
	})
	b.Run("Path", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = app.allowed("/path", http.MethodOptions)
		}
	})
}

func TestApplicationOPTIONS(t *testing.T) {
	handlerFunc := func(ctx *Context) error {
		return nil
	}

	app := New()
	app.Post("/path", handlerFunc)

	// test not allowed
	// * (server)
	r, _ := http.NewRequest(http.MethodOptions, "*", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "OPTIONS, POST", w.Header().Get("Allow"))

	// path
	r, _ = http.NewRequest(http.MethodOptions, "/path", nil)
	w = httptest.NewRecorder()
	app.ServeHTTP(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "OPTIONS, POST", w.Header().Get("Allow"))

	r, _ = http.NewRequest(http.MethodOptions, "/doesnotexist", nil)
	w = httptest.NewRecorder()
	app.ServeHTTP(w, r)
	assert.Equal(t, http.StatusNotFound, w.Code)

	// add another method
	app.Get("/path", handlerFunc)

	// set a global OPTIONS handler
	app.GlobalOPTIONS = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Adjust status code to 204
		w.WriteHeader(http.StatusNoContent)
	})

	// test again
	// * (server)
	r, _ = http.NewRequest(http.MethodOptions, "*", nil)
	w = httptest.NewRecorder()
	app.ServeHTTP(w, r)
	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "GET, OPTIONS, POST", w.Header().Get("Allow"))

	// path
	r, _ = http.NewRequest(http.MethodOptions, "/path", nil)
	w = httptest.NewRecorder()
	app.ServeHTTP(w, r)
	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "GET, OPTIONS, POST", w.Header().Get("Allow"))

	// custom handler
	var custom bool
	app.Options("/path", func(ctx *Context) error {
		custom = true
		return nil
	})

	// test again
	// * (server)
	r, _ = http.NewRequest(http.MethodOptions, "*", nil)
	w = httptest.NewRecorder()
	app.ServeHTTP(w, r)
	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "GET, OPTIONS, POST", w.Header().Get("Allow"))
	assert.False(t, custom, "custom handler called on *")

	// path
	r, _ = http.NewRequest(http.MethodOptions, "/path", nil)
	w = httptest.NewRecorder()
	app.ServeHTTP(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, custom, "custom handler not called*")
}

func TestApplicationNotAllowed(t *testing.T) {
	handlerFunc := func(ctx *Context) error {
		return nil
	}

	app := New()
	app.Post("/path", handlerFunc)

	// test not allowed
	r, _ := http.NewRequest(http.MethodGet, "/path", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	assert.Equal(t, "OPTIONS, POST", w.Header().Get("Allow"))

	// add another method
	app.Delete("/path", handlerFunc)
	app.Options("/path", handlerFunc) // must be ignored

	// test again
	r, _ = http.NewRequest(http.MethodGet, "/path", nil)
	w = httptest.NewRecorder()
	app.ServeHTTP(w, r)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	assert.Equal(t, "DELETE, OPTIONS, POST", w.Header().Get("Allow"))

	// test custom handler
	w = httptest.NewRecorder()
	responseText := "custom method"
	app.MethodNotAllowed = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		w.Write([]byte(responseText))
	})
	app.ServeHTTP(w, r)
	assert.Equal(t, responseText, w.Body.String())
	assert.Equal(t, http.StatusTeapot, w.Code)
	assert.Equal(t, "DELETE, OPTIONS, POST", w.Header().Get("Allow"))
}

func TestApplicationNotFound(t *testing.T) {
	handlerFunc := func(ctx *Context) error {
		return nil
	}

	app := New()
	app.Get("/path", handlerFunc)
	app.Get("/dir/", handlerFunc)
	app.Get("/", handlerFunc)

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
		app.ServeHTTP(w, r)
		assert.Equal(t, tr.code, w.Code)
		if w.Code != http.StatusNotFound {
			assert.Equal(t, tr.location, fmt.Sprint(w.Header().Get("Location")))
		}
	}

	// Test custom not found handler
	var notFound bool
	app.NotFound = http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusNotFound)
		notFound = true
	})
	r, _ := http.NewRequest(http.MethodGet, "/nope", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	assert.True(t, notFound)
	assert.Equal(t, http.StatusNotFound, w.Code)

	// Test other method than GET (want 308 instead of 301)
	app.Patch("/path", handlerFunc)
	r, _ = http.NewRequest(http.MethodPatch, "/path/", nil)
	w = httptest.NewRecorder()
	app.ServeHTTP(w, r)
	assert.Equal(t, http.StatusPermanentRedirect, w.Code)
	assert.Equal(t, "map[Location:[/path]]", fmt.Sprint(w.Header()))

	// Test special case where no node for the prefix "/" exists
	app = New()
	app.Get("/a", handlerFunc)
	r, _ = http.NewRequest(http.MethodGet, "/", nil)
	w = httptest.NewRecorder()
	app.ServeHTTP(w, r)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestApplicationLookup(t *testing.T) {
	routed := false
	wantHandle := func(ctx *Context) error {
		routed = true
		return nil
	}
	wantParams := Params{Param{"name", "gopher"}}

	app := New()

	// try empty trees first
	route, _, tsr := app.Lookup(http.MethodGet, "/nope")
	assert.Nil(t, route, "Got route for unregistered pattern: %v", route)
	assert.False(t, tsr, "Got wrong TSR recommendation!")

	// insert route and try again
	app.Get("/user/:name", wantHandle)
	route, params, _ := app.Lookup(http.MethodGet, "/user/gopher")
	assert.NotNil(t, route, "Got no route!")
	route.handle(newContext(nil, nil))
	assert.True(t, routed)
	assert.Equal(t, wantParams, params)

	routed = false

	// route without param
	app.Get("/user", wantHandle)
	route, params, _ = app.Lookup(http.MethodGet, "/user")
	assert.NotNil(t, route, "Got no route!")
	route.handle(newContext(nil, nil))
	assert.True(t, routed)
	assert.Len(t, params, 0)

	route, _, tsr = app.Lookup(http.MethodGet, "/user/gopher/")
	assert.Nil(t, route, "Got route for unregistered pattern: %v", route)
	assert.True(t, tsr, "Got no TSR recommendation!")

	route, _, tsr = app.Lookup(http.MethodGet, "/nope")
	assert.Nilf(t, route, "Got route for unregistered pattern: %v", route)
	assert.False(t, tsr, "Got wrong TSR recommendation!")
}

func TestApplicationParamsFromContext(t *testing.T) {
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
	app := New()
	app.Handle(http.MethodGet, "/user", handlerFuncNil)
	app.Handle(http.MethodGet, "/user/:name", handlerFunc)

	w := new(mockResponseWriter)
	r, _ := http.NewRequest(http.MethodGet, "/user/gopher", nil)
	app.ServeHTTP(w, r)
	assert.True(t, routed)

	routed = false
	r, _ = http.NewRequest(http.MethodGet, "/user", nil)
	app.ServeHTTP(w, r)
	assert.True(t, routed)
}

func TestApplicationMatchedRoutePath(t *testing.T) {
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

	app := New()
	app.Handle(http.MethodGet, route1, handle1)
	app.Handle(http.MethodGet, route2, handle2)
	app.Handle(http.MethodGet, route3, handle3)

	w := new(mockResponseWriter)
	r, _ := http.NewRequest(http.MethodGet, "/user/gopher", nil)
	app.ServeHTTP(w, r)
	assert.True(t, routed1)

	w = new(mockResponseWriter)
	r, _ = http.NewRequest(http.MethodGet, "/user/gopher/details", nil)
	app.ServeHTTP(w, r)
	assert.True(t, routed2)

	w = new(mockResponseWriter)
	r, _ = http.NewRequest(http.MethodGet, "/", nil)
	app.ServeHTTP(w, r)
	assert.True(t, routed3)
}

type mockFileSystem struct {
	opened bool
}

func (mfs *mockFileSystem) Open(name string) (http.File, error) {
	mfs.opened = true
	return nil, errors.New("this is just a mock")
}

func TestApplicationServeFiles(t *testing.T) {
	app := New()
	mfs := &mockFileSystem{}

	recv := catchPanic(func() {
		app.ServeFiles("/noFilepath", mfs)
	})
	assert.NotNil(t, recv, "registering path not ending with '*filepath' did not panic")

	app.ServeFiles("/*filepath", mfs)
	w := new(mockResponseWriter)
	r, _ := http.NewRequest(http.MethodGet, "/favicon.ico", nil)
	app.ServeHTTP(w, r)
	assert.True(t, mfs.opened, "serving file failed")
}

func TestApplicationNamedRoute(t *testing.T) {
	tests := []struct {
		path        string
		name        string
		args        []string
		expectedURL string
	}{
		{"/", "home", nil, "/"},
		{"/users/:id", "user", []string{"id", "foo"}, "/users/foo"},
	}
	app := New()
	for _, test := range tests {
		app.Handle(http.MethodGet, test.path, func(ctx *Context) error {

			return nil
		}, RouteName(test.name))
		url, err := app.RouteURL(test.name, test.args...)
		assert.Nil(t, err)
		assert.Equal(t, test.expectedURL, url.String())
	}

	// unregistered name.
	_, err := app.RouteURL("unregistered")
	assert.NotNil(t, err)

	// registers same route name.
	recv := catchPanic(func() {
		app.Handle(http.MethodGet, "/same", func(ctx *Context) error {
			return nil
		}, RouteName("home"))
	})
	assert.NotNil(t, recv)
}

func ExampleApplication_RouteURL() {
	app := New()
	app.Get("/hello/:name", func(ctx *Context) error {
		return nil
	}, RouteName("hello"))
	// nested routes group
	api := app.Group("/api")

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
		url, _ := app.RouteURL(route.name, route.args...)
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

func ExampleApplication_ServeFiles() {
	app := New()

	app.ServeFiles("/static/*filepath", http.Dir("/path/to/static"))

	// sometimes, it is useful to treat http.FileServer as NotFoundHandler,
	// such as "/favicon.ico".
	app.NotFound = http.FileServer(http.Dir("public"))
}

type testErrorHandler struct {
	status int
}

func (eh testErrorHandler) Handle(ctx *Context, err error) {
	ctx.Error(eh.status, err.Error())
}

func TestApplication_ErrorHandler(t *testing.T) {
	app := New()
	app.ErrorHandler = &testErrorHandler{http.StatusInternalServerError}
	app.Get("/error/:msg", func(ctx *Context) error {
		return errors.New(ctx.Params.String("msg"))
	})

	msgs := []string{"foo", "bar"}
	for _, msg := range msgs {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/error/"+msg, nil)
		app.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, fmt.Sprintln(msg), w.Body.String())
	}
}

func TestApplication_HandleError(t *testing.T) {
	app := New()
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
		app.HandleError(ctx, test.err)
		assert.Equal(t, test.code, w.Code)
		assert.Equal(t, fmt.Sprintln(test.body), w.Body.String())
	}
}

func TestApplicationUseRawPath(t *testing.T) {
	app := New()
	app.UseRawPath = true
	handled := false
	handle := func(ctx *Context) error {
		expected := Params{Param{"name", "foo/bar"}}
		assert.Equal(t, expected, ctx.Params)
		handled = true
		return nil
	}
	app.Get("/hello/:name", handle)
	req := httptest.NewRequest(http.MethodGet, "/hello/foo%2fbar", nil)
	app.ServeHTTP(nil, req)
	assert.True(t, handled, "raw path routing failed")
}

func TestApplicationUseRawPathMixed(t *testing.T) {
	app := New()
	app.UseRawPath = true
	handled := false
	handle := func(ctx *Context) error {
		expected := Params{Param{"date", "2020/03/23"}, Param{"slug", "hello world"}}
		assert.Equal(t, expected, ctx.Params)
		handled = true
		return nil
	}
	app.Get("/post/:date/:slug", handle)
	req := httptest.NewRequest(http.MethodGet, "/post/2020%2f03%2f23/hello%20world", nil)
	app.ServeHTTP(nil, req)
	assert.True(t, handled, "raw path routing failed")
}

func TestApplicationUseRawPathCatchAll(t *testing.T) {
	app := New()
	app.UseRawPath = true
	handled := false
	handle := func(ctx *Context) error {
		expected := Params{Param{"slug", "/2020/03/23-hello world"}}
		assert.Equal(t, expected, ctx.Params)
		handled = true
		return nil
	}
	app.Get("/post/*slug", handle)
	req := httptest.NewRequest(http.MethodGet, "/post/2020%2f03%2f23-hello%20world", nil)
	app.ServeHTTP(nil, req)
	assert.True(t, handled, "raw path routing failed")
}

func TestApplicationUse(t *testing.T) {
	app := New()
	app.Use(
		echoMiddleware("m1"),
		echoMiddleware("m2"),
	)
	app.Get("/", echoHandler("foobar"))
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	app.ServeHTTP(w, req)
	assert.Equal(t, "m1 m2 foobar", w.Body.String())

	app.Use(terminatedMiddleware())
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	app.ServeHTTP(w, req)
	assert.Equal(t, "m1 m2 terminated", w.Body.String())
}
