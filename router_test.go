// Copyright 2016 The Gem Authors. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package gem

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/go-gem/tests"
	"github.com/valyala/fasthttp"
)

func TestRouter(t *testing.T) {
	routed := false

	router := NewRouter()
	router.Handle("GET", "/panic", func(ctx *Context) {
		panic("panic handler")
	})
	router.Handle("GET", "/user/:name", func(ctx *Context) {
		routed = true
		want := "gem"
		if !reflect.DeepEqual(ctx.UserValue("name"), want) {
			t.Fatalf("wrong wildcard values: want %v, got %v", want, ctx.UserValue("name"))
		}
		ctx.Success("foo/bar", []byte("success"))
	})

	srv := New("", router.Handler())

	var err error

	test1 := tests.New(srv)
	test1.Url = "/user/gem?baz"
	test1.Expect().Status(fasthttp.StatusOK)
	if err = test1.Run(); err != nil {
		t.Error(err)
	}
	if !routed {
		t.Fatal("routing failed")
	}

	// test default panic
	test2 := tests.New(srv)
	test2.Url = "/panic"
	test2.Expect().Status(fasthttp.StatusInternalServerError)
	if err = test2.Run(); err != nil {
		t.Error(err)
	}
}

func TestRouterAPI(t *testing.T) {
	var get, head, options, post, put, patch, deleted bool

	router := NewRouter()

	router.GET("/GET", func(ctx *Context) {
		get = true
	})
	router.HEAD("/GET", func(ctx *Context) {
		head = true
	})
	router.OPTIONS("/GET", func(ctx *Context) {
		options = true
	})
	router.POST("/POST", func(ctx *Context) {
		post = true
	})
	router.PUT("/PUT", func(ctx *Context) {
		put = true
	})
	router.PATCH("/PATCH", func(ctx *Context) {
		patch = true
	})
	router.DELETE("/DELETE", func(ctx *Context) {
		deleted = true
	})

	srv := New("", router.Handler())

	var err error

	test1 := tests.New(srv)
	test1.Method = MethodGet
	test1.Url = "/GET"
	test1.Expect().Custom(func(_ fasthttp.Response) (err error) {
		if !get {
			err = errors.New("routing GET failed")
		}
		return
	})
	if err = test1.Run(); err != nil {
		t.Error(err)
	}

	test2 := tests.New(srv)
	test2.Method = MethodHead
	test2.Url = "/GET"
	test2.Expect().Custom(func(_ fasthttp.Response) (err error) {
		if !head {
			err = errors.New("routing HEAD failed")
		}
		return
	})
	if err = test2.Run(); err != nil {
		t.Error(err)
	}

	test3 := tests.New(srv)
	test3.Method = MethodOptions
	test3.Url = "/GET"
	test3.Expect().Custom(func(_ fasthttp.Response) (err error) {
		if !options {
			err = errors.New("routing OPTIONS failed")
		}
		return
	})
	if err = test3.Run(); err != nil {
		t.Error(err)
	}

	test4 := tests.New(srv)
	test4.Method = MethodPost
	test4.Url = "/POST"
	test4.Expect().Custom(func(_ fasthttp.Response) (err error) {
		if !post {
			err = errors.New("routing POST failed")
		}
		return
	})
	if err = test4.Run(); err != nil {
		t.Error(err)
	}

	test5 := tests.New(srv)
	test5.Method = MethodPut
	test5.Url = "/PUT"
	test5.Expect().Custom(func(_ fasthttp.Response) (err error) {
		if !put {
			err = errors.New("routing PUT failed")
		}
		return
	})
	if err = test5.Run(); err != nil {
		t.Error(err)
	}

	test6 := tests.New(srv)
	test6.Method = MethodPatch
	test6.Url = "/PATCH"
	test6.Expect().Custom(func(_ fasthttp.Response) (err error) {
		if !patch {
			err = errors.New("routing PATCH failed")
		}
		return
	})
	if err = test6.Run(); err != nil {
		t.Error(err)
	}

	test7 := tests.New(srv)
	test7.Method = MethodDelete
	test7.Url = "/DELETE"
	test7.Expect().Custom(func(_ fasthttp.Response) (err error) {
		if !deleted {
			err = errors.New("routing DELETE failed")
		}
		return
	})
	if err = test7.Run(); err != nil {
		t.Error(err)
	}
}

func TestRouterRoot(t *testing.T) {
	router := NewRouter()
	recv := catchPanic(func() {
		router.GET("noSlashRoot", nil)
	})
	if recv == nil {
		t.Fatal("registering path not beginning with '/' did not panic")
	}
}

func TestRouterChaining(t *testing.T) {
	router := NewRouter()

	router2 := NewRouter()
	router.NotFound = func(ctx *Context) {
		router2.Handler().Handle(ctx)
	}

	fooHit := false
	router.POST("/foo", func(ctx *Context) {
		fooHit = true
		ctx.SetStatusCode(fasthttp.StatusOK)
	})

	barHit := false
	router2.POST("/bar", func(ctx *Context) {
		barHit = true
		ctx.SetStatusCode(fasthttp.StatusOK)
	})

	srv := New("", router.Handler())

	var err error

	test1 := tests.New(srv)
	test1.Method = MethodPost
	test1.Url = "/foo"
	test1.Expect().Status(fasthttp.StatusOK).Custom(func(_ fasthttp.Response) (err error) {
		if !fooHit {
			err = errors.New("Regular routing failed with router chaining.")
		}
		return
	})
	if err = test1.Run(); err != nil {
		t.Error(err)
	}

	test2 := tests.New(srv)
	test2.Method = MethodPost
	test2.Url = "/bar"
	test2.Expect().Status(fasthttp.StatusOK).Custom(func(_ fasthttp.Response) (err error) {
		if !barHit {
			err = errors.New("Chained routing failed with router chaining.")
		}
		return
	})
	if err = test2.Run(); err != nil {
		t.Error(err)
	}

	test3 := tests.New(srv)
	test3.Method = MethodPost
	test3.Url = "/qax"
	test3.Expect().Status(fasthttp.StatusNotFound).Custom(func(_ fasthttp.Response) (err error) {
		if !barHit {
			err = errors.New("NotFound behavior failed with router chaining.")
		}
		return
	})
	if err = test3.Run(); err != nil {
		t.Error(err)
	}
}

func TestRouterOPTIONS(t *testing.T) {
	// TODO: because fasthttp is not support OPTIONS method now,
	// these test cases will be used in the future.
	handlerFunc := func(_ *Context) {}

	router := NewRouter()
	router.POST("/path", handlerFunc)

	// test not allowed
	// * (server)
	srv := New("", router.Handler())

	var err error

	test1 := tests.New(srv)
	test1.Method = MethodOptions
	test1.Url = "*"
	test1.Expect().Status(fasthttp.StatusOK).Custom(func(resp fasthttp.Response) (err error) {
		if allow := string(resp.Header.Peek("Allow")); allow != "POST, OPTIONS" {
			err = fmt.Errorf("unexpected Allow header value: %s", allow)
		}
		return
	})
	if err = test1.Run(); err != nil {
		t.Error(err)
	}

	// path
	test2 := tests.New(srv)
	test2.Method = MethodOptions
	test2.Url = "/path"
	test2.Expect().Status(fasthttp.StatusOK).Custom(func(resp fasthttp.Response) (err error) {
		if allow := string(resp.Header.Peek("Allow")); allow != "POST, OPTIONS" {
			err = fmt.Errorf("unexpected Allow header value: %s", allow)
		}
		return
	})
	if err = test2.Run(); err != nil {
		t.Error(err)
	}

	test3 := tests.New(srv)
	test3.Method = MethodOptions
	test3.Url = "/doesnotexist"
	test3.Expect().Status(fasthttp.StatusNotFound)
	if err = test3.Run(); err != nil {
		t.Error(err)
	}

	// add another method
	router.GET("/path", handlerFunc)

	// test again
	// * (server)
	test4 := tests.New(srv)
	test4.Method = MethodOptions
	test4.Url = "*"
	test4.Expect().Status(fasthttp.StatusOK).Custom(func(resp fasthttp.Response) (err error) {
		if allow := string(resp.Header.Peek("Allow")); allow != "POST, GET, OPTIONS" && allow != "GET, POST, OPTIONS" {
			err = fmt.Errorf("unexpected Allow header value: %s", allow)
		}
		return
	})
	if err = test4.Run(); err != nil {
		t.Error(err)
	}

	// path
	test5 := tests.New(srv)
	test5.Method = MethodOptions
	test5.Url = "/path"
	test5.Expect().Status(fasthttp.StatusOK).Custom(func(resp fasthttp.Response) (err error) {
		if allow := string(resp.Header.Peek("Allow")); allow != "POST, GET, OPTIONS" && allow != "GET, POST, OPTIONS" {
			err = fmt.Errorf("unexpected Allow header value: %s", allow)
		}
		return
	})
	if err = test5.Run(); err != nil {
		t.Error(err)
	}

	// custom handler
	var custom bool
	router.OPTIONS("/path", func(_ *Context) {
		custom = true
	})

	// test again
	// * (server)
	test6 := tests.New(srv)
	test6.Method = MethodOptions
	test6.Url = "*"
	test6.Expect().Status(fasthttp.StatusOK).Custom(func(resp fasthttp.Response) error {
		if allow := string(resp.Header.Peek("Allow")); allow != "POST, GET, OPTIONS" && allow != "GET, POST, OPTIONS" {
			return fmt.Errorf("unexpected Allow header value: %s", allow)
		}
		if custom {
			return errors.New("custom handler called on *")
		}
		return nil
	})
	if err = test6.Run(); err != nil {
		t.Error(err)
	}

	// path
	test7 := tests.New(srv)
	test7.Method = MethodOptions
	test7.Url = "/path"
	test7.Expect().Status(fasthttp.StatusOK).Custom(func(resp fasthttp.Response) (err error) {
		if !custom {
			err = errors.New("custom handler not called")
		}
		return
	})
	if err = test7.Run(); err != nil {
		t.Error(err)
	}
}

func TestRouterNotAllowed(t *testing.T) {
	handlerFunc := func(_ *Context) {}

	router := NewRouter()
	router.POST("/path", handlerFunc)

	// Test not allowed
	srv := New("", router.Handler())

	var err error

	test1 := tests.New(srv)
	test1.Url = "/path"
	test1.Expect().Status(fasthttp.StatusMethodNotAllowed).Custom(func(resp fasthttp.Response) (err error) {
		if allow := string(resp.Header.Peek("Allow")); allow != "POST, OPTIONS" {
			err = errors.New("unexpected Allow header value: " + allow)
		}
		return
	})
	if err = test1.Run(); err != nil {
		t.Error(err)
	}

	// add another method
	router.DELETE("/path", handlerFunc)
	router.OPTIONS("/path", handlerFunc) // must be ignored

	// test again
	test2 := tests.New(srv)
	test2.Url = "/path"
	test2.Expect().Status(fasthttp.StatusMethodNotAllowed).Custom(func(resp fasthttp.Response) (err error) {
		if allow := string(resp.Header.Peek("Allow")); allow != "POST, DELETE, OPTIONS" && allow != "DELETE, POST, OPTIONS" {
			err = errors.New("unexpected Allow header value: " + allow)
		}
		return
	})
	if err = test2.Run(); err != nil {
		t.Error(err)
	}

	responseText := "custom method"
	router.MethodNotAllowed = func(ctx *Context) {
		ctx.SetStatusCode(fasthttp.StatusTeapot)
		ctx.Write([]byte(responseText))
	}

	test3 := tests.New(srv)
	test3.Url = "/path"
	test3.Expect().Status(fasthttp.StatusTeapot).Custom(func(resp fasthttp.Response) error {
		if allow := string(resp.Header.Peek("Allow")); allow != "POST, DELETE, OPTIONS" && allow != "DELETE, POST, OPTIONS" {
			return errors.New("unexpected Allow header value: " + allow)
		}
		if !bytes.Equal(resp.Body(), []byte(responseText)) {
			return fmt.Errorf("unexpected response got %q want %q", resp.Body(), responseText)
		}
		return nil
	})
	if err = test3.Run(); err != nil {
		t.Error(err)
	}
}

func TestConvert(t *testing.T) {
	router := NewRouter()

	respText := "Hello world"

	fastHandler := func(ctx *fasthttp.RequestCtx) {
		ctx.SetBodyString(respText)
	}

	router.GET("/", Convert(fastHandler))

	srv := New("", router.Handler())

	var err error

	test1 := tests.New(srv)
	test1.Expect().Status(fasthttp.StatusOK).Custom(func(resp fasthttp.Response) (err error) {
		if !strings.EqualFold(respText, string(resp.Body())) {
			err = fmt.Errorf("Expected response body %q, got %q", respText, resp.Body())
		}
		return nil
	})
	if err = test1.Run(); err != nil {
		t.Error(err)
	}
}

func TestRouterNotFound(t *testing.T) {
	handlerFunc := func(_ *Context) {}

	router := NewRouter()
	router.GET("/path", handlerFunc)
	router.GET("/dir/", handlerFunc)
	router.GET("/", handlerFunc)

	testRoutes := []struct {
		route string
		code  int
	}{
		{"/path/", 301},   // TSR -/
		{"/dir", 301},     // TSR +/
		{"/", 200},        // TSR +/
		{"/PATH", 301},    // Fixed Case
		{"/DIR", 301},     // Fixed Case
		{"/PATH/", 301},   // Fixed Case -/
		{"/DIR/", 301},    // Fixed Case +/
		{"/../path", 200}, // CleanPath
		{"/nope", 404},    // NotFound
	}

	srv := New("", router.Handler())

	var err error

	for _, tr := range testRoutes {
		test := tests.New(srv)
		test.Url = tr.route
		test.Expect().Status(tr.code)
		if err = test.Run(); err != nil {
			t.Error(err)
		}
	}

	// Test custom not found handler
	var notFound bool
	router.NotFound = func(ctx *Context) {
		ctx.SetStatusCode(404)
		notFound = true
	}
	test1 := tests.New(srv)
	test1.Url = "/nope"
	test1.Expect().Custom(func(resp fasthttp.Response) error {
		if !(resp.Header.StatusCode() == 404 && notFound == true) {
			return fmt.Errorf("Custom NotFound handler failed: Code=%d, Header=%v", resp.Header.StatusCode(), string(resp.Header.Peek("Location")))
		}
		return nil
	})
	if err = test1.Run(); err != nil {
		t.Error(err)
	}

	// Test other method than GET (want 307 instead of 301)
	router.PATCH("/path", handlerFunc)
	test2 := tests.New(srv)
	test2.Method = MethodPatch
	test2.Url = "/path/"
	test2.Expect().Custom(func(resp fasthttp.Response) error {
		if !(resp.Header.StatusCode() == fasthttp.StatusTemporaryRedirect) {
			return fmt.Errorf("Custom NotFound handler failed: Code=%d, Header=%v", resp.Header.StatusCode(), string(resp.Header.Peek("Location")))
		}
		return nil
	})
	if err = test2.Run(); err != nil {
		t.Error(err)
	}

	// Test special case where no node for the prefix "/" exists
	router = NewRouter()
	router.GET("/a", handlerFunc)
	srv = New("", router.Handler())

	test3 := tests.New(srv)
	test3.Url = "/"
	test3.Expect().Custom(func(resp fasthttp.Response) error {
		if !(resp.Header.StatusCode() == 404) {
			return fmt.Errorf("NotFound handling route / failed: Code=%d", resp.Header.StatusCode())
		}
		return nil
	})
	if err = test3.Run(); err != nil {
		t.Error(err)
	}
}

func TestRouterPanicHandler(t *testing.T) {
	panicHandled := false

	router := NewRouter()
	router.PanicHandler = func(ctx *Context, p interface{}) {
		panicHandled = true
	}

	router.Handle("PUT", "/user/:name", func(_ *Context) {
		panic("oops!")
	})

	defer func() {
		if rcv := recover(); rcv != nil {
			t.Fatal("handling panic failed")
		}
	}()

	srv := New("", router.Handler())

	var err error

	test1 := tests.New(srv)
	test1.Method = "PUT"
	test1.Url = "/user/gopher"
	test1.Expect().Custom(func(_ fasthttp.Response) error {
		if !panicHandled {
			return fmt.Errorf("simulating failed")
		}
		return nil
	})
	if err = test1.Run(); err != nil {
		t.Error(err)
	}
}

func TestRouterLookup(t *testing.T) {
	routed := false
	wantHandle := func(_ *Context) {
		routed = true
	}

	ctx := &Context{
		RequestCtx: &fasthttp.RequestCtx{},
	}

	router := NewRouter()
	// try empty router first
	handle, tsr := router.Lookup("GET", "/nope", ctx)
	if handle != nil {
		t.Fatalf("Got handle for unregistered pattern: %v", handle)
	}
	if tsr {
		t.Error("Got wrong TSR recommendation!")
	}

	// insert route and try again
	router.GET("/user/:name", wantHandle)

	handle, tsr = router.Lookup("GET", "/user/gopher", ctx)
	if handle == nil {
		t.Fatal("Got no handle!")
	} else {
		handle(nil)
		if !routed {
			t.Fatal("Routing failed!")
		}
	}

	if !reflect.DeepEqual(ctx.UserValue("name"), "gopher") {
		t.Fatalf("Wrong parameter values: want %v, got %v", "gopher", ctx.UserValue("name"))
	}

	handle, tsr = router.Lookup("GET", "/user/gopher/", ctx)
	if handle != nil {
		t.Fatalf("Got handle for unregistered pattern: %v", handle)
	}
	if !tsr {
		t.Error("Got no TSR recommendation!")
	}

	handle, tsr = router.Lookup("GET", "/nope", ctx)
	if handle != nil {
		t.Fatalf("Got handle for unregistered pattern: %v", handle)
	}
	if tsr {
		t.Error("Got wrong TSR recommendation!")
	}
}

type testMiddleware struct {
	key string
	val string
}

func (m testMiddleware) Handle(next Handler) Handler {
	return HandlerFunc(func(ctx *Context) {
		ctx.Response.Header.Set(m.key, m.val)

		next.Handle(ctx)
	})
}

func TestMiddleware(t *testing.T) {
	m := testMiddleware{key: "Test-Middleware", val: "Test"}

	router := NewRouter()
	router.Use(m)

	router.GET("/", func(ctx *Context) {
		ctx.Write([]byte("Hello"))
	})

	srv := New("", router.Handler())

	var err error

	test1 := tests.New(srv)
	test1.Expect().Status(fasthttp.StatusOK).Custom(func(resp fasthttp.Response) error {
		if !strings.EqualFold(m.val, string(resp.Header.Peek(m.key))) {
			return fmt.Errorf("Expected Test-Middleware value %q, got %q", m.val, resp.Header.Peek(m.key))
		}
		return nil
	})
	if err = test1.Run(); err != nil {
		t.Error(err)
	}

	// register middleware through HandlerConfig
	router2 := NewRouter()

	router2.GET("/", func(ctx *Context) {
		ctx.Write([]byte("Hello"))
	}, HandlerConfig{Middlewares: []Middleware{m}})

	srv = New("", router2.Handler())
	test2 := tests.New(srv)
	test2.Expect().Status(fasthttp.StatusOK).Custom(func(resp fasthttp.Response) error {
		if !strings.EqualFold(m.val, string(resp.Header.Peek(m.key))) {
			return fmt.Errorf("Expected Test-Middleware value %q, got %q", m.val, resp.Header.Peek(m.key))
		}
		return nil
	})
	if err = test2.Run(); err != nil {
		t.Error(err)
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
	recv := catchPanic(func() {
		router.ServeFiles("/noFilepath", os.TempDir())
	})
	if recv == nil {
		t.Fatal("registering path not ending with '*filepath' did not panic")
	}
	body := []byte("fake ico")
	ioutil.WriteFile(os.TempDir()+"/favicon.ico", body, 0644)

	router.ServeFiles("/*filepath", os.TempDir())

	srv := New("", router.Handler())

	var err error

	test1 := tests.New(srv)
	test1.Url = "/favicon.ico"
	test1.Timeout = 500 * time.Millisecond
	test1.Expect().Status(fasthttp.StatusOK).Custom(func(resp fasthttp.Response) error {
		if !bytes.Equal(resp.Body(), body) {
			t.Fatalf("Unexpected body %q. Expected %q", resp.Body(), body)
		}
		return nil
	})
	if err = test1.Run(); err != nil {
		t.Error(err)
	}
}

var cleanTests = []struct {
	path, result string
}{
	// Already clean
	{"/", "/"},
	{"/abc", "/abc"},
	{"/a/b/c", "/a/b/c"},
	{"/abc/", "/abc/"},
	{"/a/b/c/", "/a/b/c/"},

	// missing root
	{"", "/"},
	{"abc", "/abc"},
	{"abc/def", "/abc/def"},
	{"a/b/c", "/a/b/c"},

	// Remove doubled slash
	{"//", "/"},
	{"/abc//", "/abc/"},
	{"/abc/def//", "/abc/def/"},
	{"/a/b/c//", "/a/b/c/"},
	{"/abc//def//ghi", "/abc/def/ghi"},
	{"//abc", "/abc"},
	{"///abc", "/abc"},
	{"//abc//", "/abc/"},

	// Remove . elements
	{".", "/"},
	{"./", "/"},
	{"/abc/./def", "/abc/def"},
	{"/./abc/def", "/abc/def"},
	{"/abc/.", "/abc/"},

	// Remove .. elements
	{"..", "/"},
	{"../", "/"},
	{"../../", "/"},
	{"../..", "/"},
	{"../../abc", "/abc"},
	{"/abc/def/ghi/../jkl", "/abc/def/jkl"},
	{"/abc/def/../ghi/../jkl", "/abc/jkl"},
	{"/abc/def/..", "/abc"},
	{"/abc/def/../..", "/"},
	{"/abc/def/../../..", "/"},
	{"/abc/def/../../..", "/"},
	{"/abc/def/../../../ghi/jkl/../../../mno", "/mno"},

	// Combinations
	{"abc/./../def", "/def"},
	{"abc//./../def", "/def"},
	{"abc/../../././../def", "/def"},
}

func TestPathClean(t *testing.T) {
	for _, test := range cleanTests {
		if s := CleanPath(test.path); s != test.result {
			t.Errorf("CleanPath(%q) = %q, want %q", test.path, s, test.result)
		}
		if s := CleanPath(test.result); s != test.result {
			t.Errorf("CleanPath(%q) = %q, want %q", test.result, s, test.result)
		}
	}
}

func TestPathCleanMallocs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	if runtime.GOMAXPROCS(0) > 1 {
		t.Log("skipping AllocsPerRun checks; GOMAXPROCS>1")
		return
	}

	for _, test := range cleanTests {
		allocs := testing.AllocsPerRun(100, func() {
			CleanPath(test.result)
		})
		if allocs > 0 {
			t.Errorf("CleanPath(%q): %v allocs, want zero", test.result, allocs)
		}
	}
}

func printChildren(n *node, prefix string) {
	fmt.Printf(" %02d:%02d %s%s[%d] %v %t %d \r\n", n.priority, n.maxParams, prefix, n.path, len(n.children), n.handle, n.wildChild, n.nType)
	for l := len(n.path); l > 0; l-- {
		prefix += " "
	}
	for _, child := range n.children {
		printChildren(child, prefix)
	}
}

// Used as a workaround since we can't compare functions or their addresses
var fakeHandlerValue string

func fakeHandler(val string) HandlerFunc {
	return func(*Context) {
		fakeHandlerValue = val
	}
}

type Params []Param

type Param struct {
	Key   string
	Value string
}

func (ps Params) String(name string) string {
	for _, p := range ps {
		if p.Key == name {
			return p.Value
		}
	}
	return ""
}

type testRequests []struct {
	path       string
	nilHandler bool
	route      string
	ps         Params
}

func acquarieContext(path string) *Context {
	fastRequest := fasthttp.Request{}
	fastRequest.SetRequestURI(path)
	return &Context{
		RequestCtx: &fasthttp.RequestCtx{Request: fastRequest},
	}
}

func checkRequests(t *testing.T, tree *node, requests testRequests) {
	for _, request := range requests {
		requestCtx := acquarieContext(request.path)
		handler, _ := tree.getValue(request.path, requestCtx)

		if handler == nil {
			if !request.nilHandler {
				t.Errorf("handle mismatch for route '%s': Expected non-nil handle", request.path)
			}
		} else if request.nilHandler {
			t.Errorf("handle mismatch for route '%s': Expected nil handle", request.path)
		} else {
			handler(nil)
			if fakeHandlerValue != request.route {
				t.Errorf("handle mismatch for route '%s': Wrong handle (%s != %s)", request.path, fakeHandlerValue, request.route)
			}
		}

		for _, p := range request.ps {
			if requestCtx.UserValue(p.Key) != p.Value {
				t.Errorf(" mismatch for route '%s'", request.path)
			}
		}
	}
}

func checkPriorities(t *testing.T, n *node) uint32 {
	var prio uint32
	for i := range n.children {
		prio += checkPriorities(t, n.children[i])
	}

	if n.handle != nil {
		prio++
	}

	if n.priority != prio {
		t.Errorf(
			"priority mismatch for node '%s': is %d, should be %d",
			n.path, n.priority, prio,
		)
	}

	return prio
}

func checkMaxParams(t *testing.T, n *node) uint8 {
	var maxParams uint8
	for i := range n.children {
		params := checkMaxParams(t, n.children[i])
		if params > maxParams {
			maxParams = params
		}
	}
	if n.nType > root && !n.wildChild {
		maxParams++
	}

	if n.maxParams != maxParams {
		t.Errorf(
			"maxParams mismatch for node '%s': is %d, should be %d",
			n.path, n.maxParams, maxParams,
		)
	}

	return maxParams
}

func TestCountParams(t *testing.T) {
	if countParams("/path/:param1/static/*catch-all") != 2 {
		t.Fail()
	}
	if countParams(strings.Repeat("/:param", 256)) != 255 {
		t.Fail()
	}
}

func TestTreeAddAndGet(t *testing.T) {
	tree := &node{}

	routes := [...]string{
		"/hi",
		"/contact",
		"/co",
		"/c",
		"/a",
		"/ab",
		"/doc/",
		"/doc/go_faq.html",
		"/doc/go1.html",
		"/α",
		"/β",
	}
	for _, route := range routes {
		tree.addRoute(route, fakeHandler(route))
	}

	//printChildren(tree, "")

	checkRequests(t, tree, testRequests{
		{"/a", false, "/a", nil},
		{"/", true, "", nil},
		{"/hi", false, "/hi", nil},
		{"/contact", false, "/contact", nil},
		{"/co", false, "/co", nil},
		{"/con", true, "", nil},  // key mismatch
		{"/cona", true, "", nil}, // key mismatch
		{"/no", true, "", nil},   // no matching child
		{"/ab", false, "/ab", nil},
		{"/α", false, "/α", nil},
		{"/β", false, "/β", nil},
	})

	checkPriorities(t, tree)
	checkMaxParams(t, tree)
}

func TestTreeWildcard(t *testing.T) {
	tree := &node{}

	routes := [...]string{
		"/",
		"/cmd/:tool/:sub",
		"/cmd/:tool/",
		"/src/*filepath",
		"/search/",
		"/search/:query",
		"/user_:name",
		"/user_:name/about",
		"/files/:dir/*filepath",
		"/doc/",
		"/doc/go_faq.html",
		"/doc/go1.html",
		"/info/:user/public",
		"/info/:user/project/:project",
	}
	for _, route := range routes {
		tree.addRoute(route, fakeHandler(route))
	}

	//printChildren(tree, "")

	checkRequests(t, tree, testRequests{
		{"/", false, "/", nil},
		{"/cmd/test/", false, "/cmd/:tool/", Params{Param{"tool", "test"}}},
		{"/cmd/test", true, "", Params{Param{"tool", "test"}}},
		{"/cmd/test/3", false, "/cmd/:tool/:sub", Params{Param{"tool", "test"}, Param{"sub", "3"}}},
		{"/src/", false, "/src/*filepath", Params{Param{"filepath", "/"}}},
		{"/src/some/file.png", false, "/src/*filepath", Params{Param{"filepath", "/some/file.png"}}},
		{"/search/", false, "/search/", nil},
		{"/search/someth!ng+in+ünìcodé", false, "/search/:query", Params{Param{"query", "someth!ng+in+ünìcodé"}}},
		{"/search/someth!ng+in+ünìcodé/", true, "", Params{Param{"query", "someth!ng+in+ünìcodé"}}},
		{"/user_gopher", false, "/user_:name", Params{Param{"name", "gopher"}}},
		{"/user_gopher/about", false, "/user_:name/about", Params{Param{"name", "gopher"}}},
		{"/files/js/inc/framework.js", false, "/files/:dir/*filepath", Params{Param{"dir", "js"}, Param{"filepath", "/inc/framework.js"}}},
		{"/info/gordon/public", false, "/info/:user/public", Params{Param{"user", "gordon"}}},
		{"/info/gordon/project/go", false, "/info/:user/project/:project", Params{Param{"user", "gordon"}, Param{"project", "go"}}},
	})

	checkPriorities(t, tree)
	checkMaxParams(t, tree)
}

type testRoute struct {
	path     string
	conflict bool
}

func testRoutes(t *testing.T, routes []testRoute) {
	tree := &node{}

	for _, route := range routes {
		recv := catchPanic(func() {
			tree.addRoute(route.path, nil)
		})

		if route.conflict {
			if recv == nil {
				t.Errorf("no panic for conflicting route '%s'", route.path)
			}
		} else if recv != nil {
			t.Errorf("unexpected panic for route '%s': %v", route.path, recv)
		}
	}

	//printChildren(tree, "")
}

func TestTreeWildcardConflict(t *testing.T) {
	routes := []testRoute{
		{"/cmd/:tool/:sub", false},
		{"/cmd/vet", true},
		{"/src/*filepath", false},
		{"/src/*filepathx", true},
		{"/src/", true},
		{"/src1/", false},
		{"/src1/*filepath", true},
		{"/src2*filepath", true},
		{"/search/:query", false},
		{"/search/invalid", true},
		{"/user_:name", false},
		{"/user_x", true},
		{"/user_:name", false},
		{"/id:id", false},
		{"/id/:id", true},
	}
	testRoutes(t, routes)
}

func TestTreeChildConflict(t *testing.T) {
	routes := []testRoute{
		{"/cmd/vet", false},
		{"/cmd/:tool/:sub", true},
		{"/src/AUTHORS", false},
		{"/src/*filepath", true},
		{"/user_x", false},
		{"/user_:name", true},
		{"/id/:id", false},
		{"/id:id", true},
		{"/:id", true},
		{"/*filepath", true},
	}
	testRoutes(t, routes)
}

func TestTreeDupliatePath(t *testing.T) {
	tree := &node{}

	routes := [...]string{
		"/",
		"/doc/",
		"/src/*filepath",
		"/search/:query",
		"/user_:name",
	}
	for _, route := range routes {
		recv := catchPanic(func() {
			tree.addRoute(route, fakeHandler(route))
		})
		if recv != nil {
			t.Fatalf("panic inserting route '%s': %v", route, recv)
		}

		// Add again
		recv = catchPanic(func() {
			tree.addRoute(route, nil)
		})
		if recv == nil {
			t.Fatalf("no panic while inserting duplicate route '%s", route)
		}
	}

	//printChildren(tree, "")

	checkRequests(t, tree, testRequests{
		{"/", false, "/", nil},
		{"/doc/", false, "/doc/", nil},
		{"/src/some/file.png", false, "/src/*filepath", Params{Param{"filepath", "/some/file.png"}}},
		{"/search/someth!ng+in+ünìcodé", false, "/search/:query", Params{Param{"query", "someth!ng+in+ünìcodé"}}},
		{"/user_gopher", false, "/user_:name", Params{Param{"name", "gopher"}}},
	})
}

func TestEmptyWildcardName(t *testing.T) {
	tree := &node{}

	routes := [...]string{
		"/user:",
		"/user:/",
		"/cmd/:/",
		"/src/*",
	}
	for _, route := range routes {
		recv := catchPanic(func() {
			tree.addRoute(route, nil)
		})
		if recv == nil {
			t.Fatalf("no panic while inserting route with empty wildcard name '%s", route)
		}
	}
}

func TestTreeCatchAllConflict(t *testing.T) {
	routes := []testRoute{
		{"/src/*filepath/x", true},
		{"/src2/", false},
		{"/src2/*filepath/x", true},
	}
	testRoutes(t, routes)
}

func TestTreeCatchAllConflictRoot(t *testing.T) {
	routes := []testRoute{
		{"/", false},
		{"/*filepath", true},
	}
	testRoutes(t, routes)
}

func TestTreeDoubleWildcard(t *testing.T) {
	const panicMsg = "only one wildcard per path segment is allowed"

	routes := [...]string{
		"/:foo:bar",
		"/:foo:bar/",
		"/:foo*bar",
	}

	for _, route := range routes {
		tree := &node{}
		recv := catchPanic(func() {
			tree.addRoute(route, nil)
		})

		if rs, ok := recv.(string); !ok || !strings.HasPrefix(rs, panicMsg) {
			t.Fatalf(`"Expected panic "%s" for route '%s', got "%v"`, panicMsg, route, recv)
		}
	}
}

func TestTreeTrailingSlashRedirect(t *testing.T) {
	tree := &node{}
	c := &Context{
		RequestCtx: &fasthttp.RequestCtx{},
	}

	routes := [...]string{
		"/hi",
		"/b/",
		"/search/:query",
		"/cmd/:tool/",
		"/src/*filepath",
		"/x",
		"/x/y",
		"/y/",
		"/y/z",
		"/0/:id",
		"/0/:id/1",
		"/1/:id/",
		"/1/:id/2",
		"/aa",
		"/a/",
		"/admin",
		"/admin/:category",
		"/admin/:category/:page",
		"/doc",
		"/doc/go_faq.html",
		"/doc/go1.html",
		"/no/a",
		"/no/b",
		"/api/hello/:name",
	}
	for _, route := range routes {
		recv := catchPanic(func() {
			tree.addRoute(route, fakeHandler(route))
		})
		if recv != nil {
			t.Fatalf("panic inserting route '%s': %v", route, recv)
		}
	}

	//printChildren(tree, "")

	tsrRoutes := [...]string{
		"/hi/",
		"/b",
		"/search/gopher/",
		"/cmd/vet",
		"/src",
		"/x/",
		"/y",
		"/0/go/",
		"/1/go",
		"/a",
		"/admin/",
		"/admin/config/",
		"/admin/config/permissions/",
		"/doc/",
	}
	for _, route := range tsrRoutes {
		handler, tsr := tree.getValue(route, c)
		if handler != nil {
			t.Fatalf("non-nil handler for TSR route '%s", route)
		} else if !tsr {
			t.Errorf("expected TSR recommendation for route '%s'", route)
		}
	}

	noTsrRoutes := [...]string{
		"/",
		"/no",
		"/no/",
		"/_",
		"/_/",
		"/api/world/abc",
	}
	for _, route := range noTsrRoutes {
		handler, tsr := tree.getValue(route, c)
		if handler != nil {
			t.Fatalf("non-nil handler for No-TSR route '%s", route)
		} else if tsr {
			t.Errorf("expected no TSR recommendation for route '%s'", route)
		}
	}
}

func TestTreeRootTrailingSlashRedirect(t *testing.T) {
	tree := &node{}
	c := &Context{
		RequestCtx: &fasthttp.RequestCtx{},
	}

	recv := catchPanic(func() {
		tree.addRoute("/:test", fakeHandler("/:test"))
	})
	if recv != nil {
		t.Fatalf("panic inserting test route: %v", recv)
	}

	handler, tsr := tree.getValue("/", c)
	if handler != nil {
		t.Fatalf("non-nil handler")
	} else if tsr {
		t.Errorf("expected no TSR recommendation")
	}
}

func TestTreeFindCaseInsensitivePath(t *testing.T) {
	tree := &node{}

	routes := [...]string{
		"/hi",
		"/b/",
		"/ABC/",
		"/search/:query",
		"/cmd/:tool/",
		"/src/*filepath",
		"/x",
		"/x/y",
		"/y/",
		"/y/z",
		"/0/:id",
		"/0/:id/1",
		"/1/:id/",
		"/1/:id/2",
		"/aa",
		"/a/",
		"/doc",
		"/doc/go_faq.html",
		"/doc/go1.html",
		"/doc/go/away",
		"/no/a",
		"/no/b",
		"/Π",
		"/u/apfêl/",
		"/u/äpfêl/",
		"/u/öpfêl",
		"/v/Äpfêl/",
		"/v/Öpfêl",
		"/w/♬",  // 3 byte
		"/w/♭/", // 3 byte, last byte differs
		"/w/𠜎",  // 4 byte
		"/w/𠜏/", // 4 byte
	}

	for _, route := range routes {
		recv := catchPanic(func() {
			tree.addRoute(route, fakeHandler(route))
		})
		if recv != nil {
			t.Fatalf("panic inserting route '%s': %v", route, recv)
		}
	}

	// Check out == in for all registered routes
	// With fixTrailingSlash = true
	for _, route := range routes {
		out, found := tree.findCaseInsensitivePath(route, true)
		if !found {
			t.Errorf("Route '%s' not found!", route)
		} else if string(out) != route {
			t.Errorf("Wrong result for route '%s': %s", route, string(out))
		}
	}
	// With fixTrailingSlash = false
	for _, route := range routes {
		out, found := tree.findCaseInsensitivePath(route, false)
		if !found {
			t.Errorf("Route '%s' not found!", route)
		} else if string(out) != route {
			t.Errorf("Wrong result for route '%s': %s", route, string(out))
		}
	}

	tests := []struct {
		in    string
		out   string
		found bool
		slash bool
	}{
		{"/HI", "/hi", true, false},
		{"/HI/", "/hi", true, true},
		{"/B", "/b/", true, true},
		{"/B/", "/b/", true, false},
		{"/abc", "/ABC/", true, true},
		{"/abc/", "/ABC/", true, false},
		{"/aBc", "/ABC/", true, true},
		{"/aBc/", "/ABC/", true, false},
		{"/abC", "/ABC/", true, true},
		{"/abC/", "/ABC/", true, false},
		{"/SEARCH/QUERY", "/search/QUERY", true, false},
		{"/SEARCH/QUERY/", "/search/QUERY", true, true},
		{"/CMD/TOOL/", "/cmd/TOOL/", true, false},
		{"/CMD/TOOL", "/cmd/TOOL/", true, true},
		{"/SRC/FILE/PATH", "/src/FILE/PATH", true, false},
		{"/x/Y", "/x/y", true, false},
		{"/x/Y/", "/x/y", true, true},
		{"/X/y", "/x/y", true, false},
		{"/X/y/", "/x/y", true, true},
		{"/X/Y", "/x/y", true, false},
		{"/X/Y/", "/x/y", true, true},
		{"/Y/", "/y/", true, false},
		{"/Y", "/y/", true, true},
		{"/Y/z", "/y/z", true, false},
		{"/Y/z/", "/y/z", true, true},
		{"/Y/Z", "/y/z", true, false},
		{"/Y/Z/", "/y/z", true, true},
		{"/y/Z", "/y/z", true, false},
		{"/y/Z/", "/y/z", true, true},
		{"/Aa", "/aa", true, false},
		{"/Aa/", "/aa", true, true},
		{"/AA", "/aa", true, false},
		{"/AA/", "/aa", true, true},
		{"/aA", "/aa", true, false},
		{"/aA/", "/aa", true, true},
		{"/A/", "/a/", true, false},
		{"/A", "/a/", true, true},
		{"/DOC", "/doc", true, false},
		{"/DOC/", "/doc", true, true},
		{"/NO", "", false, true},
		{"/DOC/GO", "", false, true},
		{"/π", "/Π", true, false},
		{"/π/", "/Π", true, true},
		{"/u/ÄPFÊL/", "/u/äpfêl/", true, false},
		{"/u/ÄPFÊL", "/u/äpfêl/", true, true},
		{"/u/ÖPFÊL/", "/u/öpfêl", true, true},
		{"/u/ÖPFÊL", "/u/öpfêl", true, false},
		{"/v/äpfêL/", "/v/Äpfêl/", true, false},
		{"/v/äpfêL", "/v/Äpfêl/", true, true},
		{"/v/öpfêL/", "/v/Öpfêl", true, true},
		{"/v/öpfêL", "/v/Öpfêl", true, false},
		{"/w/♬/", "/w/♬", true, true},
		{"/w/♭", "/w/♭/", true, true},
		{"/w/𠜎/", "/w/𠜎", true, true},
		{"/w/𠜏", "/w/𠜏/", true, true},
	}
	// With fixTrailingSlash = true
	for _, test := range tests {
		out, found := tree.findCaseInsensitivePath(test.in, true)
		if found != test.found || (found && (string(out) != test.out)) {
			t.Errorf("Wrong result for '%s': got %s, %t; want %s, %t",
				test.in, string(out), found, test.out, test.found)
			return
		}
	}
	// With fixTrailingSlash = false
	for _, test := range tests {
		out, found := tree.findCaseInsensitivePath(test.in, false)
		if test.slash {
			if found {
				// test needs a trailingSlash fix. It must not be found!
				t.Errorf("Found without fixTrailingSlash: %s; got %s", test.in, string(out))
			}
		} else {
			if found != test.found || (found && (string(out) != test.out)) {
				t.Errorf("Wrong result for '%s': got %s, %t; want %s, %t",
					test.in, string(out), found, test.out, test.found)
				return
			}
		}
	}
}

func TestTreeInvalidNodeType(t *testing.T) {
	const panicMsg = "invalid node type"

	tree := &node{}
	c := &Context{
		RequestCtx: &fasthttp.RequestCtx{},
	}
	tree.addRoute("/", fakeHandler("/"))
	tree.addRoute("/:page", fakeHandler("/:page"))

	// set invalid node type
	tree.children[0].nType = 42

	// normal lookup
	recv := catchPanic(func() {
		tree.getValue("/test", c)
	})
	if rs, ok := recv.(string); !ok || rs != panicMsg {
		t.Fatalf("Expected panic '"+panicMsg+"', got '%v'", recv)
	}

	// case-insensitive lookup
	recv = catchPanic(func() {
		tree.findCaseInsensitivePath("/test", true)
	})
	if rs, ok := recv.(string); !ok || rs != panicMsg {
		t.Fatalf("Expected panic '"+panicMsg+"', got '%v'", recv)
	}
}

type handlerStruct struct {
	handeled *bool
}

func (h handlerStruct) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	*h.handeled = true
}

func catchPanic(testFunc func()) (recv interface{}) {
	defer func() {
		recv = recover()
	}()

	testFunc()
	return
}
