// Copyright 2020 CleverGo. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package clevergo

import (
	"net/http/httptest"
	"testing"
)

func TestSetContentType(t *testing.T) {
	tests := []string{
		"text/html",
		"text/html; charset=utf-8",
		"text/plain",
		"text/plain; charset=utf-8",
		"application/json",
		"application/xml",
	}

	for _, test := range tests {
		w := httptest.NewRecorder()
		SetContentType(w, test)
		if w.Header().Get("Content-Type") != test {
			t.Errorf("expected content type %q, got %q", test, w.Header().Get("Content-Type"))
		}
	}
}

func TestSetContentTypeHTML(t *testing.T) {
	w := httptest.NewRecorder()
	SetContentTypeHTML(w)
	if w.Header().Get("Content-Type") != "text/html; charset=utf-8" {
		t.Errorf("expected content type %q, got %q", "text/html; charset=utf-8", w.Header().Get("Content-Type"))
	}
}
func TestSetContentTypeText(t *testing.T) {
	w := httptest.NewRecorder()
	SetContentTypeText(w)
	if w.Header().Get("Content-Type") != "text/plain; charset=utf-8" {
		t.Errorf("expected content type %q, got %q", "text/plain; charset=utf-8", w.Header().Get("Content-Type"))
	}
}
func TestSetContentTypeJSON(t *testing.T) {
	w := httptest.NewRecorder()
	SetContentTypeJSON(w)
	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected content type %q, got %q", "application/json", w.Header().Get("Content-Type"))
	}
}
func TestSetContentTypeXML(t *testing.T) {
	w := httptest.NewRecorder()
	SetContentTypeXML(w)
	if w.Header().Get("Content-Type") != "application/xml" {
		t.Errorf("expected content type %q, got %q", "application/xml", w.Header().Get("Content-Type"))
	}
}
