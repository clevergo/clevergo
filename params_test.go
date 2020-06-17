// Copyright 2013 Julien Schmidt. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package clevergo

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func ExampleParams() {
	router := New()
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
		assert.Equal(t, ps[i].Value, ps.String(ps[i].Key))
	}
	assert.Equal(t, "", ps.String("noKey"))
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
		val, err := ps.Int(name)
		assert.Nil(t, err)
		assert.Equal(t, value, val)
	}
	_, err := ps.Int("noKey")
	assert.NotNil(t, err)
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
		val, err := ps.Int64(name)
		assert.Nil(t, err)
		assert.Equal(t, value, val)
	}
	_, err := ps.Int64("noKey")
	assert.NotNil(t, err)
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
		val, err := ps.Uint64(name)
		assert.Nil(t, err)
		assert.Equal(t, value, val)
	}
	_, err := ps.Uint64("noKey")
	assert.NotNil(t, err)
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
		val, err := ps.Float64(name)
		assert.Nil(t, err)
		assert.Equal(t, value, val)
	}
	_, err := ps.Float64("noKey")
	assert.NotNil(t, err)
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
		val, err := ps.Bool(name)
		assert.Nil(t, err)
		assert.Equal(t, value, val)
	}
	_, err := ps.Bool("noKey")
	assert.NotNil(t, err)
}
