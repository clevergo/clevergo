// Copyright 2020 CleverGo. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package clevergo

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

var (
	tmpDir      string
	keyFile     string
	certFile    string
	keyFileData = []byte(`
-----BEGIN PRIVATE KEY-----
MIICdgIBADANBgkqhkiG9w0BAQEFAASCAmAwggJcAgEAAoGBALjTr4iR1sELNuL9
1y//5ZQKzfzKDqxkuXJb28nHzYlrJja0hss7IujSIEMiBUOsRu14TPKmCuVs32Qp
SgcnbdEjcfwsQ8LlN2PJecy+BwNwv//p5CieklUY1hSBAran9Y75ma51KAMvV6Ft
cmjS8/iRhzkbopo5QpkTwYsJYFhhAgMBAAECgYEAp6QBm7rD8fatAvhAnR3a6wtd
yMKwynbVqb9dvEiIyfKxB394n490233zm1CZO8df0faCvLgUPAIjISM+LPz7YeLO
vSqMiqiGRfOEJTEcZV3WEVagbA50RfJgglqEUvwHY0uAgjx6lwNWQ4IolbX3DPDK
PaONRq1/SjM5BGz6cs0CQQDZeW4fym90qUOGeGTrqju1PVTtxnDGIqHkY2N4kbCP
tHdqbzZNGhxUIy2WNO9v2KYVvEOywvHGStMcop4D6yDXAkEA2ZGthTzk6MeiiwMg
Cq/AAEOgX+OiwZ9iwzUDTX/91l6c2bXioByHgwDqYAcGmqEJDlErS5oMdpyGfv2B
/zJRhwJARdq+Z9HDiUqRWRE1AYnV0fqYXCQAt3QKYm0WV3UcrJxAO1zrqUp4zQHb
s8LfIiMJ/jNR34rE1HfWZf1KGmIdUwJADJGX3pyX9MKjpzg0/6kLhHhjqWZzHpBg
mjpTyIReW6X3lbQmNW2wfmbtI0MEpKYs6cDSqXlqwudj9a4bdmynvQJAXwDzClYE
nvQ/mo4fIOrOItYGUqB3RAmwdawRtAq/w3fiJ+6yTNUZUZnTPf5ATY377Sdsjv5S
yuQSTVqq8SNJJA==
-----END PRIVATE KEY-----
`)
	certFileData = []byte(`
-----BEGIN CERTIFICATE-----
MIICEjCCAXugAwIBAgIRAI5eXpJ842d0UxYz0z3AB94wDQYJKoZIhvcNAQELBQAw
EjEQMA4GA1UEChMHQWNtZSBDbzAgFw03MDAxMDEwMDAwMDBaGA8yMDg0MDEyOTE2
MDAwMFowEjEQMA4GA1UEChMHQWNtZSBDbzCBnzANBgkqhkiG9w0BAQEFAAOBjQAw
gYkCgYEAuNOviJHWwQs24v3XL//llArN/MoOrGS5clvbycfNiWsmNrSGyzsi6NIg
QyIFQ6xG7XhM8qYK5WzfZClKBydt0SNx/CxDwuU3Y8l5zL4HA3C//+nkKJ6SVRjW
FIECtqf1jvmZrnUoAy9XoW1yaNLz+JGHORuimjlCmRPBiwlgWGECAwEAAaNmMGQw
DgYDVR0PAQH/BAQDAgKkMBMGA1UdJQQMMAoGCCsGAQUFBwMBMA8GA1UdEwEB/wQF
MAMBAf8wLAYDVR0RBCUwI4IJbG9jYWxob3N0hwR/AAABhxAAAAAAAAAAAAAAAAAA
AAABMA0GCSqGSIb3DQEBCwUAA4GBABcGvfOZd3nU5MTi4i9OhPLoZoMmrLED1scM
XYJ48XMFgWBSjtYAWMKhin2tCLNsm0JKbragbhFH/va42OfQjarAaJvIGpMIEcvT
6iBMZSG2ZCysBKXbuZa4OYvXfRpaUN9NokCrPgc8GFLJMSYt/Dd93r/h9JPRHFXi
4l4rVVaB
-----END CERTIFICATE-----
`)
)

func TestMain(m *testing.M) {
	var err error
	tmpDir, err = ioutil.TempDir("", "clevergo")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tmpDir) // clean up

	certFile = filepath.Join(tmpDir, "cert.pem")
	if err := ioutil.WriteFile(certFile, certFileData, 0666); err != nil {
		log.Fatal(err)
	}
	keyFile = filepath.Join(tmpDir, "key.pem")
	if err := ioutil.WriteFile(certFile, keyFileData, 0666); err != nil {
		log.Fatal(err)
	}

	code := m.Run()
	os.Exit(code)
}

func TestNew(t *testing.T) {
	tests := []struct {
		addr string
	}{
		{":http"},
		{":8080"},
		{"/tmp/clevergo.sock"},
	}
	for _, test := range tests {
		app := New(test.addr)
		if app.Addr != test.addr {
			t.Errorf("expected address %q, got %q", test.addr, app.Addr)
		}
	}
}

func TestApplicationUse(t *testing.T) {
	tests := [][]Middleware{
		{echoMiddleware("one"), echoMiddleware("two")},
		{echoMiddleware("foo"), echoMiddleware("bar")},
	}
	for _, middlewares := range tests {
		app := New("")
		app.Use(middlewares...)
		if len(app.middlewares) != len(middlewares) {
			t.Fatalf("middlewares count doesn't match, expected %d, got %d", len(middlewares), len(app.middlewares))
		}

		handler1 := Chain(echoHandler(""), middlewares...)
		resp1 := httptest.NewRecorder()
		handler1.ServeHTTP(resp1, nil)
		handler2 := Chain(echoHandler(""), app.middlewares...)
		resp2 := httptest.NewRecorder()
		handler2.ServeHTTP(resp2, nil)
		for resp1.Body.String() != resp2.Body.String() {
			t.Errorf("failed to use middlewares, expected body %q, got %q", resp1.Body.String(), resp2.Body.String())
		}
	}
}

func TestApplicationCleanUp(t *testing.T) {
	cleanOne := false
	cleanTwo := false
	app := New("")
	app.RegisterOnCleanUp(func() { cleanOne = true })
	app.RegisterOnCleanUp(func() { cleanTwo = true })
	app.CleanUp()
	if !cleanOne {
		t.Error("failed to invoke clean up one")
	}
	if !cleanTwo {
		t.Error("failed to invoke clean up two")
	}
}

func TestApplicationListenAndServe(t *testing.T) {
	addr := "localhost:12345"
	body := "ListenAndServe"
	app := New(addr)
	app.Handle(http.MethodGet, "/", echoHandler(body))

	started := make(chan bool)
	go func() {
		started <- true
		app.ListenAndServe()
	}()

	<-started
	defer app.Close()

	req := httptest.NewRequest(http.MethodGet, "http://"+addr+"/", nil)
	resp := httptest.NewRecorder()
	app.ServeHTTP(resp, req)
	if resp.Body.String() != body {
		t.Errorf("expected body %q, got %q", body, resp.Body.String())
	}
}

func TestApplicationListenAndServeTLS(t *testing.T) {
	addr := "localhost:12345"
	body := "ListenAndServeTLS"
	app := New(addr)
	app.Handle(http.MethodGet, "/", echoHandler(body))

	started := make(chan bool)
	go func() {
		started <- true
		app.ListenAndServeTLS(certFile, keyFile)
	}()

	<-started
	defer app.Close()

	req, _ := http.NewRequest(http.MethodGet, "https://localhost:12345/", nil)
	resp := httptest.NewRecorder()
	app.ServeHTTP(resp, req)
	if resp.Body.String() != body {
		t.Errorf("expected body %q, got %q", body, resp.Body.String())
	}
}
func TestApplicationListenAndServeUnix(t *testing.T) {
	addr := filepath.Join(tmpDir, "socket.sock")
	body := "ListenAndServeUnix"
	app := New(addr)
	app.Handle(http.MethodGet, "/", echoHandler(body))

	started := make(chan bool)
	go func() {
		started <- true
		app.ListenAndServeUnix()
	}()

	<-started
	defer app.Close()

	client := http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", addr)
			},
		},
	}

	req, _ := http.NewRequest(http.MethodGet, "http://unix", nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to read from unix socket: %s", err)
	}
	actualBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %s", err)
	}
	if string(actualBody) != body {
		t.Errorf("expected body %q, got %q", body, string(actualBody))
	}
}
func TestApplicationListenAndServeUnixError(t *testing.T) {
	addr := "/invalid/socket/addr"
	app := New(addr)
	err := app.ListenAndServeUnix()
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestApplicationServeTLS(t *testing.T) {
	addr := "localhost:22222"
	body := "ServeTLS"
	app := New(addr)
	app.Handle(http.MethodGet, "/", echoHandler(body))

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	started := make(chan bool)
	go func() {
		started <- true
		app.ServeTLS(ln, certFile, keyFile)
	}()

	<-started
	defer app.Close()

	req, _ := http.NewRequest(http.MethodGet, "https://localhost:22222/", nil)
	resp := httptest.NewRecorder()
	app.ServeHTTP(resp, req)
	if resp.Body.String() != body {
		t.Errorf("expected body %q, got %q", body, resp.Body.String())
	}
}

func ExampleApplication() {
	// application is wrapper of Router and http.Server.
	app := New("localhost:8080")
	// clean up.
	defer app.CleanUp()

	// here is a simple server header middleware, just for showing the use case of middleware.
	var serverHeaderMiddleware Middleware = func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Server", "clevergo")
			next.ServeHTTP(w, r)
		})
	}

	// initialize components, such as database.
	/*
		db, err := sql.Open("", "")
		if err != nil {
			log.Fatal(err)
		}
		// these functions will be called in order before closing application,
		// it equals to defer db.Close(), one benefits is that it allows you
		// to registers these functions in any place, makes main function more
		// clearer.
		app.RegisterOnCleanUp(func() {
			db.Close()
		})
	*/

	// use middlewares, global middleware that apply for all routes.
	// it is easy to use third-party middleware, such as recovery, compress and
	// logging middleware provided by clevergo middleware, gorilla handlers
	// and other third-party packages.
	app.Use(
		// handlers.RecoveryHandler(), // provided by gorilla/handlers.
		// middleware.Compress(gzip.DefaultCompression), // provided by clevergo middleware.
		// middleware.Logging(os.Stdout), // provided by clevergo middleware.
		serverHeaderMiddleware,
	)

	// registers routes
	app.Get("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "hello world")
	})
	app.Get("/hello/:name", func(w http.ResponseWriter, r *http.Request) {
		// retrieves parameters.
		ps := GetParams(r)
		fmt.Fprintf(w, "hello %s", ps.Get("name"))
	})

	// APIs
	api := app.Group("/api", RouteGroupMiddleware(
	// middlewares for API group, such as authentication, CORS, rate limiter etc...
	))

	// nested routes group
	v1 := api.Group("/v1", RouteGroupMiddleware(
	// middlewares for v1 group
	))
	// RESTful APIs
	v1.Get(
		"/users",
		func(w http.ResponseWriter, r *http.Request) {},
		// middlewares for current route, such as request body limit.
		RouteMiddleware(),
	)
	v1.Post("/users", func(w http.ResponseWriter, r *http.Request) {})
	v1.Get("/users/:name", func(w http.ResponseWriter, r *http.Request) {})
	v1.Put("/users/:name", func(w http.ResponseWriter, r *http.Request) {})
	v1.Delete("/users/:name", func(w http.ResponseWriter, r *http.Request) {})

	// v2 etc...
	// v2 := api.Group("/v2")
	// ...

	log.Fatal(app.ListenAndServe())
	// or log.Fatal(app.ListenAndServeTLS(certFile, keyFile))
	// or log.Fatal(app.ListenAndServeUnix())
}
