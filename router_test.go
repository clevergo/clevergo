// Copyright 2013 Julien Schmidt. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package clevergo

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
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

func TestParams(t *testing.T) {
	ps := Params{
		Param{"param1", "value1"},
		Param{"param2", "value2"},
		Param{"param3", "value3"},
	}
	for i := range ps {
		if val := ps.String(ps[i].Key); val != ps[i].Value {
			t.Errorf("Wrong value for %s: Got %s; Want %s", ps[i].Key, val, ps[i].Value)
		}
	}
	if val := ps.String("noKey"); val != "" {
		t.Errorf("Expected empty string for not found key; got %q", val)
	}
}

func TestParams_Int(t *testing.T) {
	ps := Params{
		Param{"param1", "-1"},
		Param{"param2", "0"},
		Param{"param3", "1"},
	}
	tests := map[string]int{
		"param1": -1,
		"param2": 0,
		"param3": 1,
	}
	for name, value := range tests {
		if val, err := ps.Int(name); err != nil || val != value {
			t.Errorf("Wrong value for %s: Got %d; Want %d", name, val, value)
		}
	}
	if val, err := ps.Int("noKey"); err == nil {
		t.Errorf("Expected an error for not found key; got %d", val)
	}
}

func TestParams_Int64(t *testing.T) {
	ps := Params{
		Param{"param1", "-1"},
		Param{"param2", "0"},
		Param{"param3", "1"},
	}
	tests := map[string]int64{
		"param1": -1,
		"param2": 0,
		"param3": 1,
	}
	for name, value := range tests {
		if val, err := ps.Int64(name); err != nil || val != value {
			t.Errorf("Wrong value for %s: Got %d; Want %d", name, val, value)
		}
	}
	if val, err := ps.Int64("noKey"); err == nil {
		t.Errorf("Expected an error for not found key; got %d", val)
	}
}

func TestParams_Uint64(t *testing.T) {
	ps := Params{
		Param{"param1", "0"},
		Param{"param2", "1"},
	}
	tests := map[string]uint64{
		"param1": 0,
		"param2": 1,
	}
	for name, value := range tests {
		if val, err := ps.Uint64(name); err != nil || val != value {
			t.Errorf("Wrong value for %s: Got %d; Want %d", name, val, value)
		}
	}
	if val, err := ps.Uint64("noKey"); err == nil {
		t.Errorf("Expected an error for not found key; got %d", val)
	}
}

func TestParams_Float(t *testing.T) {
	ps := Params{
		Param{"param1", "-0.2"},
		Param{"param2", "0.2"},
		Param{"param3", "1.9"},
	}
	tests := map[string]float64{
		"param1": -0.2,
		"param2": 0.2,
		"param3": 1.9,
	}
	for name, value := range tests {
		if val, err := ps.Float64(name); err != nil || val != value {
			t.Errorf("Wrong value for %s: Got %f; Want %f", name, val, value)
		}
	}
	if val, err := ps.Float64("noKey"); err == nil {
		t.Errorf("Expected an error for not found key; got %f", val)
	}
}

func TestParams_Bool(t *testing.T) {
	ps := Params{
		Param{"param1", "1"},
		Param{"param2", "t"},
		Param{"param3", "T"},
		Param{"param4", "true"},
		Param{"param5", "TRUE"},
		Param{"param6", "True"},
		Param{"param7", "0"},
		Param{"param8", "f"},
		Param{"param9", "F"},
		Param{"param10", "false"},
		Param{"param11", "FALSE"},
		Param{"param12", "False"},
	}
	tests := map[string]bool{
		"param1":  true,
		"param2":  true,
		"param3":  true,
		"param4":  true,
		"param5":  true,
		"param6":  true,
		"param7":  false,
		"param8":  false,
		"param9":  false,
		"param10": false,
		"param11": false,
		"param12": false,
	}
	for name, value := range tests {
		if val, err := ps.Bool(name); err != nil || val != value {
			t.Errorf("Wrong value for %s: Got %t; Want %t", name, val, value)
		}
	}
	if val, err := ps.Bool("noKey"); err == nil {
		t.Errorf("Expected an error for not found key; got %t", val)
	}
}

func TestRouter(t *testing.T) {
	router := NewRouter()

	routed := false
	router.Handle(http.MethodGet, "/user/:name", func(ctx *Context) {
		routed = true
		want := Params{Param{"name", "gopher"}}
		if !reflect.DeepEqual(ctx.Params, want) {
			t.Fatalf("wrong wildcard values: want %v, got %v", want, ctx.Params)
		}
	})

	w := new(mockResponseWriter)

	req, _ := http.NewRequest(http.MethodGet, "/user/gopher", nil)
	router.ServeHTTP(w, req)

	if !routed {
		t.Fatal("routing failed")
	}
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
	router.Get("/GET", func(ctx *Context) {
		get = true
	})
	router.Head("/GET", func(ctx *Context) {
		head = true
	})
	router.Options("/GET", func(ctx *Context) {
		options = true
	})
	router.Post("/POST", func(ctx *Context) {
		post = true
	})
	router.Put("/PUT", func(ctx *Context) {
		put = true
	})
	router.Patch("/PATCH", func(ctx *Context) {
		patch = true
	})
	router.Delete("/DELETE", func(ctx *Context) {
		delete = true
	})
	router.Handler(http.MethodGet, "/Handler", httpHandler)
	router.HandlerFunc(http.MethodGet, "/HandlerFunc", func(w http.ResponseWriter, r *http.Request) {
		handlerFunc = true
	})

	w := new(mockResponseWriter)

	r, _ := http.NewRequest(http.MethodGet, "/GET", nil)
	router.ServeHTTP(w, r)
	if !get {
		t.Error("routing GET failed")
	}

	r, _ = http.NewRequest(http.MethodHead, "/GET", nil)
	router.ServeHTTP(w, r)
	if !head {
		t.Error("routing HEAD failed")
	}

	r, _ = http.NewRequest(http.MethodOptions, "/GET", nil)
	router.ServeHTTP(w, r)
	if !options {
		t.Error("routing OPTIONS failed")
	}

	r, _ = http.NewRequest(http.MethodPost, "/POST", nil)
	router.ServeHTTP(w, r)
	if !post {
		t.Error("routing POST failed")
	}

	r, _ = http.NewRequest(http.MethodPut, "/PUT", nil)
	router.ServeHTTP(w, r)
	if !put {
		t.Error("routing PUT failed")
	}

	r, _ = http.NewRequest(http.MethodPatch, "/PATCH", nil)
	router.ServeHTTP(w, r)
	if !patch {
		t.Error("routing PATCH failed")
	}

	r, _ = http.NewRequest(http.MethodDelete, "/DELETE", nil)
	router.ServeHTTP(w, r)
	if !delete {
		t.Error("routing DELETE failed")
	}

	r, _ = http.NewRequest(http.MethodGet, "/Handler", nil)
	router.ServeHTTP(w, r)
	if !handler {
		t.Error("routing Handler failed")
	}

	r, _ = http.NewRequest(http.MethodGet, "/HandlerFunc", nil)
	router.ServeHTTP(w, r)
	if !handlerFunc {
		t.Error("routing HandlerFunc failed")
	}
}

func TestRouterInvalidInput(t *testing.T) {
	router := NewRouter()

	handle := func(ctx *Context) {}

	recv := catchPanic(func() {
		router.Handle("", "/", handle)
	})
	if recv == nil {
		t.Fatal("registering empty method did not panic")
	}

	recv = catchPanic(func() {
		router.Get("", handle)
	})
	if recv == nil {
		t.Fatal("registering empty path did not panic")
	}

	recv = catchPanic(func() {
		router.Get("noSlashRoot", handle)
	})
	if recv == nil {
		t.Fatal("registering path not beginning with '/' did not panic")
	}

	recv = catchPanic(func() {
		router.Get("/", nil)
	})
	if recv == nil {
		t.Fatal("registering nil handler did not panic")
	}
}

func TestRouterChaining(t *testing.T) {
	router1 := NewRouter()
	router2 := NewRouter()
	router1.NotFound = router2

	fooHit := false
	router1.Post("/foo", func(ctx *Context) {
		fooHit = true
		ctx.Response.WriteHeader(http.StatusOK)
	})

	barHit := false
	router2.Post("/bar", func(ctx *Context) {
		barHit = true
		ctx.Response.WriteHeader(http.StatusOK)
	})

	r, _ := http.NewRequest(http.MethodPost, "/foo", nil)
	w := httptest.NewRecorder()
	router1.ServeHTTP(w, r)
	if !(w.Code == http.StatusOK && fooHit) {
		t.Errorf("Regular routing failed with router chaining.")
		t.FailNow()
	}

	r, _ = http.NewRequest(http.MethodPost, "/bar", nil)
	w = httptest.NewRecorder()
	router1.ServeHTTP(w, r)
	if !(w.Code == http.StatusOK && barHit) {
		t.Errorf("Chained routing failed with router chaining.")
		t.FailNow()
	}

	r, _ = http.NewRequest(http.MethodPost, "/qax", nil)
	w = httptest.NewRecorder()
	router1.ServeHTTP(w, r)
	if !(w.Code == http.StatusNotFound) {
		t.Errorf("NotFound behavior failed with router chaining.")
		t.FailNow()
	}
}

func BenchmarkAllowed(b *testing.B) {
	handlerFunc := func(ctx *Context) {}

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
	handlerFunc := func(ctx *Context) {}

	router := NewRouter()
	router.Post("/path", handlerFunc)

	// test not allowed
	// * (server)
	r, _ := http.NewRequest(http.MethodOptions, "*", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)
	if !(w.Code == http.StatusOK) {
		t.Errorf("OPTIONS handling failed: Code=%d, Header=%v", w.Code, w.Header())
	} else if allow := w.Header().Get("Allow"); allow != "OPTIONS, POST" {
		t.Error("unexpected Allow header value: " + allow)
	}

	// path
	r, _ = http.NewRequest(http.MethodOptions, "/path", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	if !(w.Code == http.StatusOK) {
		t.Errorf("OPTIONS handling failed: Code=%d, Header=%v", w.Code, w.Header())
	} else if allow := w.Header().Get("Allow"); allow != "OPTIONS, POST" {
		t.Error("unexpected Allow header value: " + allow)
	}

	r, _ = http.NewRequest(http.MethodOptions, "/doesnotexist", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	if !(w.Code == http.StatusNotFound) {
		t.Errorf("OPTIONS handling failed: Code=%d, Header=%v", w.Code, w.Header())
	}

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
	if !(w.Code == http.StatusNoContent) {
		t.Errorf("OPTIONS handling failed: Code=%d, Header=%v", w.Code, w.Header())
	} else if allow := w.Header().Get("Allow"); allow != "GET, OPTIONS, POST" {
		t.Error("unexpected Allow header value: " + allow)
	}

	// path
	r, _ = http.NewRequest(http.MethodOptions, "/path", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	if !(w.Code == http.StatusNoContent) {
		t.Errorf("OPTIONS handling failed: Code=%d, Header=%v", w.Code, w.Header())
	} else if allow := w.Header().Get("Allow"); allow != "GET, OPTIONS, POST" {
		t.Error("unexpected Allow header value: " + allow)
	}

	// custom handler
	var custom bool
	router.Options("/path", func(ctx *Context) {
		custom = true
	})

	// test again
	// * (server)
	r, _ = http.NewRequest(http.MethodOptions, "*", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	if !(w.Code == http.StatusNoContent) {
		t.Errorf("OPTIONS handling failed: Code=%d, Header=%v", w.Code, w.Header())
	} else if allow := w.Header().Get("Allow"); allow != "GET, OPTIONS, POST" {
		t.Error("unexpected Allow header value: " + allow)
	}
	if custom {
		t.Error("custom handler called on *")
	}

	// path
	r, _ = http.NewRequest(http.MethodOptions, "/path", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	if !(w.Code == http.StatusOK) {
		t.Errorf("OPTIONS handling failed: Code=%d, Header=%v", w.Code, w.Header())
	}
	if !custom {
		t.Error("custom handler not called")
	}
}

func TestRouterNotAllowed(t *testing.T) {
	handlerFunc := func(ctx *Context) {}

	router := NewRouter()
	router.Post("/path", handlerFunc)

	// test not allowed
	r, _ := http.NewRequest(http.MethodGet, "/path", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)
	if !(w.Code == http.StatusMethodNotAllowed) {
		t.Errorf("NotAllowed handling failed: Code=%d, Header=%v", w.Code, w.Header())
	} else if allow := w.Header().Get("Allow"); allow != "OPTIONS, POST" {
		t.Error("unexpected Allow header value: " + allow)
	}

	// add another method
	router.Delete("/path", handlerFunc)
	router.Options("/path", handlerFunc) // must be ignored

	// test again
	r, _ = http.NewRequest(http.MethodGet, "/path", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	if !(w.Code == http.StatusMethodNotAllowed) {
		t.Errorf("NotAllowed handling failed: Code=%d, Header=%v", w.Code, w.Header())
	} else if allow := w.Header().Get("Allow"); allow != "DELETE, OPTIONS, POST" {
		t.Error("unexpected Allow header value: " + allow)
	}

	// test custom handler
	w = httptest.NewRecorder()
	responseText := "custom method"
	router.MethodNotAllowed = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		w.Write([]byte(responseText))
	})
	router.ServeHTTP(w, r)
	if got := w.Body.String(); !(got == responseText) {
		t.Errorf("unexpected response got %q want %q", got, responseText)
	}
	if w.Code != http.StatusTeapot {
		t.Errorf("unexpected response code %d want %d", w.Code, http.StatusTeapot)
	}
	if allow := w.Header().Get("Allow"); allow != "DELETE, OPTIONS, POST" {
		t.Error("unexpected Allow header value: " + allow)
	}
}

func TestRouterNotFound(t *testing.T) {
	handlerFunc := func(ctx *Context) {}

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
		if !(w.Code == tr.code && (w.Code == http.StatusNotFound || fmt.Sprint(w.Header().Get("Location")) == tr.location)) {
			t.Errorf("NotFound handling route %s failed: Code=%d, Header=%v", tr.route, w.Code, w.Header().Get("Location"))
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
	if !(w.Code == http.StatusNotFound && notFound == true) {
		t.Errorf("Custom NotFound handler failed: Code=%d, Header=%v", w.Code, w.Header())
	}

	// Test other method than GET (want 308 instead of 301)
	router.Patch("/path", handlerFunc)
	r, _ = http.NewRequest(http.MethodPatch, "/path/", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	if !(w.Code == http.StatusPermanentRedirect && fmt.Sprint(w.Header()) == "map[Location:[/path]]") {
		t.Errorf("Custom NotFound handler failed: Code=%d, Header=%v", w.Code, w.Header())
	}

	// Test special case where no node for the prefix "/" exists
	router = NewRouter()
	router.Get("/a", handlerFunc)
	r, _ = http.NewRequest(http.MethodGet, "/", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	if !(w.Code == http.StatusNotFound) {
		t.Errorf("NotFound handling route / failed: Code=%d", w.Code)
	}
}

func TestRouterLookup(t *testing.T) {
	routed := false
	wantHandle := func(ctx *Context) {
		routed = true
	}
	wantParams := Params{Param{"name", "gopher"}}

	router := NewRouter()

	// try empty router first
	route, _, tsr := router.Lookup(http.MethodGet, "/nope")
	if route != nil {
		t.Fatalf("Got route for unregistered pattern: %v", route)
	}
	if tsr {
		t.Error("Got wrong TSR recommendation!")
	}

	// insert route and try again
	router.Get("/user/:name", wantHandle)
	route, params, _ := router.Lookup(http.MethodGet, "/user/gopher")
	if route == nil {
		t.Fatal("Got no route!")
	} else {
		route.handle(newContext(nil, nil))
		if !routed {
			t.Fatal("Routing failed!")
		}
	}
	if !reflect.DeepEqual(params, wantParams) {
		t.Fatalf("Wrong parameter values: want %v, got %v", wantParams, params)
	}
	routed = false

	// route without param
	router.Get("/user", wantHandle)
	route, params, _ = router.Lookup(http.MethodGet, "/user")
	if route == nil {
		t.Fatal("Got no route!")
	} else {
		route.handle(newContext(nil, nil))
		if !routed {
			t.Fatal("Routing failed!")
		}
	}
	if params != nil {
		t.Fatalf("Wrong parameter values: want %v, got %v", nil, params)
	}

	route, _, tsr = router.Lookup(http.MethodGet, "/user/gopher/")
	if route != nil {
		t.Fatalf("Got route for unregistered pattern: %v", route)
	}
	if !tsr {
		t.Error("Got no TSR recommendation!")
	}

	route, _, tsr = router.Lookup(http.MethodGet, "/nope")
	if route != nil {
		t.Fatalf("Got route for unregistered pattern: %v", route)
	}
	if tsr {
		t.Error("Got wrong TSR recommendation!")
	}
}

func TestRouterParamsFromContext(t *testing.T) {
	routed := false

	wantParams := Params{Param{"name", "gopher"}}
	handlerFunc := func(ctx *Context) {
		if !reflect.DeepEqual(ctx.Params, wantParams) {
			t.Fatalf("Wrong parameter values: want %v, got %v", wantParams, ctx.Params)
		}

		routed = true
	}

	var nilParams Params
	handlerFuncNil := func(ctx *Context) {
		if !reflect.DeepEqual(ctx.Params, nilParams) {
			t.Fatalf("Wrong parameter values: want %v, got %v", nilParams, ctx.Params)
		}

		routed = true
	}
	router := NewRouter()
	router.Handle(http.MethodGet, "/user", handlerFuncNil)
	router.Handle(http.MethodGet, "/user/:name", handlerFunc)

	w := new(mockResponseWriter)
	r, _ := http.NewRequest(http.MethodGet, "/user/gopher", nil)
	router.ServeHTTP(w, r)
	if !routed {
		t.Fatal("Routing failed!")
	}

	routed = false
	r, _ = http.NewRequest(http.MethodGet, "/user", nil)
	router.ServeHTTP(w, r)
	if !routed {
		t.Fatal("Routing failed!")
	}
}

func TestRouterMatchedRoutePath(t *testing.T) {
	route1 := "/user/:name"
	routed1 := false
	handle1 := func(ctx *Context) {
		if ctx.Route.path != route1 {
			t.Fatalf("Wrong matched route: want %s, got %s", route1, ctx.Route.path)
		}
		routed1 = true
	}

	route2 := "/user/:name/details"
	routed2 := false
	handle2 := func(ctx *Context) {
		if ctx.Route.path != route2 {
			t.Fatalf("Wrong matched route: want %s, got %s", route2, ctx.Route.path)
		}
		routed2 = true
	}

	route3 := "/"
	routed3 := false
	handle3 := func(ctx *Context) {
		if ctx.Route.path != route3 {
			t.Fatalf("Wrong matched route: want %s, got %s", route3, ctx.Route.path)
		}
		routed3 = true
	}

	router := NewRouter()
	router.Handle(http.MethodGet, route1, handle1)
	router.Handle(http.MethodGet, route2, handle2)
	router.Handle(http.MethodGet, route3, handle3)

	w := new(mockResponseWriter)
	r, _ := http.NewRequest(http.MethodGet, "/user/gopher", nil)
	router.ServeHTTP(w, r)
	if !routed1 || routed2 || routed3 {
		t.Fatal("Routing failed!")
	}

	w = new(mockResponseWriter)
	r, _ = http.NewRequest(http.MethodGet, "/user/gopher/details", nil)
	router.ServeHTTP(w, r)
	if !routed2 || routed3 {
		t.Fatal("Routing failed!")
	}

	w = new(mockResponseWriter)
	r, _ = http.NewRequest(http.MethodGet, "/", nil)
	router.ServeHTTP(w, r)
	if !routed3 {
		t.Fatal("Routing failed!")
	}
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
	if recv == nil {
		t.Fatal("registering path not ending with '*filepath' did not panic")
	}

	router.ServeFiles("/*filepath", mfs)
	w := new(mockResponseWriter)
	r, _ := http.NewRequest(http.MethodGet, "/favicon.ico", nil)
	router.ServeHTTP(w, r)
	if !mfs.opened {
		t.Error("serving file failed")
	}
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
		router.Handle(http.MethodGet, test.path, func(ctx *Context) {}, RouteName(test.name))
		url, err := router.URL(test.name, test.args...)
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		if url.String() != test.expectedURL {
			t.Errorf("expected url: %q, got %q", test.expectedURL, url)
		}
	}

	// unregistered name.
	_, err := router.URL("unregistered")
	if err == nil {
		t.Error("expected an error, got nil")
	}

	// registers same route name.
	recv := catchPanic(func() {
		router.Handle(http.MethodGet, "/same", func(ctx *Context) {}, RouteName("home"))
	})
	if recv == nil {
		t.Error("expected a panic, got nil")
	}
}

func ExampleRouter_URL() {
	router := NewRouter()
	router.Get("/hello/:name", func(ctx *Context) {}, RouteName("hello"))
	// nested routes group
	api := router.Group("/api")

	v1 := api.Group("/v1")
	// the group path will become the prefix of route name.
	v1.Get("/users/:name", func(ctx *Context) {}, RouteName("user"))

	v2 := api.Group("/v2")
	v2.Get("/users/:name", func(ctx *Context) {}, RouteName("user"))

	routes := []struct {
		name string
		args []string
	}{
		{"hello", []string{"name", "foo"}},
		{"hello", []string{"name", "bar"}},
		{"/api/v1/user", []string{"name", "foo"}},
		{"/api/v1/user", []string{"name", "bar"}},
		{"/api/v2/user", []string{"name", "foo"}},
		{"/api/v2/user", []string{"name", "bar"}},
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

func ExampleParams() {
	router := NewRouter()
	router.Get("/post/:year/:month/:title", func(ctx *Context) {
		// converts param value to int.
		year, _ := ctx.Params.Int("year")
		month, _ := ctx.Params.Int("month")
		// ps.Int64("name") // converts to int64.
		// ps.Uint64("name") // converts to uint64.
		// ps.Float64("name") // converts to float64.
		// ps.Bool("name") // converts to boolean.
		fmt.Printf("%s posted on %04d-%02d\n", ctx.Params.String("title"), year, month)
	})
	req := httptest.NewRequest(http.MethodGet, "/post/2020/01/foo", nil)
	router.ServeHTTP(nil, req)

	req = httptest.NewRequest(http.MethodGet, "/post/2020/02/bar", nil)
	router.ServeHTTP(nil, req)

	// Output:
	// foo posted on 2020-01
	// bar posted on 2020-02
}

func ExampleRouter_ServeFiles() {
	router := NewRouter()

	router.ServeFiles("/static/*filepath", http.Dir("/path/to/static"))

	// sometimes, it is useful to treat http.FileServer as NotFoundHandler,
	// such as "/favicon.ico".
	router.NotFound = http.FileServer(http.Dir("public"))
}
