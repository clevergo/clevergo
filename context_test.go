// Copyright 2020 CleverGo. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package clevergo

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContext_SetContentType(t *testing.T) {
	tests := []string{
		"text/html",
		"text/html; charset=utf-8",
		"text/plain",
		"text/plain; charset=utf-8",
		"application/json",
		"application/xml",
	}

	for _, test := range tests {
		ctx := newContext(httptest.NewRecorder(), nil)
		ctx.SetContentType(test)
		if ctx.Response.Header().Get("Content-Type") != test {
			t.Errorf("expected content type %q, got %q", test, ctx.Response.Header().Get("Content-Type"))
		}
	}
}

func TestContext_SetContentTypeHTML(t *testing.T) {
	ctx := newContext(httptest.NewRecorder(), nil)
	ctx.SetContentTypeHTML()
	if ctx.Response.Header().Get("Content-Type") != "text/html; charset=utf-8" {
		t.Errorf("expected content type %q, got %q", "text/html; charset=utf-8", ctx.Response.Header().Get("Content-Type"))
	}
}
func TestContext_SetContentTypeText(t *testing.T) {
	ctx := newContext(httptest.NewRecorder(), nil)
	ctx.SetContentTypeText()
	if ctx.Response.Header().Get("Content-Type") != "text/plain; charset=utf-8" {
		t.Errorf("expected content type %q, got %q", "text/plain; charset=utf-8", ctx.Response.Header().Get("Content-Type"))
	}
}
func TestContext_SetContentTypeJSON(t *testing.T) {
	ctx := newContext(httptest.NewRecorder(), nil)
	ctx.SetContentTypeJSON()
	assert.Equal(t, "application/json; charset=utf-8", ctx.Response.Header().Get("Content-Type"), "content type does not match")
}
func TestContext_SetContentTypeXML(t *testing.T) {
	ctx := newContext(httptest.NewRecorder(), nil)
	ctx.SetContentTypeXML()
	assert.Equal(t, "application/xml; charset=utf-8", ctx.Response.Header().Get("Content-Type"), "content type does not match")
}

func TestContext_Write(t *testing.T) {
	tests := [][]byte{
		[]byte("foo"),
		[]byte("bar"),
	}

	for _, test := range tests {
		w := httptest.NewRecorder()
		ctx := newContext(w, nil)
		ctx.Write(test)
		if !bytes.Equal(w.Body.Bytes(), test) {
			t.Errorf("expected body %q, got %q", test, w.Body.Bytes())
		}
	}
}
func TestContext_WriteString(t *testing.T) {
	tests := []string{
		"foo",
		"bar",
	}

	for _, test := range tests {
		w := httptest.NewRecorder()
		ctx := newContext(w, nil)
		ctx.WriteString(test)
		if w.Body.String() != test {
			t.Errorf("expected body %q, got %q", test, w.Body.String())
		}
	}
}

func TestContext_NotFound(t *testing.T) {
	w := httptest.NewRecorder()
	ctx := newContext(w, nil)
	ctx.NotFound()
	if w.Code != http.StatusNotFound {
		t.Errorf("expected status code %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestContext_Redirect(t *testing.T) {
	w := httptest.NewRecorder()
	ctx := newContext(w, httptest.NewRequest(http.MethodGet, "/", nil))
	ctx.Redirect("/redirect", http.StatusPermanentRedirect)
	if w.Code != http.StatusPermanentRedirect {
		t.Errorf("expected status code %d, got %d", http.StatusPermanentRedirect, w.Code)
	}
}
func TestContext_Error(t *testing.T) {
	tests := []struct {
		msg  string
		code int
	}{
		{"foo", http.StatusInternalServerError},
		{"bar", http.StatusForbidden},
	}

	for _, test := range tests {
		w := httptest.NewRecorder()
		ctx := newContext(w, nil)
		ctx.Error(test.msg, test.code)
		if w.Body.String() != fmt.Sprintln(test.msg) {
			t.Errorf("expected body %q, got %q", fmt.Sprintln(test.msg), w.Body.String())
		}
		if w.Code != test.code {
			t.Errorf("expected status code %d, got %d", test.code, w.Code)
		}
	}
}

func TestContext_WithValue(t *testing.T) {
	values := map[interface{}]interface{}{
		"foo":  "bar",
		"fizz": "buzz",
		0:      0,
		1:      1,
		true:   true,
		false:  false,
	}

	ctx := newContext(nil, httptest.NewRequest(http.MethodGet, "/", nil))
	for key, val := range values {
		ctx.WithValue(key, val)
	}

	for key, val := range values {
		if !reflect.DeepEqual(val, ctx.Value(key)) {
			t.Errorf("expected the value of %v: %v, got %v", key, val, ctx.Value(key))
		}
	}
}

func TestIsMethod(t *testing.T) {
	tests := []struct {
		method string
		f      func(ctx *Context, method string) bool
	}{
		{http.MethodGet, func(ctx *Context, method string) bool {
			return ctx.IsGet()
		}},
		{http.MethodDelete, func(ctx *Context, method string) bool {
			return ctx.IsDelete()
		}},
		{http.MethodPatch, func(ctx *Context, method string) bool {
			return ctx.IsPatch()
		}},
		{http.MethodPost, func(ctx *Context, method string) bool {
			return ctx.IsPost()
		}},
		{http.MethodPut, func(ctx *Context, method string) bool {
			return ctx.IsPut()
		}},
		{http.MethodOptions, func(ctx *Context, method string) bool {
			return ctx.IsOptions()
		}},
		{http.MethodHead, func(ctx *Context, method string) bool {
			return ctx.IsMethod(method)
		}},
	}
	for _, test := range tests {
		ctx := newContext(nil, httptest.NewRequest(test.method, "/", nil))
		if !test.f(ctx, test.method) {
			t.Errorf("failed to determine request method")
		}
	}
}

func TestContext_Cookie(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "foo", Value: "bar"})
	ctx := newContext(nil, req)
	actual, _ := ctx.Cookie("foo")
	expected, _ := req.Cookie("foo")
	assert.Equal(t, actual, expected, "cookie does not match")
}

func TestContext_Cookies(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "foo", Value: "bar"})
	ctx := newContext(nil, req)
	assert.Equal(t, ctx.Cookies(), req.Cookies(), "cookies does not match")
}

func TestContext_SetCookie(t *testing.T) {
	w := httptest.NewRecorder()
	ctx := &Context{Response: w}
	cookie := &http.Cookie{Name: "foo", Value: "bar"}
	ctx.SetCookie(cookie)
	actual := w.Result().Cookies()[0]
	if cookie.Name != actual.Name {
		t.Errorf("expected cookie name %s, got %s", cookie.Name, actual.Name)
	}
	if cookie.Value != actual.Value {
		t.Errorf("expected cookie value %s, got %s", cookie.Value, actual.Value)
	}
}

func TestContext_WriteHeader(t *testing.T) {
	codes := []int{http.StatusOK, http.StatusForbidden, http.StatusInternalServerError, http.StatusUnauthorized}
	for _, code := range codes {
		w := httptest.NewRecorder()
		ctx := newContext(w, nil)
		ctx.WriteHeader(code)
		if w.Code != code {
			t.Errorf("expected status code %d, got %d", code, w.Code)
		}
	}
}

func TestContext_IsAJAX(t *testing.T) {
	tests := []struct {
		value    string
		expected bool
	}{
		{"", false},
		{"XMLHttpRequest", true},
	}

	for _, test := range tests {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Requested-With", test.value)
		ctx := newContext(nil, req)
		if ctx.IsAJAX() != test.expected {
			t.Errorf("expected IsAJAX %t, got %t", test.expected, ctx.IsAJAX())
		}
	}
}

func TestContext_GetHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("foo", "bar")
	ctx := newContext(nil, req)
	for _, name := range []string{"foo", "fizz"} {
		if req.Header.Get(name) != ctx.GetHeader(name) {
			t.Errorf("expected header %s: %q, got %q", name, req.Header.Get(name), ctx.GetHeader(name))
		}
	}
}

type testBody struct {
	Status  string      `json:"status" xml:"status"`
	Message string      `json:"message" xml:"message"`
	Data    interface{} `json:"data" xml:"data"`
}

func TestContext_JSON(t *testing.T) {
	tests := []struct {
		code      int
		data      interface{}
		body      interface{}
		shouldErr bool
	}{
		{
			200,
			testBody{"success", "created", "foobar"},
			`{"status":"success","message":"created","data":"foobar"}`,
			false,
		},
		{
			500,
			testBody{"error", "internal error", nil},
			`{"status":"error","message":"internal error","data":null}`,
			false,
		},
		{
			200,
			make(chan int),
			"",
			true,
		},
	}
	for _, test := range tests {
		w := httptest.NewRecorder()
		ctx := newContext(w, nil)
		err := ctx.JSON(test.code, test.data)
		if test.shouldErr {
			assert.NotNil(t, err)
			continue
		}
		assert.Equal(t, test.code, w.Code, "status code does not match")
		assert.Equal(t, w.Header().Get("Content-Type"), "application/json; charset=utf-8", "content type does not match")
		assert.Equal(t, w.Body.String(), test.body, "resposne body does not match")
	}
}

func TestContext_String(t *testing.T) {
	tests := []struct {
		code int
		s    string
	}{
		{200, "foobar"},
		{500, "error"},
	}
	for _, test := range tests {
		w := httptest.NewRecorder()
		ctx := newContext(w, nil)
		ctx.String(test.code, test.s)
		assert.Equal(t, test.code, w.Code, "status code does not match")
		assert.Equal(t, w.Header().Get("Content-Type"), "text/plain; charset=utf-8", "content type does not match")
		assert.Equal(t, w.Body.String(), test.s, "resposne body does not match")
	}
}

func TestContext_XML(t *testing.T) {
	tests := []struct {
		code      int
		data      interface{}
		body      interface{}
		shouldErr bool
	}{
		{
			200,
			testBody{"success", "created", "foobar"},
			`<testBody><status>success</status><message>created</message><data>foobar</data></testBody>`,
			false,
		},
		{
			500,
			testBody{"error", "internal error", nil},
			`<testBody><status>error</status><message>internal error</message></testBody>`,
			false,
		},
		{
			200,
			make(chan int),
			"",
			true,
		},
	}
	for _, test := range tests {
		w := httptest.NewRecorder()
		ctx := newContext(w, nil)
		err := ctx.XML(test.code, test.data)
		if test.shouldErr {
			assert.NotNil(t, err)
			continue
		}
		assert.Equal(t, test.code, w.Code, "status code does not match")
		assert.Equal(t, w.Header().Get("Content-Type"), "application/xml; charset=utf-8", "content type does not match")
		assert.Equal(t, w.Body.String(), test.body, "resposne body does not match")
	}
}

func TestContext_HTML(t *testing.T) {
	tests := []struct {
		code int
		s    string
	}{
		{200, "<html><body>foobar</body></html>"},
		{500, "<html><body>error</body></html>"},
	}
	for _, test := range tests {
		w := httptest.NewRecorder()
		ctx := newContext(w, nil)
		ctx.HTML(test.code, test.s)
		assert.Equal(t, test.code, w.Code, "status code does not match")
		assert.Equal(t, w.Header().Get("Content-Type"), "text/html; charset=utf-8", "content type does not match")
		assert.Equal(t, w.Body.String(), test.s, "resposne body does not match")
	}
}
