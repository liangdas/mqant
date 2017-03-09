// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package log implements logging with severity levels and message categories.
package log

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

// RFC5424 log message levels.
const (
	LevelEmergency Level = iota
	LevelAlert
	LevelCritical
	LevelError
	LevelWarning
	LevelNotice
	LevelInfo
	LevelDebug
)

// Level describes the level of a log message.
type Level int

// LevelNames maps log levels to names
var LevelNames = map[Level]string{
	LevelDebug:     "Debug",
	LevelInfo:      "Info",
	LevelNotice:    "Notice",
	LevelWarning:   "Warning",
	LevelError:     "Error",
	LevelCritical:  "Critical",
	LevelAlert:     "Alert",
	LevelEmergency: "Emergency",
}

// String returns the string representation of the log level
func (l Level) String() string {
	if name, ok := LevelNames[l]; ok {
		return name
	}
	return "Unknown"
}

// Entry represents a log entry.
type Entry struct {
	Level            Level
	Category         string
	Message          string
	Time             time.Time
	CallStack        string
	ShortFile        string
	FormattedMessage string
}

// String returns the string representation of the log entry
func (e *Entry) String() string {
	return e.FormattedMessage
}

// Target represents a target where the logger can send log messages to for further processing.
type Target interface {
	// Open prepares the target for processing log messages.
	// Open will be invoked when Logger.Open() is called.
	// If an error is returned, the target will be removed from the logger.
	// errWriter should be used to write errors found while processing log messages.
	Open(errWriter io.Writer) error
	// Process processes an incoming log message.
	Process(*Entry)
	// Close closes a target.
	// Close is called when Logger.Close() is called, which gives each target
	// a chance to flush the logged messages to their destination storage.
	Close()
}

// coreLogger maintains the log messages in a channel and sends them to various targets.
type coreLogger struct {
	lock    sync.Mutex
	open    bool        // whether the logger is open
	entries chan *Entry // log entries

	ErrorWriter     io.Writer // the writer used to write errors caused by log targets
	BufferSize      int       // the size of the channel storing log entries
	CallStackDepth  int       // the number of call stack frames to be logged for each message. 0 means do not log any call stack frame.
	CallStackFilter string    // a substring that a call stack frame file path should contain in order for the frame to be counted
	MaxLevel        Level     // the maximum level of messages to be logged
	Targets         []Target  // targets for sending log messages to
}

// Formatter formats a log message into an appropriate string.
type Formatter func(*Logger, *Entry) string

// Logger records log messages and dispatches them to various targets for further processing.
type Logger struct {
	*coreLogger
	Category  string    // the category associated with this logger
	Formatter Formatter // message formatter
}

// NewLogger creates a root logger.
// The new logger takes these default options:
// ErrorWriter: os.Stderr, BufferSize: 1024, MaxLevel: LevelDebug,
// Category: app, Formatter: DefaultFormatter
func NewLogger() *Logger {
	logger := &coreLogger{
		ErrorWriter: os.Stderr,
		BufferSize:  1024,
		MaxLevel:    LevelDebug,
		Targets:     make([]Target, 0),
	}
	return &Logger{logger, "app", DefaultFormatter}
}

// GetLogger creates a logger with the specified category and log formatter.
// Messages logged through this logger will carry the same category name.
// The formatter, if not specified, will inherit from the calling logger.
// It will be used to format all messages logged through this logger.
func (l *Logger) GetLogger(category string, formatter ...Formatter) *Logger {
	if len(formatter) > 0 {
		return &Logger{l.coreLogger, category, formatter[0]}
	}
	return &Logger{l.coreLogger, category, l.Formatter}
}

// Emergency logs a message indicating the system is unusable.
// Please refer to Error() for how to use this method.
func (l *Logger) Emergency(format string, a ...interface{}) {
	l.Log(LevelEmergency, format, a...)
}

// Alert logs a message indicating action must be taken immediately.
// Please refer to Error() for how to use this method.
func (l *Logger) Alert(format string, a ...interface{}) {
	l.Log(LevelAlert, format, a...)
}

// Critical logs a message indicating critical conditions.
// Please refer to Error() for how to use this method.
func (l *Logger) Critical(format string, a ...interface{}) {
	l.Log(LevelCritical, format, a...)
}

// Error logs a message indicating an error condition.
// This method takes one or multiple parameters. If a single parameter
// is provided, it will be treated as the log message. If multiple parameters
// are provided, they will be passed to fmt.Sprintf() to generate the log message.
func (l *Logger) Error(format string, a ...interface{}) {
	l.Log(LevelError, format, a...)
}

// Warning logs a message indicating a warning condition.
// Please refer to Error() for how to use this method.
func (l *Logger) Warning(format string, a ...interface{}) {
	l.Log(LevelWarning, format, a...)
}

// Notice logs a message meaning normal but significant condition.
// Please refer to Error() for how to use this method.
func (l *Logger) Notice(format string, a ...interface{}) {
	l.Log(LevelNotice, format, a...)
}

// Info logs a message for informational purpose.
// Please refer to Error() for how to use this method.
func (l *Logger) Info(format string, a ...interface{}) {
	l.Log(LevelInfo, format, a...)
}

// Debug logs a message for debugging purpose.
// Please refer to Error() for how to use this method.
func (l *Logger) Debug(format string, a ...interface{}) {
	l.Log(LevelDebug, format, a...)
}

// Log logs a message of a specified severity level.
func (l *Logger) Log(level Level, format string, a ...interface{}) {
	if level > l.MaxLevel || !l.open {
		return
	}
	message := format
	if len(a) > 0 {
		message = fmt.Sprintf(format, a...)
	}
	entry := &Entry{
		Category: l.Category,
		Level:    level,
		Message:  message,
		Time:     time.Now(),
	}
	if l.CallStackDepth > 0 {
		entry.CallStack, entry.ShortFile = GetCallStack(4, l.CallStackDepth, l.CallStackFilter)
	}
	entry.FormattedMessage = l.Formatter(l, entry)
	l.entries <- entry
}

// Open prepares the logger and the targets for logging purpose.
// Open must be called before any message can be logged.
func (l *coreLogger) Open() error {
	l.lock.Lock()
	defer l.lock.Unlock()

	if l.open {
		return nil
	}

	if l.ErrorWriter == nil {
		return errors.New("Logger.ErrorWriter must be set.")
	}
	if l.BufferSize < 0 {
		return errors.New("Logger.BufferSize must be no less than 0.")
	}
	if l.CallStackDepth < 0 {
		return errors.New("Logger.CallStackDepth must be no less than 0.")
	}

	l.entries = make(chan *Entry, l.BufferSize)
	var targets []Target
	for _, target := range l.Targets {
		if err := target.Open(l.ErrorWriter); err != nil {
			fmt.Fprintf(l.ErrorWriter, "Failed to open target: %v", err)
		} else {
			targets = append(targets, target)
		}
	}
	l.Targets = targets

	go l.process()

	l.open = true

	return nil
}

// process sends the messages to targets for processing.
func (l *coreLogger) process() {
	for {
		entry := <-l.entries
		for _, target := range l.Targets {
			target.Process(entry)
		}
		if entry == nil {
			break
		}
	}
}

// Close closes the logger and the targets.
// Existing messages will be processed before the targets are closed.
// New incoming messages will be discarded after calling this method.
func (l *coreLogger) Close() {
	if !l.open {
		return
	}
	l.open = false
	// use a nil entry to signal the close of logger
	l.entries <- nil
	for _, target := range l.Targets {
		target.Close()
	}
}

// DefaultFormatter is the default formatter used to format every log message.
func DefaultFormatter(l *Logger, e *Entry) string {
	return fmt.Sprintf("%v [%v][%v] %v%v", e.Time.Format(time.RFC3339), e.Level, e.Category, e.Message, e.CallStack)
}

// GetCallStack returns the current call stack information as a string.
// The skip parameter specifies how many top frames should be skipped, while
// the frames parameter specifies at most how many frames should be returned.
func GetCallStack(skip int, frames int, filter string) (CallStack, sf string) {
	buf := new(bytes.Buffer)
	for i, count := skip, 0; count < frames; i++ {
		_, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		moduleOf := moduleOf(file)
		shortfile := shortfile(file)
		if sf == "" {
			sf = fmt.Sprintf("%s:%d", shortfile, line)
		}
		file = moduleOf + "/" + shortfile
		if filter == "" || strings.Contains(file, filter) {
			fmt.Fprintf(buf, "\n%s:%d", file, line)
			count++
		}
	}
	return buf.String(), sf
}
func moduleOf(file string) string {
	pos := strings.LastIndex(file, "/")
	if pos != -1 {
		pos1 := strings.LastIndex(file[:pos], "/src/")
		if pos1 != -1 {
			return file[pos1+5 : pos]
		}
	}
	return "UNKNOWN"
}
func shortfile(file string) string {
	short := file
	for i := len(file) - 1; i > 0; i-- {
		if file[i] == '/' {
			short = file[i+1:]
			break
		}
	}
	file = short
	return file
}
