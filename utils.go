// Copyright 2020 CleverGo. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package clevergo

import "net/http"

// SetContentType sets the content type header.
func SetContentType(w http.ResponseWriter, v string) {
	w.Header().Set("Content-Type", v)
}

// SetContentTypeHTML sets the content type as HTML.
func SetContentTypeHTML(w http.ResponseWriter) {
	SetContentType(w, "text/html; charset=utf-8")
}

// SetContentTypeText sets the content type as text.
func SetContentTypeText(w http.ResponseWriter) {
	SetContentType(w, "text/plain; charset=utf-8")
}

// SetContentTypeJSON sets the content type as JSON.
func SetContentTypeJSON(w http.ResponseWriter) {
	SetContentType(w, "application/json")
}

// SetContentTypeXML sets the content type as XML.
func SetContentTypeXML(w http.ResponseWriter) {
	SetContentType(w, "application/xml")
}
