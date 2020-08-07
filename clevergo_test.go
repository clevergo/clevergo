// Copyright 2020 CleverGo. All rights reserved.
// Use of this source code is governed by a MIT style license that can be found
// in the LICENSE file.

package clevergo

import (
	stdlog "log"
	"os"
	"testing"

	"clevergo.tech/log"
	"github.com/stretchr/testify/assert"
)

func TestSetLogger(t *testing.T) {
	defaultLogger = nil
	logger := log.New(os.Stderr, "", stdlog.LstdFlags)
	SetLogger(logger)
	assert.Equal(t, logger, defaultLogger)

	assert.Panics(t, assert.PanicTestFunc(func() {
		SetLogger(nil)
	}))
}
