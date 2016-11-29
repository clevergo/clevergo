// Copyright 2016 The Gem Authors. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package gem

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"testing"

	"github.com/go-gem/tests"
	"github.com/valyala/fasthttp"
)

type project struct {
	Name string `json:"name" xml:"name"`
}

var (
	contentType = []byte("Content-Type")
)

func TestContext(t *testing.T) {
	router := NewRouter()
	respHtml := "OK"
	router.GET("/html", func(ctx *Context) {
		ctx.HTML(fasthttp.StatusOK, respHtml)
	})

	invalidValue := make(chan bool)
	p := project{Name: "GEM"}
	respJson, err := json.Marshal(&p)
	if err != nil {
		t.Fatalf("json.Marshal error %s", err)
	}

	router.GET("/json", func(ctx *Context) {
		ctx.JSON(fasthttp.StatusOK, p)
	})
	router.GET("/json2", func(ctx *Context) {
		ctx.JSON(fasthttp.StatusOK, invalidValue)
	})

	jsonpCallback := []byte("callback")
	var respJsonp []byte
	respJsonp = append(respJsonp, jsonpCallback...)
	respJsonp = append(respJsonp, "("...)
	respJsonp = append(respJsonp, respJson...)
	respJsonp = append(respJsonp, ")"...)

	router.GET("/jsonp", func(ctx *Context) {
		ctx.JSONP(fasthttp.StatusOK, p, jsonpCallback)
	})

	router.GET("/jsonp2", func(ctx *Context) {
		ctx.JSONP(fasthttp.StatusOK, invalidValue, jsonpCallback)
	})

	router.GET("/xml", func(ctx *Context) {
		ctx.XML(fasthttp.StatusOK, p, xml.Header)
	})
	router.GET("/xml2", func(ctx *Context) {
		ctx.XML(fasthttp.StatusOK, invalidValue, xml.Header)
	})

	srv := New("", router.Handler)

	// HTML
	test1 := tests.New(srv)
	test1.Url = "/html"
	test1.Expect().Status(fasthttp.StatusOK).Custom(func(resp fasthttp.Response) error {
		if !bytes.Equal(resp.Header.PeekBytes(contentType), HeaderContentTypeHTMLBytes) {
			return fmt.Errorf("unexpected Content-Type got %q want %q", resp.Header.PeekBytes(contentType), HeaderContentTypeHTMLBytes)
		}
		if !bytes.Equal(resp.Body(), []byte(respHtml)) {
			return fmt.Errorf("unexpected response got %q want %q", string(resp.Body()), respHtml)
		}
		return nil
	})
	if err = test1.Run(); err != nil {
		t.Error(err)
	}

	// JSON
	test2 := tests.New(srv)
	test2.Url = "/json"
	test2.Expect().Status(fasthttp.StatusOK).Custom(func(resp fasthttp.Response) error {
		if !bytes.Equal(resp.Header.PeekBytes(contentType), HeaderContentTypeJSONBytes) {
			return fmt.Errorf("unexpected Content-Type got %q want %q", resp.Header.PeekBytes(contentType), HeaderContentTypeJSONBytes)
		}
		if !bytes.Equal(resp.Body(), []byte(respJson)) {
			return fmt.Errorf("unexpected response got %q want %q", string(resp.Body()), respJson)
		}
		return nil
	})
	if err = test2.Run(); err != nil {
		t.Error(err)
	}

	test3 := tests.New(srv)
	test3.Url = "/json2"
	test3.Expect().Status(fasthttp.StatusInternalServerError).Custom(func(resp fasthttp.Response) error {
		return nil
	})
	if err = test3.Run(); err != nil {
		t.Error(err)
	}

	// JSONP
	test4 := tests.New(srv)
	test4.Url = "/jsonp"
	test4.Expect().Status(fasthttp.StatusOK).Custom(func(resp fasthttp.Response) error {
		if !bytes.Equal(resp.Header.PeekBytes(contentType), HeaderContentTypeJSONPBytes) {
			return fmt.Errorf("unexpected Content-Type got %q want %q", resp.Header.PeekBytes(contentType), HeaderContentTypeJSONPBytes)
		}
		if !bytes.Equal(resp.Body(), []byte(respJsonp)) {
			return fmt.Errorf("unexpected response got %q want %q", string(resp.Body()), respJsonp)
		}
		return nil
	})
	if err = test4.Run(); err != nil {
		t.Error(err)
	}

	test5 := tests.New(srv)
	test5.Url = "/jsonp2"
	test5.Expect().Status(fasthttp.StatusInternalServerError).Custom(func(resp fasthttp.Response) error {
		return nil
	})
	if err = test5.Run(); err != nil {
		t.Error(err)
	}

	// XML
	test6 := tests.New(srv)
	test6.Url = "/xml"
	test6.Expect().Status(fasthttp.StatusOK).Custom(func(resp fasthttp.Response) error {
		if !bytes.Equal(resp.Header.PeekBytes(contentType), HeaderContentTypeXMLBytes) {
			return fmt.Errorf("unexpected Content-Type got %q want %q", resp.Header.PeekBytes(contentType), HeaderContentTypeXMLBytes)
		}
		p2 := project{}
		if err := xml.Unmarshal(resp.Body(), &p2); err != nil {
			return fmt.Errorf("xml.Unmarshal error %s", err)
		}
		if p2.Name != p.Name {
			return fmt.Errorf("unexpected project's name got %q want %q", p2.Name, p.Name)
		}
		return nil
	})
	if err = test6.Run(); err != nil {
		t.Error(err)
	}

	test7 := tests.New(srv)
	test7.Url = "/xml2"
	test7.Expect().Status(fasthttp.StatusInternalServerError).Custom(func(resp fasthttp.Response) error {
		return nil
	})
	if err = test7.Run(); err != nil {
		t.Error(err)
	}

	var handleErr error
	router.GET("/", func(ctx *Context) {
		if !ctx.IsAjax() {
			handleErr = fmt.Errorf("expected c.IsAjax() = %t, got %t", true, ctx.IsAjax())
			return
		}

		if srv.logger != ctx.Logger() {
			handleErr = fmt.Errorf("unexpected logger")
			return
		}

		if srv.sessionsStore != ctx.SessionsStore() {
			handleErr = fmt.Errorf("unexpected sessions store")
			return
		}
	})

	test8 := tests.New(srv)
	test8.Headers[HeaderXRequestedWith] = HeaderXMLHttpRequest
	test8.Expect().Status(fasthttp.StatusOK)
	if err = test8.Run(); err != nil {
		t.Error(err)
	}
	if handleErr != nil {
		t.Error(handleErr)
	}
}

func TestContext_Param(t *testing.T) {
	router := NewRouter()
	srv := New("", router.Handler)

	router.GET("/user/:name", func(ctx *Context) {
		ctx.HTML(fasthttp.StatusOK, ctx.Param("name"))
	})

	router.POST("/user/:name", func(ctx *Context) {
		ctx.SetUserValue("name", 1)
		ctx.HTML(fasthttp.StatusOK, ctx.Param("name"))
	})

	var err error

	test1 := tests.New(srv)
	test1.Url = "/user/foo"
	test1.Expect().Status(fasthttp.StatusOK).Custom(func(resp fasthttp.Response) error {
		body := string(resp.Body())
		if body != "foo" {
			return fmt.Errorf("expected body %q, got %q", "foo", body)
		}

		return nil
	})
	if err = test1.Run(); err != nil {
		t.Error(err)
	}

	test2 := tests.New(srv)
	test2.Url = "/user/foo"
	test2.Method = MethodPost
	test2.Expect().Status(fasthttp.StatusOK).Custom(func(resp fasthttp.Response) error {
		body := string(resp.Body())
		if body != "" {
			return fmt.Errorf("expected empty body, got %q", body)
		}

		return nil
	})
	if err = test2.Run(); err != nil {
		t.Error(err)
	}
}

func TestContext_ParamInt(t *testing.T) {
	router := NewRouter()
	srv := New("", router.Handler)

	var page int

	router.GET("/list/:page", func(ctx *Context) {
		page = ctx.ParamInt("page")
	})

	var err error

	test1 := tests.New(srv)
	test1.Url = "/list/2"
	test1.Expect().Status(fasthttp.StatusOK)
	if err = test1.Run(); err != nil {
		t.Error(err)
	}
	if page != 2 {
		t.Errorf("expected page: %d, got %d", 2, page)
	}

	// empty page
	test2 := tests.New(srv)
	test2.Url = "/list/invalid_number"
	test2.Expect().Status(fasthttp.StatusOK)
	if err = test2.Run(); err != nil {
		t.Error(err)
	}
	if page != 0 {
		t.Errorf("expected page: %d, got %d", 0, page)
	}
}
