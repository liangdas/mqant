// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package log_test

import (
	"strings"
	"testing"

	"mqant/logger/ozzo-log"
)

func TestFilterAllow(t *testing.T) {
	tests := []struct {
		cats     []string
		cat      string
		expected bool
	}{
		{[]string{}, "", true},
		{[]string{}, "system", true},
		{[]string{"system"}, "", false},
		{[]string{"system"}, "system", true},
		{[]string{"system"}, "system.db", false},
		{[]string{"system.*"}, "", false},
		{[]string{"system.*"}, "system", false},
		{[]string{"system.*"}, "system.", true},
		{[]string{"system.*"}, "system.db", true},
	}
	for _, test := range tests {
		filter := log.Filter{MaxLevel: log.LevelDebug, Categories: test.cats}
		filter.Init()
		e := &log.Entry{Category: test.cat}
		if filter.Allow(e) != test.expected {
			t.Errorf("filter(%q).Allow(%q) = %v, expected %v", strings.Join(test.cats, ","), test.cat, filter.Allow(e), test.expected)
		}
	}
}
