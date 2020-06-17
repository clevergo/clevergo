// Copyright 2020 CleverGo. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package clevergo

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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
		assert.Equal(t, test, ctx.Response.Header().Get("Content-Type"))
	}
}

func TestContext_SetContentTypeHTML(t *testing.T) {
	ctx := newContext(httptest.NewRecorder(), nil)
	ctx.SetContentTypeHTML()
	assert.Equal(t, "text/html; charset=utf-8", ctx.Response.Header().Get("Content-Type"))
}
func TestContext_SetContentTypeText(t *testing.T) {
	ctx := newContext(httptest.NewRecorder(), nil)
	ctx.SetContentTypeText()
	assert.Equal(t, "text/plain; charset=utf-8", ctx.Response.Header().Get("Content-Type"))
}
func TestContext_SetContentTypeJSON(t *testing.T) {
	ctx := newContext(httptest.NewRecorder(), nil)
	ctx.SetContentTypeJSON()
	assert.Equal(t, "application/json; charset=utf-8", ctx.Response.Header().Get("Content-Type"))
}
func TestContext_SetContentTypeXML(t *testing.T) {
	ctx := newContext(httptest.NewRecorder(), nil)
	ctx.SetContentTypeXML()
	assert.Equal(t, "application/xml; charset=utf-8", ctx.Response.Header().Get("Content-Type"))
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
		assert.Equal(t, string(test), w.Body.String())
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
		assert.Equal(t, test, w.Body.String())
	}
}

func TestContext_NotFound(t *testing.T) {
	w := httptest.NewRecorder()
	ctx := newContext(w, nil)
	assert.Nil(t, ctx.NotFound())
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestContext_Redirect(t *testing.T) {
	w := httptest.NewRecorder()
	ctx := newContext(w, httptest.NewRequest(http.MethodGet, "/", nil))
	assert.Nil(t, ctx.Redirect(http.StatusPermanentRedirect, "/redirect"))
	assert.Equal(t, http.StatusPermanentRedirect, w.Code)
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
		assert.Nil(t, ctx.Error(test.code, test.msg))
		assert.Equal(t, fmt.Sprintln(test.msg), w.Body.String())
		assert.Equal(t, test.code, w.Code)
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
		assert.Equal(t, val, ctx.Value(key))
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
		assert.True(t, test.f(ctx, test.method), "failed to determine request method")
	}
}

func TestContext_Cookie(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "foo", Value: "bar"})
	ctx := newContext(nil, req)
	actual, _ := ctx.Cookie("foo")
	expected, _ := req.Cookie("foo")
	assert.Equal(t, expected, actual)
}

func TestContext_Cookies(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "foo", Value: "bar"})
	ctx := newContext(nil, req)
	assert.Equal(t, req.Cookies(), ctx.Cookies())
}

func TestContext_SetCookie(t *testing.T) {
	w := httptest.NewRecorder()
	ctx := &Context{Response: w}
	cookie := &http.Cookie{Name: "foo", Value: "bar"}
	ctx.SetCookie(cookie)
	actual := w.Result().Cookies()[0]
	assert.Equal(t, cookie.Name, actual.Name)
	assert.Equal(t, cookie.Value, actual.Value)
}

func TestContext_WriteHeader(t *testing.T) {
	codes := []int{http.StatusOK, http.StatusForbidden, http.StatusInternalServerError, http.StatusUnauthorized}
	for _, code := range codes {
		w := httptest.NewRecorder()
		ctx := newContext(w, nil)
		ctx.WriteHeader(code)
		assert.Equal(t, code, w.Code)
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
		assert.Equal(t, test.expected, ctx.IsAJAX())
	}
}

func TestContext_GetHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("foo", "bar")
	ctx := newContext(nil, req)
	for _, name := range []string{"foo", "fizz"} {
		assert.Equal(t, req.Header.Get(name), ctx.GetHeader(name))
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
		assert.Equal(t, test.code, w.Code)
		assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))
		assert.Equal(t, test.body, w.Body.String())
	}
}

func TestContext_JSONBlob(t *testing.T) {
	tests := []struct {
		code int
		data string
	}{
		{
			200,
			`{"status":"success","message":"created","data":"foobar"}`,
		},
		{
			500,
			`{"status":"error","message":"internal error","data":null}`,
		},
	}
	for _, test := range tests {
		w := httptest.NewRecorder()
		ctx := newContext(w, nil)
		ctx.JSONBlob(test.code, []byte(test.data))
		assert.Equal(t, test.code, w.Code)
		assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))
		assert.Equal(t, test.data, w.Body.String())
	}
}

func TestContext_JSONP(t *testing.T) {
	tests := []struct {
		query       string
		contentType string
		code        int
		data        interface{}
		body        interface{}
		shouldErr   bool
	}{
		{
			"",
			"application/json; charset=utf-8",
			200,
			testBody{"success", "created", "foobar"},
			`{"status":"success","message":"created","data":"foobar"}`,
			false,
		},
		{
			"?mycallback=foobar",
			"application/json; charset=utf-8",
			200,
			testBody{"success", "created", "foobar"},
			`{"status":"success","message":"created","data":"foobar"}`,
			false,
		},
		{
			"?callback=foobar",
			"application/javascript; charset=utf-8",
			500,
			testBody{"error", "internal error", nil},
			`foobar({"status":"error","message":"internal error","data":null})`,
			false,
		},
		{
			"?callback=foobar",
			"",
			200,
			make(chan int),
			"",
			true,
		},
	}
	for _, test := range tests {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/"+test.query, nil)
		ctx := newContext(w, req)
		err := ctx.JSONP(test.code, test.data)
		if test.shouldErr {
			assert.NotNil(t, err)
			continue
		}
		assert.Equal(t, test.code, w.Code)
		assert.Equal(t, test.contentType, w.Header().Get("Content-Type"))
		assert.Equal(t, test.body, w.Body.String())
	}
}

func TestContext_JSONPBlob(t *testing.T) {
	tests := []struct {
		query       string
		contentType string
		code        int
		data        string
		body        string
	}{
		{
			"",
			"application/json; charset=utf-8",
			200,
			`{"status":"success","message":"created","data":"foobar"}`,
			`{"status":"success","message":"created","data":"foobar"}`,
		},
		{
			"?mycallback=foobar",
			"application/json; charset=utf-8",
			200,
			`{"status":"success","message":"created","data":"foobar"}`,
			`{"status":"success","message":"created","data":"foobar"}`,
		},
		{
			"?callback=foobar",
			"application/javascript; charset=utf-8",
			500,
			`{"status":"error","message":"internal error","data":null}`,
			`foobar({"status":"error","message":"internal error","data":null})`,
		},
	}
	for _, test := range tests {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/"+test.query, nil)
		ctx := newContext(w, req)
		ctx.JSONPBlob(test.code, []byte(test.data))
		assert.Equal(t, test.code, w.Code)
		assert.Equal(t, test.contentType, w.Header().Get("Content-Type"))
		assert.Equal(t, test.body, w.Body.String())
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
		assert.Equal(t, test.code, w.Code)
		assert.Equal(t, "text/plain; charset=utf-8", w.Header().Get("Content-Type"))
		assert.Equal(t, test.s, w.Body.String())
	}
}

func TestContext_Stringf(t *testing.T) {
	tests := []struct {
		code     int
		format   string
		a        []interface{}
		expected string
	}{
		{200, "hello world", nil, "hello world"},
		{500, "hello %s", []interface{}{"foobar"}, "hello foobar"},
		{500, "%d+%d=%d", []interface{}{1, 2, 1 + 2}, "1+2=3"},
	}
	for _, test := range tests {
		w := httptest.NewRecorder()
		ctx := newContext(w, nil)
		ctx.Stringf(test.code, test.format, test.a...)
		assert.Equal(t, test.code, w.Code)
		assert.Equal(t, "text/plain; charset=utf-8", w.Header().Get("Content-Type"))
		assert.Equal(t, test.expected, w.Body.String())
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
		assert.Equal(t, test.code, w.Code)
		assert.Equal(t, "application/xml; charset=utf-8", w.Header().Get("Content-Type"))
		assert.Equal(t, test.body, w.Body.String())
	}
}

func TestContext_XMLBlob(t *testing.T) {
	tests := []struct {
		code int
		data string
	}{
		{
			200,
			`<testBody><status>success</status><message>created</message><data>foobar</data></testBody>`,
		},
		{
			500,
			`<testBody><status>error</status><message>internal error</message></testBody>`,
		},
	}
	for _, test := range tests {
		w := httptest.NewRecorder()
		ctx := newContext(w, nil)
		ctx.XMLBlob(test.code, []byte(test.data))
		assert.Equal(t, test.code, w.Code)
		assert.Equal(t, "application/xml; charset=utf-8", w.Header().Get("Content-Type"))
		assert.Equal(t, test.data, w.Body.String())
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
		assert.Equal(t, test.code, w.Code)
		assert.Equal(t, "text/html; charset=utf-8", w.Header().Get("Content-Type"))
		assert.Equal(t, test.s, w.Body.String())
	}
}

func TestContext_HTMLBlob(t *testing.T) {
	tests := []struct {
		code int
		bs   []byte
	}{
		{200, []byte("<html><body>foobar</body></html>")},
		{500, []byte("<html><body>error</body></html>")},
	}
	for _, test := range tests {
		w := httptest.NewRecorder()
		ctx := newContext(w, nil)
		ctx.HTMLBlob(test.code, test.bs)
		assert.Equal(t, test.code, w.Code)
		assert.Equal(t, "text/html; charset=utf-8", w.Header().Get("Content-Type"))
		assert.Equal(t, test.bs, w.Body.Bytes())
	}
}

func TestContext_Context(t *testing.T) {
	cases := []struct {
		key   int
		value string
	}{
		{0, "0"},
		{1, "1"},
	}
	for _, test := range cases {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req = req.WithContext(context.WithValue(req.Context(), test.key, test.value))
		c := newContext(nil, req)
		ctx := c.Context()
		assert.Equal(t, req.Context(), ctx)
		assert.Equal(t, test.value, ctx.Value(test.key))
	}
}

func TestContext_FormValue(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/?foo=bar", nil)
	ctx := newContext(nil, req)
	for _, key := range []string{"foo", "fizz"} {
		assert.Equal(t, req.FormValue(key), ctx.FormValue(key))
	}
}

func TestContext_PostFormValue(t *testing.T) {
	body := bytes.NewBuffer([]byte("foo=bar"))
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	ctx := newContext(nil, req)
	for _, key := range []string{"foo", "fizz"} {
		assert.Equal(t, req.PostFormValue(key), ctx.PostFormValue(key))
	}
}

func TestContext_Host(t *testing.T) {
	cases := []struct {
		host string
	}{
		{""},
		{"example.com"},
		{"foobar.com"},
	}
	for _, test := range cases {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.Host = test.host
		ctx := newContext(nil, req)
		assert.Equal(t, req.Host, ctx.Host())
	}
}

func TestContext_QueryParams(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/?foo=bar&fizz=buzz", nil)
	ctx := newContext(nil, req)
	assert.Equal(t, ctx.QueryParams(), req.URL.Query())
	assert.Equal(t, ctx.query, req.URL.Query())
	for _, key := range []string{"foo", "fizz", "go"} {
		assert.Equal(t, req.URL.Query().Get(key), ctx.QueryParam(key))
	}
}

func TestContext_DefaultQuery(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/?foo=bar&empty=", nil)
	ctx := newContext(nil, req)
	tests := []struct {
		key          string
		defaultValue string
		value        string
	}{
		{"foo", "", "bar"},
		{"foo", "abc", "bar"},
		{"empty", "", ""},
		{"empty", "abc", ""},
		{"fizz", "", ""},
		{"fizz", "buzz", "buzz"},
	}
	for _, test := range tests {
		assert.Equal(t, test.value, ctx.DefaultQuery(test.key, test.defaultValue))
	}
}

func TestContext_QueryString(t *testing.T) {
	for _, query := range []string{"/", "/?foo=bar", "/hello?fizz=buzz"} {
		req := httptest.NewRequest(http.MethodPost, query, nil)
		ctx := newContext(nil, req)
		assert.Equal(t, req.URL.RawQuery, ctx.QueryString())
	}
}

type fakeRenderer struct {
}

func (r *fakeRenderer) Render(w io.Writer, name string, data interface{}, ctx *Context) error {
	if name == "" {
		return errors.New("empty template name")
	}
	w.Write([]byte(name))
	return nil
}

func TestContext_Render(t *testing.T) {
	w := httptest.NewRecorder()
	app := New()
	ctx := newContext(w, nil)
	ctx.app = app

	err := ctx.Render(http.StatusOK, "foo", nil)
	assert.Equal(t, ErrRendererNotRegister, err)

	app.Renderer = new(fakeRenderer)

	err = ctx.Render(http.StatusOK, "", nil)
	assert.EqualError(t, err, "empty template name")

	ctx.Render(http.StatusForbidden, "foo", nil)
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Equal(t, "text/html; charset=utf-8", w.Header().Get("Content-Type"))
	assert.Equal(t, "foo", w.Body.String())
}

func TestContext_RouteURL(t *testing.T) {
	app := New()
	app.Get("/", echoHandler("foo"), RouteName("foo"))
	ctx := newContext(nil, nil)
	ctx.app = app

	actual, _ := ctx.RouteURL("foo")
	expected, _ := app.RouteURL("foo")
	assert.Equal(t, expected, actual)

	_, actualErr := ctx.RouteURL("bar")
	_, expectedErr := app.RouteURL("bar")
	assert.Equal(t, expectedErr, actualErr)
}

func TestContext_ServeFile(t *testing.T) {
	w1 := httptest.NewRecorder()
	w2 := httptest.NewRecorder()
	ctx := newContext(w2, httptest.NewRequest(http.MethodGet, "/", nil))
	assert.Nil(t, ctx.ServeFile("foo"))
	http.ServeFile(w1, httptest.NewRequest(http.MethodGet, "/", nil), "foo")
	assert.Equal(t, w1, w2)
}

func TestContext_ServeContent(t *testing.T) {
	w1 := httptest.NewRecorder()
	w2 := httptest.NewRecorder()
	ctx := newContext(w2, httptest.NewRequest(http.MethodGet, "/", nil))
	now := time.Now()
	buf := bytes.NewReader([]byte("bar"))
	assert.Nil(t, ctx.ServeContent("foo", now, buf))
	http.ServeContent(w1, httptest.NewRequest(http.MethodGet, "/", nil), "foo", now, buf)
	assert.Equal(t, w1, w2)
}

func TestContext_BasicAuth(t *testing.T) {
	requests := []*http.Request{
		httptest.NewRequest(http.MethodGet, "/", nil),
	}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.SetBasicAuth("foo", "bar")
	requests = append(requests, req)
	for _, req := range requests {
		ctx := newContext(nil, req)
		user1, pass1, ok1 := req.BasicAuth()
		user2, pass2, ok2 := ctx.BasicAuth()
		assert.Equal(t, user1, user2)
		assert.Equal(t, pass1, pass2)
		assert.Equal(t, ok1, ok2)
	}
}

func TestContext_SendFile(t *testing.T) {
	w := httptest.NewRecorder()
	ctx := newContext(w, nil)
	buf := bytes.NewReader([]byte("bar"))
	assert.Nil(t, ctx.SendFile("foo.txt", buf))
	assert.Equal(t, "bar", w.Body.String())
	assert.Equal(t, w.Header().Get("Content-Disposition"), `attachment; filename="foo.txt"`)
}

type fakeDecoder struct {
	err error
}

func (d *fakeDecoder) Decode(req *http.Request, v interface{}) error {
	return d.err
}

type fakeForm struct {
	Name string `json:"name"`
}

func TestContext_Decode(t *testing.T) {
	ctx := newContext(nil, httptest.NewRequest(http.MethodPost, "/", nil))
	ctx.app = New()
	v := new(fakeForm)
	assert.Equal(t, ErrDecoderNotRegister, ctx.Decode(v))

	decodeErr := errors.New("decoder error")
	ctx.app.Decoder = &fakeDecoder{err: decodeErr}
	assert.Equal(t, decodeErr, ctx.Decode(v))

	ctx.app.Decoder = &fakeDecoder{}
	assert.Nil(t, ctx.Decode(v))
}

func TestContextSetHeader(t *testing.T) {
	cases := map[string]string{
		"X-Foo":  "Bar",
		"X-Fizz": "Buzz",
	}
	for k, v := range cases {
		w := httptest.NewRecorder()
		ctx := newContext(w, nil)
		ctx.SetHeader(k, v)
		assert.Equal(t, v, w.Header().Get(k))
	}
}
