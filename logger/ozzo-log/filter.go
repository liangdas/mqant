// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package log

import (
	"strings"
)

// Filter checks if a log message meets the level and category requirements.
type Filter struct {
	catNames    map[string]bool
	catPrefixes []string

	MaxLevel   Level    // the maximum severity level that is allowed
	MinLevel   Level    // the minimum severity level that is allowed
	Categories []string // the allowed message categories. Categories can use "*" as a suffix for wildcard matching.
}

// Init initializes the filter.
// Init must be called before Allow is called.
func (t *Filter) Init() {
	t.catNames = make(map[string]bool, 0)
	t.catPrefixes = make([]string, 0)
	for _, cat := range t.Categories {
		if strings.HasSuffix(cat, "*") {
			t.catPrefixes = append(t.catPrefixes, cat[:len(cat)-1])
		} else {
			t.catNames[cat] = true
		}
	}
}

// Allow checks if a message meets the severity level and category requirements.
func (t *Filter) Allow(e *Entry) bool {
	if e == nil {
		return true
	}
	if e.Level > t.MaxLevel {
		return false
	}
	if e.Level < t.MinLevel {
		return false
	}
	if t.catNames[e.Category] {
		return true
	}
	for _, cat := range t.catPrefixes {
		if strings.HasPrefix(e.Category, cat) {
			return true
		}
	}
	return len(t.catNames) == 0 && len(t.catPrefixes) == 0
}
