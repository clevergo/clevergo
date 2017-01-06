// Copyright 2016 The Gem Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license
// that can be found in the LICENSE file.

package gem

import (
	"fmt"
	"strconv"
)

// String convert v to string.
//
// By default, empty string and non nil error
// would be returned.
func String(v interface{}) (string, error) {
	switch value := v.(type) {
	case string:
		return value, nil
	case int:
		return strconv.Itoa(value), nil
	case []byte:
		return string(value), nil
	}

	return "", fmt.Errorf("unsupport to convert type %T to string", v)
}

// Int convert v to int.
//
// By default, zero and non nil error
// would be returned.
func Int(v interface{}) (int, error) {
	switch value := v.(type) {
	case int:
		return value, nil
	case string:
		return strconv.Atoi(v.(string))
	}

	return 0, fmt.Errorf("unsupport to convert type %T to int", v)
}
