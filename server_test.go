// Copyright 2016 The Gem Authors. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package gem

import (
	"fmt"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/go-gem/sessions"
)

func TestVersion(t *testing.T) {
	if version != Version() {
		t.Errorf("expected version: %q, got %q.\n", version, Version())
	}
}

func TestName(t *testing.T) {
	if name != Name() {
		t.Errorf("expected name: %q, got %q.\n", name, Name())
	}
}

var emptyHandler = func(ctx *Context) {}

func TestServer_SetLogger(t *testing.T) {
	var logger Logger
	srv := New("", emptyHandler)
	srv.SetLogger(logger)
	if srv.logger != logger {
		t.Errorf("failed to set logger")
	}
}

func TestServer_SetSessionsStoret(t *testing.T) {
	var store sessions.Store
	srv := New("", emptyHandler)
	srv.SetSessionsStore(store)
	if srv.sessionsStore != store {
		t.Errorf("failed to set sessions store")
	}
}

func TestServer_SetSignal(t *testing.T) {
	var signals = map[os.Signal]Signal{
		syscall.SIGHUP:  SignalRestart,
		syscall.SIGUSR1: SignalRestart,
		syscall.SIGTERM: SignalShutdown,
		syscall.SIGUSR2: SignalRestart,
	}

	srv := New("", emptyHandler)

	for sig1, sig2 := range signals {
		srv.SetSignal(sig1, sig2)
	}

	for sig1, sig2 := range signals {
		if srv.signals[sig1] != sig2 {
			t.Errorf("expected signal %+v, got %+v", sig2, srv.signals[sig1])
		}
	}

	// invalid signal
	err := srv.SetSignal(syscall.SIGINT, -1)
	expectedErr := fmt.Sprintf("unsupported signal: %v", -1)
	if err == nil || err.Error() != expectedErr {
		t.Errorf("excepted error: %q, got %q", expectedErr, err)
	}
}

func TestServer_Init(t *testing.T) {
	addrs := ":8080,:4343,:6060"
	os.Setenv("GEM_SERVER_ADDRS", addrs)
	initServersFdOffset()
	if len(serversFdOffset) != 3 {
		t.Fatalf("expected length of serversFdOffset: %d, got %d", 3, len(serversFdOffset))
	}
	if serversFdOffset[":8080"] != 0 {
		t.Error(`expected serversFdOffset[":8080"] == 0, got false`)
	}
	if serversFdOffset[":4343"] != 1 {
		t.Error(`expected serversFdOffset[":4343"] == 1, got false`)
	}
	if serversFdOffset[":6060"] != 2 {
		t.Error(`expected serversFdOffset[":6060"] == 2, got false`)
	}
}

func TestServer_SetWaitTimeout(t *testing.T) {
	timeout := time.Minute

	srv := New("", emptyHandler)
	srv.SetWaitTimeout(timeout)

	if srv.waitTimeout != timeout {
		t.Errorf(`expected waitTimeout: %v, got %v`, timeout, srv.waitTimeout)
	}
}

func TestServer_LoadConfig(t *testing.T) {
	srv := New("", emptyHandler)

	config := &ServerConfig{
		Name:                 "fasthttp",
		Concurrency:          10000,
		DisableKeepalive:     true,
		ReadBufferSize:       1024,
		WriteBufferSize:      1024,
		ReadTimeout:          time.Second,
		WriteTimeout:         time.Second * 2,
		MaxConnsPerIP:        10,
		MaxRequestsPerConn:   100,
		MaxKeepaliveDuration: time.Hour,
		MaxRequestBodySize:   1024,
		ReduceMemoryUsage:    true,
		GetOnly:              true,
		DisableHeaderNamesNormalizing: true,
	}

	srv.LoadConfig(config)

	if srv.server.Name != config.Name {
		t.Errorf("expected server Name %v, got %v", config.Name, srv.server.Name)
	}
	if srv.server.Concurrency != config.Concurrency {
		t.Errorf("expected server Concurrency %v, got %v", config.Concurrency, srv.server.Concurrency)
	}
	if srv.server.DisableKeepalive != config.DisableKeepalive {
		t.Errorf("expected server DisableKeepalive %v, got %v", config.DisableKeepalive, srv.server.DisableKeepalive)
	}
	if srv.server.ReadBufferSize != config.ReadBufferSize {
		t.Errorf("expected server ReadBufferSize %v, got %v", config.ReadBufferSize, srv.server.ReadBufferSize)
	}
	if srv.server.WriteBufferSize != config.WriteBufferSize {
		t.Errorf("expected server WriteBufferSize %v, got %v", config.WriteBufferSize, srv.server.WriteBufferSize)
	}
	if srv.server.ReadTimeout != config.ReadTimeout {
		t.Errorf("expected server ReadTimeout %v, got %v", config.ReadTimeout, srv.server.ReadTimeout)
	}
	if srv.server.WriteTimeout != config.WriteTimeout {
		t.Errorf("expected server WriteTimeout %v, got %v", config.WriteTimeout, srv.server.WriteTimeout)
	}
	if srv.server.MaxConnsPerIP != config.MaxConnsPerIP {
		t.Errorf("expected server MaxConnsPerIP %v, got %v", config.MaxConnsPerIP, srv.server.MaxConnsPerIP)
	}
	if srv.server.MaxRequestsPerConn != config.MaxRequestsPerConn {
		t.Errorf("expected server MaxRequestsPerConn %v, got %v", config.MaxRequestsPerConn, srv.server.MaxRequestsPerConn)
	}
	if srv.server.MaxKeepaliveDuration != config.MaxKeepaliveDuration {
		t.Errorf("expected server MaxKeepaliveDuration %v, got %v", config.MaxKeepaliveDuration, srv.server.MaxKeepaliveDuration)
	}
	if srv.server.MaxRequestBodySize != config.MaxRequestBodySize {
		t.Errorf("expected server MaxRequestBodySize %v, got %v", config.MaxRequestBodySize, srv.server.MaxRequestBodySize)
	}
	if srv.server.ReduceMemoryUsage != config.ReduceMemoryUsage {
		t.Errorf("expected server ReduceMemoryUsage %v, got %v", config.ReduceMemoryUsage, srv.server.ReduceMemoryUsage)
	}
	if srv.server.GetOnly != config.GetOnly {
		t.Errorf("expected server GetOnly %v, got %v", config.GetOnly, srv.server.GetOnly)
	}
	if srv.server.DisableHeaderNamesNormalizing != config.DisableHeaderNamesNormalizing {
		t.Errorf("expected server DisableHeaderNamesNormalizing %v, got %v", config.DisableHeaderNamesNormalizing, srv.server.DisableHeaderNamesNormalizing)
	}
}

func TestServer_Other(t *testing.T) {
	srv := New("", emptyHandler)
	srv.stop()
	if srv.server.DisableKeepalive != true {
		t.Errorf("expected server DisableKeepalive %v, got %v", true, srv.server.DisableKeepalive)
	}
}
