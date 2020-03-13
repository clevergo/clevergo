// Copyright 2020 CleverGo. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package clevergo

import (
	"errors"
	"net/http"
	"testing"
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
		if err.Code != test.code {
			t.Errorf("expected error code %d, got %d", test.code, err.Code)
		}
		if err.Error() != test.msg {
			t.Errorf("expected error message %s, got %s", test.msg, err.Error())
		}
	}
}
