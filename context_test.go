// Copyright 2016 The Gem Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license
// that can be found in the LICENSE file.

package gem

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"net/http"
	"reflect"
	"testing"
)

func TestContext_UserValue(t *testing.T) {
	ctx := &Context{}

	if ctx.UserValue("empty") != nil {
		t.Errorf("expected %v, got %v", nil, ctx.UserValue("empty"))
	}

	ctx.userValue = new(userValue)

	if ctx.UserValue("empty") != nil {
		t.Errorf("expected %v, got %v", nil, ctx.UserValue("empty"))
	}

	str := "foo"
	ctx.SetUserValue("string", str)

	num := 2016
	ctx.SetUserValue("integer", num)

	if !reflect.DeepEqual(ctx.UserValue("integer"), num) {
		t.Errorf("expect %d, got %d", num, ctx.UserValue("integer"))
	}
	if !reflect.DeepEqual(ctx.UserValue("string"), str) {
		t.Errorf("expect %q, got %q", str, ctx.UserValue("string"))
	}
}

func TestContext_IsDelete(t *testing.T) {
	req, _ := http.NewRequest(MethodDelete, "", nil)

	ctx := &Context{Request: req}
	if !ctx.IsDelete() {
		t.Errorf("expected ctx.IsDelete(): %t, got %t", true, ctx.IsDelete())
	}

	req.Method = MethodPost
	if ctx.IsDelete() {
		t.Errorf("expected ctx.IsDelete() = %t, got %t", false, ctx.IsDelete())
	}
}

func TestContext_IsGet(t *testing.T) {
	req, _ := http.NewRequest(MethodGet, "", nil)

	ctx := &Context{Request: req}
	if !ctx.IsGet() {
		t.Errorf("expected ctx.IsGet(): %t, got %t", true, ctx.IsDelete())
	}

	req.Method = MethodPost
	if ctx.IsGet() {
		t.Errorf("expected ctx.IsGet() = %t, got %t", false, ctx.IsDelete())
	}
}

func TestContext_IsPost(t *testing.T) {
	req, _ := http.NewRequest(MethodPost, "", nil)

	ctx := &Context{Request: req}
	if !ctx.IsPost() {
		t.Errorf("expected ctx.IsPost(): %t, got %t", true, ctx.IsDelete())
	}

	req.Method = MethodGet
	if ctx.IsPost() {
		t.Errorf("expected ctx.IsPost() = %t, got %t", false, ctx.IsDelete())
	}
}

func TestContext_IsPut(t *testing.T) {
	req, _ := http.NewRequest(MethodPut, "", nil)

	ctx := &Context{Request: req}
	if !ctx.IsPut() {
		t.Errorf("expected ctx.IsPut(): %t, got %t", true, ctx.IsDelete())
	}

	req.Method = MethodPost
	if ctx.IsPut() {
		t.Errorf("expected ctx.IsPut() = %t, got %t", false, ctx.IsDelete())
	}
}

func TestContext_IsHead(t *testing.T) {
	req, _ := http.NewRequest(MethodHead, "", nil)

	ctx := &Context{Request: req}
	if !ctx.IsHead() {
		t.Errorf("expected ctx.IsHead(): %t, got %t", true, ctx.IsDelete())
	}

	req.Method = MethodPost
	if ctx.IsHead() {
		t.Errorf("expected ctx.IsHead() = %t, got %t", false, ctx.IsDelete())
	}
}

func TestContext_IsAjax(t *testing.T) {
	req, _ := http.NewRequest(MethodGet, "", nil)

	ctx := &Context{Request: req}
	if ctx.IsAjax() {
		t.Errorf("expected ctx.IsAjax(): %t, got %t", false, ctx.IsAjax())
	}

	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	if !ctx.IsAjax() {
		t.Errorf("expected ctx.IsAjax() = %t, got %t", true, ctx.IsAjax())
	}
}

func TestContext_FormFile(t *testing.T) {
	req, _ := http.NewRequest(MethodGet, "", nil)

	ctx := &Context{Request: req}

	key := "key"
	f1, fh1, err1 := ctx.FormFile(key)
	f2, fh2, err2 := req.FormFile(key)

	if f1 != f2 || fh1 != fh2 || err1 != err2 {
		t.Error("failed to get form file")
	}
}

func TestContext_FormValue(t *testing.T) {
	req, _ := http.NewRequest(MethodGet, "", nil)

	ctx := &Context{Request: req}
	key := "key"
	if ctx.FormValue(key) != req.FormValue(key) {
		t.Error("failed to get form value")
	}
}

func TestContext_URL(t *testing.T) {
	req, _ := http.NewRequest(MethodGet, "", nil)

	ctx := &Context{Request: req}
	if ctx.URL() != req.URL {
		t.Error("failed to get request url")
	}
}

type testUser struct {
	Name string `json:"name" xml:"name"`
}

var user = testUser{Name: "foo"}

func TestContext_HTML(t *testing.T) {
	resp := &mockResponseWriter{}
	ctx := &Context{Response: resp}
	body := "foo"
	ctx.HTML(http.StatusOK, body)
	if resp.statusCode != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, resp.statusCode)
	}
	if !bytes.Equal(resp.body, []byte(body)) {
		t.Errorf("expected response body %q, got %q", body, resp.body)
	}
}

func TestContext_JSON(t *testing.T) {
	respJson, err := json.Marshal(user)
	if err != nil {
		t.Fatal(err)
	}

	resp := &mockResponseWriter{}
	ctx := &Context{Response: resp, server: New("")}
	ctx.JSON(http.StatusOK, user)
	if resp.statusCode != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, resp.statusCode)
	}
	if !bytes.Equal(resp.body, respJson) {
		t.Errorf("expected response body %q, got %q", respJson, resp.body)
	}

	ctx.JSON(http.StatusInternalServerError, make(chan struct{}))
	if resp.statusCode != http.StatusInternalServerError {
		t.Errorf("expected status code %d, got %d", http.StatusInternalServerError, resp.statusCode)
	}
}

func TestContext_XML(t *testing.T) {
	resp := &mockResponseWriter{}
	ctx := &Context{Response: resp, server: New("")}
	ctx.XML(http.StatusOK, user)
	if resp.statusCode != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, resp.statusCode)
	}

	userObj := testUser{}
	if err := xml.Unmarshal(resp.body, &userObj); err != nil {
		t.Errorf("fialed to unmarshal xml data %q", err)
	} else {
		if userObj.Name != user.Name {
			t.Errorf("expected user name %q, got %q", user.Name, userObj.Name)
		}
	}

	ctx.XML(http.StatusInternalServerError, make(chan struct{}))
	if resp.statusCode != http.StatusInternalServerError {
		t.Errorf("expected status code %d, got %d", http.StatusInternalServerError, resp.statusCode)
	}

	header := `<?xml version="2.0" encoding="UTF-8"?>` + "\n"
	ctx.XML(http.StatusOK, user, header)
	if len(resp.body) < len(header) {
		t.Error("incorrect xml header")
	} else if !bytes.Equal(resp.body[:len(header)], []byte(header)) {
		t.Errorf("expected xml header %q, got %q", resp.body[:len(header)], header)
	}
}

func TestContext_Logger(t *testing.T) {
	srv := New("")
	ctx := &Context{server: srv}
	if ctx.Logger() != srv.logger {
		t.Error("failed to get logger")
	}
}

func TestContext_SetContentType(t *testing.T) {
	contentType := "text/html"

	resp := &mockResponseWriter{}
	ctx := &Context{Response: resp}
	ctx.SetContentType(contentType)

	if resp.Header().Get("Content-Type") != contentType {
		t.Error("failed to set content type")
	}
}

func TestContext_Write(t *testing.T) {
	resp := &mockResponseWriter{}
	ctx := &Context{Response: resp}

	msg := []byte("Hello world.")

	n1, err1 := ctx.Write(msg)
	n2, err2 := resp.Write(msg)

	if n1 != n2 || err1 != err2 {
		t.Error("failed to write response")
	}
}

func TestContext_Push(t *testing.T) {
	resp := &mockResponseWriter{}
	ctx := &Context{Response: resp}

	err := ctx.Push("/", nil)
	if err != errNotSupportHTTP2ServerPush {
		t.Errorf("expected push error: %q, got %q", errNotSupportHTTP2ServerPush, err)
	}

	resp2 := &mockResponseWriter2{resp}
	ctx.Response = resp2
	err = ctx.Push("/", nil)
	err2 := resp2.Push("/", nil)
	if err != err2 {
		t.Errorf("expected push error: %q, got %q", err2, err)
	}
}
