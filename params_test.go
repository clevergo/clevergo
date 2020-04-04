// Copyright 2013 Julien Schmidt. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package clevergo

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func ExampleParams() {
	router := NewRouter()
	router.Get("/post/:year/:month/:title", func(ctx *Context) error {
		// converts param value to int.
		year, _ := ctx.Params.Int("year")
		month, _ := ctx.Params.Int("month")
		// ps.Int64("name") // converts to int64.
		// ps.Uint64("name") // converts to uint64.
		// ps.Float64("name") // converts to float64.
		// ps.Bool("name") // converts to boolean.
		fmt.Printf("%s posted on %04d-%02d\n", ctx.Params.String("title"), year, month)
		return nil
	})
	req := httptest.NewRequest(http.MethodGet, "/post/2020/01/foo", nil)
	router.ServeHTTP(nil, req)

	req = httptest.NewRequest(http.MethodGet, "/post/2020/02/bar", nil)
	router.ServeHTTP(nil, req)

	// Output:
	// foo posted on 2020-01
	// bar posted on 2020-02
}

func TestParams(t *testing.T) {
	ps := Params{
		Param{"param1", "value1"},
		Param{"param2", "value2"},
		Param{"param3", "value3"},
	}
	for i := range ps {
		if val := ps.String(ps[i].Key); val != ps[i].Value {
			t.Errorf("Wrong value for %s: Got %s; Want %s", ps[i].Key, val, ps[i].Value)
		}
	}
	if val := ps.String("noKey"); val != "" {
		t.Errorf("Expected empty string for not found key; got %q", val)
	}
}

func TestParams_Int(t *testing.T) {
	ps := Params{
		Param{"param1", "-1"},
		Param{"param2", "0"},
		Param{"param3", "1"},
	}
	tests := map[string]int{
		"param1": -1,
		"param2": 0,
		"param3": 1,
	}
	for name, value := range tests {
		if val, err := ps.Int(name); err != nil || val != value {
			t.Errorf("Wrong value for %s: Got %d; Want %d", name, val, value)
		}
	}
	if val, err := ps.Int("noKey"); err == nil {
		t.Errorf("Expected an error for not found key; got %d", val)
	}
}

func TestParams_Int64(t *testing.T) {
	ps := Params{
		Param{"param1", "-1"},
		Param{"param2", "0"},
		Param{"param3", "1"},
	}
	tests := map[string]int64{
		"param1": -1,
		"param2": 0,
		"param3": 1,
	}
	for name, value := range tests {
		if val, err := ps.Int64(name); err != nil || val != value {
			t.Errorf("Wrong value for %s: Got %d; Want %d", name, val, value)
		}
	}
	if val, err := ps.Int64("noKey"); err == nil {
		t.Errorf("Expected an error for not found key; got %d", val)
	}
}

func TestParams_Uint64(t *testing.T) {
	ps := Params{
		Param{"param1", "0"},
		Param{"param2", "1"},
	}
	tests := map[string]uint64{
		"param1": 0,
		"param2": 1,
	}
	for name, value := range tests {
		if val, err := ps.Uint64(name); err != nil || val != value {
			t.Errorf("Wrong value for %s: Got %d; Want %d", name, val, value)
		}
	}
	if val, err := ps.Uint64("noKey"); err == nil {
		t.Errorf("Expected an error for not found key; got %d", val)
	}
}

func TestParams_Float(t *testing.T) {
	ps := Params{
		Param{"param1", "-0.2"},
		Param{"param2", "0.2"},
		Param{"param3", "1.9"},
	}
	tests := map[string]float64{
		"param1": -0.2,
		"param2": 0.2,
		"param3": 1.9,
	}
	for name, value := range tests {
		if val, err := ps.Float64(name); err != nil || val != value {
			t.Errorf("Wrong value for %s: Got %f; Want %f", name, val, value)
		}
	}
	if val, err := ps.Float64("noKey"); err == nil {
		t.Errorf("Expected an error for not found key; got %f", val)
	}
}

func TestParams_Bool(t *testing.T) {
	ps := Params{
		Param{"param1", "1"},
		Param{"param2", "t"},
		Param{"param3", "T"},
		Param{"param4", "true"},
		Param{"param5", "TRUE"},
		Param{"param6", "True"},
		Param{"param7", "0"},
		Param{"param8", "f"},
		Param{"param9", "F"},
		Param{"param10", "false"},
		Param{"param11", "FALSE"},
		Param{"param12", "False"},
	}
	tests := map[string]bool{
		"param1":  true,
		"param2":  true,
		"param3":  true,
		"param4":  true,
		"param5":  true,
		"param6":  true,
		"param7":  false,
		"param8":  false,
		"param9":  false,
		"param10": false,
		"param11": false,
		"param12": false,
	}
	for name, value := range tests {
		if val, err := ps.Bool(name); err != nil || val != value {
			t.Errorf("Wrong value for %s: Got %t; Want %t", name, val, value)
		}
	}
	if val, err := ps.Bool("noKey"); err == nil {
		t.Errorf("Expected an error for not found key; got %t", val)
	}
}
