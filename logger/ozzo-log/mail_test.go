// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package log_test

import (
	"testing"

	"mqant/logger/ozzo-log"
)

func TestNewMailTarget(t *testing.T) {
	target := log.NewMailTarget()
	if target.MaxLevel != log.LevelDebug {
		t.Errorf("NewMailTarget.MaxLevel = %v, expected %v", target.MaxLevel, log.LevelDebug)
	}
}
