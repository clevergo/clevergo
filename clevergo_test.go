// Copyright 2020 CleverGo. All rights reserved.
// Use of this source code is governed by a MIT license that can be found
// in the LICENSE file.

package clevergo

import (
	"testing"

	"clevergo.tech/log"
	"github.com/stretchr/testify/assert"
)

func TestSetLogger(t *testing.T) {
	defaultLogger = nil
	logger := log.New()
	SetLogger(logger)
	assert.Equal(t, logger, defaultLogger)

	assert.Panics(t, assert.PanicTestFunc(func() {
		SetLogger(nil)
	}))
}
