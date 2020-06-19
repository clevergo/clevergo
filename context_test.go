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
		c := newContext(httptest.NewRecorder(), nil)
		c.SetContentType(test)
		assert.Equal(t, test, c.Response.Header().Get("Content-Type"))
	}
}

func TestContext_SetContentTypeHTML(t *testing.T) {
	c := newContext(httptest.NewRecorder(), nil)
	c.SetContentTypeHTML()
	assert.Equal(t, "text/html; charset=utf-8", c.Response.Header().Get("Content-Type"))
}
func TestContext_SetContentTypeText(t *testing.T) {
	c := newContext(httptest.NewRecorder(), nil)
	c.SetContentTypeText()
	assert.Equal(t, "text/plain; charset=utf-8", c.Response.Header().Get("Content-Type"))
}
func TestContext_SetContentTypeJSON(t *testing.T) {
	c := newContext(httptest.NewRecorder(), nil)
	c.SetContentTypeJSON()
	assert.Equal(t, "application/json; charset=utf-8", c.Response.Header().Get("Content-Type"))
}
func TestContext_SetContentTypeXML(t *testing.T) {
	c := newContext(httptest.NewRecorder(), nil)
	c.SetContentTypeXML()
	assert.Equal(t, "application/xml; charset=utf-8", c.Response.Header().Get("Content-Type"))
}

func TestContext_Write(t *testing.T) {
	tests := [][]byte{
		[]byte("foo"),
		[]byte("bar"),
	}

	for _, test := range tests {
		w := httptest.NewRecorder()
		c := newContext(w, nil)
		c.Write(test)
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
		c := newContext(w, nil)
		c.WriteString(test)
		assert.Equal(t, test, w.Body.String())
	}
}

func TestContext_NotFound(t *testing.T) {
	w := httptest.NewRecorder()
	c := newContext(w, nil)
	assert.Nil(t, c.NotFound())
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestContext_Redirect(t *testing.T) {
	w := httptest.NewRecorder()
	c := newContext(w, httptest.NewRequest(http.MethodGet, "/", nil))
	assert.Nil(t, c.Redirect(http.StatusPermanentRedirect, "/redirect"))
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
		c := newContext(w, nil)
		assert.Nil(t, c.Error(test.code, test.msg))
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

	c := newContext(nil, httptest.NewRequest(http.MethodGet, "/", nil))
	for key, val := range values {
		c.WithValue(key, val)
	}

	for key, val := range values {
		assert.Equal(t, val, c.Value(key))
	}
}

func TestIsMethod(t *testing.T) {
	tests := []struct {
		method string
		f      func(c *Context, method string) bool
	}{
		{http.MethodGet, func(c *Context, method string) bool {
			return c.IsGet()
		}},
		{http.MethodDelete, func(c *Context, method string) bool {
			return c.IsDelete()
		}},
		{http.MethodPatch, func(c *Context, method string) bool {
			return c.IsPatch()
		}},
		{http.MethodPost, func(c *Context, method string) bool {
			return c.IsPost()
		}},
		{http.MethodPut, func(c *Context, method string) bool {
			return c.IsPut()
		}},
		{http.MethodOptions, func(c *Context, method string) bool {
			return c.IsOptions()
		}},
		{http.MethodHead, func(c *Context, method string) bool {
			return c.IsMethod(method)
		}},
	}
	for _, test := range tests {
		c := newContext(nil, httptest.NewRequest(test.method, "/", nil))
		assert.True(t, test.f(c, test.method), "failed to determine request method")
	}
}

func TestContext_Cookie(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "foo", Value: "bar"})
	c := newContext(nil, req)
	actual, _ := c.Cookie("foo")
	expected, _ := req.Cookie("foo")
	assert.Equal(t, expected, actual)
}

func TestContext_Cookies(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "foo", Value: "bar"})
	c := newContext(nil, req)
	assert.Equal(t, req.Cookies(), c.Cookies())
}

func TestContext_SetCookie(t *testing.T) {
	w := httptest.NewRecorder()
	c := &Context{Response: w}
	cookie := &http.Cookie{Name: "foo", Value: "bar"}
	c.SetCookie(cookie)
	actual := w.Result().Cookies()[0]
	assert.Equal(t, cookie.Name, actual.Name)
	assert.Equal(t, cookie.Value, actual.Value)
}

func TestContext_WriteHeader(t *testing.T) {
	codes := []int{http.StatusOK, http.StatusForbidden, http.StatusInternalServerError, http.StatusUnauthorized}
	for _, code := range codes {
		w := httptest.NewRecorder()
		c := newContext(w, nil)
		c.WriteHeader(code)
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
		c := newContext(nil, req)
		assert.Equal(t, test.expected, c.IsAJAX())
	}
}

func TestContext_GetHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("foo", "bar")
	c := newContext(nil, req)
	for _, name := range []string{"foo", "fizz"} {
		assert.Equal(t, req.Header.Get(name), c.GetHeader(name))
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
		c := newContext(w, nil)
		err := c.JSON(test.code, test.data)
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
		c := newContext(w, nil)
		c.JSONBlob(test.code, []byte(test.data))
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
		c := newContext(w, req)
		err := c.JSONP(test.code, test.data)
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
		c := newContext(w, req)
		c.JSONPBlob(test.code, []byte(test.data))
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
		c := newContext(w, nil)
		c.String(test.code, test.s)
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
		c := newContext(w, nil)
		c.Stringf(test.code, test.format, test.a...)
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
		c := newContext(w, nil)
		err := c.XML(test.code, test.data)
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
		c := newContext(w, nil)
		c.XMLBlob(test.code, []byte(test.data))
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
		c := newContext(w, nil)
		c.HTML(test.code, test.s)
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
		c := newContext(w, nil)
		c.HTMLBlob(test.code, test.bs)
		assert.Equal(t, test.code, w.Code)
		assert.Equal(t, "text/html; charset=utf-8", w.Header().Get("Content-Type"))
		assert.Equal(t, test.bs, w.Body.Bytes())
	}
}

func TestContext_Context(t *testing.T) {
	cases := []struct {
		key   interface{}
		value string
	}{
		{0, "0"},
		{1.0, "1"},
	}
	for _, test := range cases {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req = req.WithContext(context.WithValue(req.Context(), test.key, test.value))
		c := newContext(nil, req)
		ctx := c.Context()
		assert.Equal(t, req.Context(), ctx)
		assert.Equal(t, test.value, c.Value(test.key))
	}
}

func TestContext_FormValue(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/?foo=bar", nil)
	c := newContext(nil, req)
	for _, key := range []string{"foo", "fizz"} {
		assert.Equal(t, req.FormValue(key), c.FormValue(key))
	}
}

func TestContext_PostFormValue(t *testing.T) {
	body := bytes.NewBuffer([]byte("foo=bar"))
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	c := newContext(nil, req)
	for _, key := range []string{"foo", "fizz"} {
		assert.Equal(t, req.PostFormValue(key), c.PostFormValue(key))
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
		c := newContext(nil, req)
		assert.Equal(t, req.Host, c.Host())
	}
}

func TestContext_QueryParams(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/?foo=bar&fizz=buzz", nil)
	c := newContext(nil, req)
	assert.Equal(t, c.QueryParams(), req.URL.Query())
	assert.Equal(t, c.query, req.URL.Query())
	for _, key := range []string{"foo", "fizz", "go"} {
		assert.Equal(t, req.URL.Query().Get(key), c.QueryParam(key))
	}
}

func TestContext_DefaultQuery(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/?foo=bar&empty=", nil)
	c := newContext(nil, req)
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
		assert.Equal(t, test.value, c.DefaultQuery(test.key, test.defaultValue))
	}
}

func TestContext_QueryString(t *testing.T) {
	for _, query := range []string{"/", "/?foo=bar", "/hello?fizz=buzz"} {
		req := httptest.NewRequest(http.MethodPost, query, nil)
		c := newContext(nil, req)
		assert.Equal(t, req.URL.RawQuery, c.QueryString())
	}
}

type fakeRenderer struct {
}

func (r *fakeRenderer) Render(w io.Writer, name string, data interface{}, c *Context) error {
	if name == "" {
		return errors.New("empty template name")
	}
	w.Write([]byte(name))
	return nil
}

func TestContext_Render(t *testing.T) {
	w := httptest.NewRecorder()
	app := New()
	c := newContext(w, nil)
	c.app = app

	err := c.Render(http.StatusOK, "foo", nil)
	assert.Equal(t, ErrRendererNotRegister, err)

	app.Renderer = new(fakeRenderer)

	err = c.Render(http.StatusOK, "", nil)
	assert.EqualError(t, err, "empty template name")

	c.Render(http.StatusForbidden, "foo", nil)
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Equal(t, "text/html; charset=utf-8", w.Header().Get("Content-Type"))
	assert.Equal(t, "foo", w.Body.String())
}

func TestContext_RouteURL(t *testing.T) {
	app := New()
	app.Get("/", echoHandler("foo"), RouteName("foo"))
	c := newContext(nil, nil)
	c.app = app

	actual, _ := c.RouteURL("foo")
	expected, _ := app.RouteURL("foo")
	assert.Equal(t, expected, actual)

	_, actualErr := c.RouteURL("bar")
	_, expectedErr := app.RouteURL("bar")
	assert.Equal(t, expectedErr, actualErr)
}

func TestContext_ServeFile(t *testing.T) {
	w1 := httptest.NewRecorder()
	w2 := httptest.NewRecorder()
	c := newContext(w2, httptest.NewRequest(http.MethodGet, "/", nil))
	assert.Nil(t, c.ServeFile("foo"))
	http.ServeFile(w1, httptest.NewRequest(http.MethodGet, "/", nil), "foo")
	assert.Equal(t, w1, w2)
}

func TestContext_ServeContent(t *testing.T) {
	w1 := httptest.NewRecorder()
	w2 := httptest.NewRecorder()
	c := newContext(w2, httptest.NewRequest(http.MethodGet, "/", nil))
	now := time.Now()
	buf := bytes.NewReader([]byte("bar"))
	assert.Nil(t, c.ServeContent("foo", now, buf))
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
		c := newContext(nil, req)
		user1, pass1, ok1 := req.BasicAuth()
		user2, pass2, ok2 := c.BasicAuth()
		assert.Equal(t, user1, user2)
		assert.Equal(t, pass1, pass2)
		assert.Equal(t, ok1, ok2)
	}
}

func TestContext_SendFile(t *testing.T) {
	w := httptest.NewRecorder()
	c := newContext(w, nil)
	buf := bytes.NewReader([]byte("bar"))
	assert.Nil(t, c.SendFile("foo.txt", buf))
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
	c := newContext(nil, httptest.NewRequest(http.MethodPost, "/", nil))
	c.app = New()
	v := new(fakeForm)
	assert.Equal(t, ErrDecoderNotRegister, c.Decode(v))

	decodeErr := errors.New("decoder error")
	c.app.Decoder = &fakeDecoder{err: decodeErr}
	assert.Equal(t, decodeErr, c.Decode(v))

	c.app.Decoder = &fakeDecoder{}
	assert.Nil(t, c.Decode(v))
}

func TestContextSetHeader(t *testing.T) {
	cases := map[string]string{
		"X-Foo":  "Bar",
		"X-Fizz": "Buzz",
	}
	for k, v := range cases {
		w := httptest.NewRecorder()
		c := newContext(w, nil)
		c.SetHeader(k, v)
		assert.Equal(t, v, w.Header().Get(k))
	}
}

func TestContextLogger(t *testing.T) {
	app := New()
	ctx := &Context{
		app: app,
	}
	assert.Equal(t, app.Logger, ctx.Logger())
}
