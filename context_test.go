// Copyright 2020 CleverGo. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package clevergo

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
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
		ctx := NewContext(httptest.NewRecorder(), nil)
		ctx.SetContentType(test)
		if ctx.Response.Header().Get("Content-Type") != test {
			t.Errorf("expected content type %q, got %q", test, ctx.Response.Header().Get("Content-Type"))
		}
	}
}

func TestContext_SetContentTypeHTML(t *testing.T) {
	ctx := NewContext(httptest.NewRecorder(), nil)
	ctx.SetContentTypeHTML()
	if ctx.Response.Header().Get("Content-Type") != "text/html; charset=utf-8" {
		t.Errorf("expected content type %q, got %q", "text/html; charset=utf-8", ctx.Response.Header().Get("Content-Type"))
	}
}
func TestContext_SetContentTypeText(t *testing.T) {
	ctx := NewContext(httptest.NewRecorder(), nil)
	ctx.SetContentTypeText()
	if ctx.Response.Header().Get("Content-Type") != "text/plain; charset=utf-8" {
		t.Errorf("expected content type %q, got %q", "text/plain; charset=utf-8", ctx.Response.Header().Get("Content-Type"))
	}
}
func TestContext_SetContentTypeJSON(t *testing.T) {
	ctx := NewContext(httptest.NewRecorder(), nil)
	ctx.SetContentTypeJSON()
	if ctx.Response.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected content type %q, got %q", "application/json", ctx.Response.Header().Get("Content-Type"))
	}
}
func TestContext_SetContentTypeXML(t *testing.T) {
	ctx := NewContext(httptest.NewRecorder(), nil)
	ctx.SetContentTypeXML()
	if ctx.Response.Header().Get("Content-Type") != "application/xml" {
		t.Errorf("expected content type %q, got %q", "application/xml", ctx.Response.Header().Get("Content-Type"))
	}
}

func TestContext_Write(t *testing.T) {
	tests := [][]byte{
		[]byte("foo"),
		[]byte("bar"),
	}

	for _, test := range tests {
		w := httptest.NewRecorder()
		ctx := NewContext(w, nil)
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
		ctx := NewContext(w, nil)
		ctx.WriteString(test)
		if w.Body.String() != test {
			t.Errorf("expected body %q, got %q", test, w.Body.String())
		}
	}
}

func TestContext_NotFound(t *testing.T) {
	w := httptest.NewRecorder()
	ctx := NewContext(w, nil)
	ctx.NotFound()
	if w.Code != http.StatusNotFound {
		t.Errorf("expected status code %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestContext_Redirect(t *testing.T) {
	w := httptest.NewRecorder()
	ctx := NewContext(w, httptest.NewRequest(http.MethodGet, "/", nil))
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
		ctx := NewContext(w, nil)
		ctx.Error(test.msg, test.code)
		if w.Body.String() != fmt.Sprintln(test.msg) {
			t.Errorf("expected body %q, got %q", fmt.Sprintln(test.msg), w.Body.String())
		}
		if w.Code != test.code {
			t.Errorf("expected status code %d, got %d", test.code, w.Code)
		}
	}
}
