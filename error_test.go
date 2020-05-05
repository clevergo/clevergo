// Copyright 2020 CleverGo. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package clevergo

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewError(t *testing.T) {
	tests := []struct {
		code int
		msg  string
	}{
		{http.StatusForbidden, "forbidden"},
		{http.StatusInternalServerError, "internal server error"},
	}

	for _, test := range tests {
		err := NewError(test.code, errors.New(test.msg))
		assert.Equal(t, test.code, err.Code)
		assert.Equal(t, test.msg, err.Error())
	}
}
