// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package log

import (
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
)

type consoleBrush func(string) string

func newConsoleBrush(format string) consoleBrush {
	return func(text string) string {
		return "\033[" + format + "m" + text + "\033[0m"
	}
}

var brushes = map[Level]consoleBrush{
	LevelDebug:     newConsoleBrush("39"),   // default
	LevelInfo:      newConsoleBrush("32"),   // green
	LevelNotice:    newConsoleBrush("36"),   // cyan
	LevelWarning:   newConsoleBrush("33"),   // yellow
	LevelError:     newConsoleBrush("31"),   // red
	LevelCritical:  newConsoleBrush("35"),   // magenta
	LevelAlert:     newConsoleBrush("1;91"), // bold light red
	LevelEmergency: newConsoleBrush("1;95"), // bold light magenta
}

// ConsoleTarget writes filtered log messages to console window.
type ConsoleTarget struct {
	*Filter
	ColorMode bool      // whether to use colors to differentiate log levels
	Writer    io.Writer // the writer to write log messages
	close     chan bool
}

// NewConsoleTarget creates a ConsoleTarget.
// The new ConsoleTarget takes these default options:
// MaxLevel: LevelDebug, ColorMode: true, Writer: os.Stdout
func NewConsoleTarget() *ConsoleTarget {
	return &ConsoleTarget{
		Filter:    &Filter{MaxLevel: LevelDebug},
		ColorMode: true,
		Writer:    os.Stdout,
		close:     make(chan bool, 0),
	}
}

// Open prepares ConsoleTarget for processing log messages.
func (t *ConsoleTarget) Open(io.Writer) error {
	t.Filter.Init()
	if t.Writer == nil {
		return errors.New("ConsoleTarget.Writer cannot be nil")
	}
	if runtime.GOOS == "windows" {
		t.ColorMode = false
	}
	return nil
}

// Process writes a log message using Writer.
func (t *ConsoleTarget) Process(e *Entry) {
	if e == nil {
		t.close <- true
		return
	}
	if !t.Allow(e) {
		return
	}
	msg := e.String()
	if t.ColorMode {
		brush, ok := brushes[e.Level]
		if ok {
			msg = brush(msg)
		}
	}
	fmt.Fprintln(t.Writer, msg)
}

// Close closes the console target.
func (t *ConsoleTarget) Close() {
	<-t.close
}
