// Copyright 2014 beego Author. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package logs provide a general log interface
// Usage:
//
// import "github.com/astaxie/beego/logs"
//
//	log := NewLogger(10000)
//	log.SetLogger("console", "")
//
//	> the first params stand for how many channel
//
// Use it like this:
//
//	log.Trace("trace")
//	log.Info("info")
//	log.Warn("warning")
//	log.Debug("debug")
//	log.Critical("critical")
//
//  more docs http://beego.me/docs/module/logs.md
package logs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

// RFC5424 log message levels.
const (
	LevelEmergency = iota
	LevelAlert
	LevelCritical
	LevelError
	LevelWarning
	LevelNotice
	LevelInformational
	LevelDebug
)

// levelLogLogger is defined to implement log.Logger
// the real log level will be LevelEmergency
const levelLoggerImpl = -1

// Name for adapter with beego official support
const (
	AdapterConsole   = "console"
	AdapterFile      = "file"
	AdapterMultiFile = "multifile"
	AdapterMail      = "smtp"
	AdapterConn      = "conn"
	AdapterEs        = "es"
	AdapterJianLiao  = "jianliao"
	AdapterSlack     = "slack"
	AdapterDingtalk  = "dingtalk"
	AdapterAliLS     = "alils"
)

// Legacy log level constants to ensure backwards compatibility.
const (
	LevelInfo  = LevelInformational
	LevelTrace = LevelDebug
	LevelWarn  = LevelWarning
)

type newLoggerFunc func() Logger

// Logger defines the behavior of a log provider.
type Logger interface {
	Init(config string) error
	WriteMsg(when time.Time, msg string, level int) error
	WriteOriginalMsg(when time.Time, msg string, level int) error
	Destroy()
	Flush()
}

var adapters = make(map[string]newLoggerFunc)
var levelPrefix = [LevelDebug + 1]string{"[M] ", "[A] ", "[C] ", "[E] ", "[W] ", "[N] ", "[I] ", "[D] "}

// Register makes a log provide available by the provided name.
// If Register is called twice with the same name or if driver is nil,
// it panics.
func Register(name string, log newLoggerFunc) {
	if log == nil {
		panic("logs: Register provide is nil")
	}
	if _, dup := adapters[name]; dup {
		panic("logs: Register called twice for provider " + name)
	}
	adapters[name] = log
}

// BeeLogger is default logger in beego application.
// it can contain several providers and log message into all providers.
type BeeLogger struct {
	lock                sync.Mutex
	level               int
	init                bool
	enableFuncCallDepth bool
	loggerFuncCallDepth int
	asynchronous        bool
	msgChanLen          int64
	contentType         string //text/plain application/json
	msgChan             chan *logMsg
	signalChan          chan string
	wg                  sync.WaitGroup
	outputs             []*nameLogger
	ProcessID           string
}

const defaultAsyncMsgLen = 1e3

type nameLogger struct {
	Logger
	name string
}

type logMsg struct {
	level    int
	original bool
	msg      string
	when     time.Time
}

var logMsgPool *sync.Pool

// NewLogger returns a new BeeLogger.
// channelLen means the number of messages in chan(used where asynchronous is true).
// if the buffering chan is full, logger adapters write to file or other way.
func NewLogger(channelLens ...int64) *BeeLogger {
	bl := new(BeeLogger)
	bl.level = LevelDebug
	bl.loggerFuncCallDepth = 2
	bl.contentType = "text/plain"
	bl.msgChanLen = append(channelLens, 0)[0]
	if bl.msgChanLen <= 0 {
		bl.msgChanLen = defaultAsyncMsgLen
	}
	bl.signalChan = make(chan string, 1)
	bl.setLogger(AdapterConsole)
	return bl
}

// Async set the log to asynchronous and start the goroutine
func (bl *BeeLogger) Async(msgLen ...int64) *BeeLogger {
	bl.lock.Lock()
	defer bl.lock.Unlock()
	if bl.asynchronous {
		return bl
	}
	bl.asynchronous = true
	if len(msgLen) > 0 && msgLen[0] > 0 {
		bl.msgChanLen = msgLen[0]
	}
	bl.msgChan = make(chan *logMsg, bl.msgChanLen)
	logMsgPool = &sync.Pool{
		New: func() interface{} {
			return &logMsg{}
		},
	}
	bl.wg.Add(1)
	go bl.startLogger()
	return bl
}

// SetLogger provides a given logger adapter into BeeLogger with config string.
// config need to be correct JSON as string: {"interval":360}.
func (bl *BeeLogger) setLogger(adapterName string, configs ...string) error {
	config := append(configs, "{}")[0]
	for _, l := range bl.outputs {
		if l.name == adapterName {
			return fmt.Errorf("logs: duplicate adaptername %q (you have set this logger before)", adapterName)
		}
	}

	log, ok := adapters[adapterName]
	if !ok {
		return fmt.Errorf("logs: unknown adaptername %q (forgotten Register?)", adapterName)
	}

	lg := log()
	err := lg.Init(config)
	if err != nil {
		fmt.Fprintln(os.Stderr, "logs.BeeLogger.SetLogger: "+err.Error())
		return err
	}
	bl.outputs = append(bl.outputs, &nameLogger{name: adapterName, Logger: lg})
	return nil
}

// SetLogger provides a given logger adapter into BeeLogger with config string.
// config need to be correct JSON as string: {"interval":360}.
func (bl *BeeLogger) SetLogger(adapterName string, configs ...string) error {
	bl.lock.Lock()
	defer bl.lock.Unlock()
	if !bl.init {
		bl.outputs = []*nameLogger{}
		bl.init = true
	}
	return bl.setLogger(adapterName, configs...)
}

// DelLogger remove a logger adapter in BeeLogger.
func (bl *BeeLogger) DelLogger(adapterName string) error {
	bl.lock.Lock()
	defer bl.lock.Unlock()
	outputs := []*nameLogger{}
	for _, lg := range bl.outputs {
		if lg.name == adapterName {
			lg.Destroy()
		} else {
			outputs = append(outputs, lg)
		}
	}
	if len(outputs) == len(bl.outputs) {
		return fmt.Errorf("logs: unknown adaptername %q (forgotten Register?)", adapterName)
	}
	bl.outputs = outputs
	return nil
}

func (bl *BeeLogger) writeToLoggers(original bool, when time.Time, msg string, level int) {
	for _, l := range bl.outputs {
		if original == true {
			err := l.WriteOriginalMsg(when, msg, level)
			if err != nil {
				fmt.Fprintf(os.Stderr, "unable to WriteMsg to adapter:%v,error:%v\n", l.name, err)
			}
		} else {
			err := l.WriteMsg(when, msg, level)
			if err != nil {
				fmt.Fprintf(os.Stderr, "unable to WriteMsg to adapter:%v,error:%v\n", l.name, err)
			}
		}

	}
}

func (bl *BeeLogger) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}
	// writeMsg will always add a '\n' character
	if p[len(p)-1] == '\n' {
		p = p[0 : len(p)-1]
	}
	// set levelLoggerImpl to ensure all log message will be write out
	err = bl.writeMsg(nil, levelLoggerImpl, string(p))
	if err == nil {
		return len(p), err
	}
	return 0, err
}

func (bl *BeeLogger) formatText(when time.Time, span *BeegoTraceSpan, logLevel int, msg string, v ...interface{}) (string, error) {
	if len(v) > 0 {
		msg = fmt.Sprintf(msg, v...)
	}
	if bl.enableFuncCallDepth {
		//_, file, line, ok := runtime.Caller(bl.loggerFuncCallDepth)
		//if !ok {
		//	file = "???"
		//	line = 0
		//}
		//_, filename := path.Split(file)
		//msg = "[" + filename + ":" + strconv.Itoa(line) + "] " + msg
		if logLevel <= LevelWarn {
			CallStack, ShortFile := GetCallStack(4, bl.loggerFuncCallDepth, "")
			msg = "[" + ShortFile + "] " + msg + " " + CallStack
		} else {
			_, ShortFile := GetCallStack(4, bl.loggerFuncCallDepth, "")
			msg = "[" + ShortFile + "] " + msg
		}

	}

	//set level info in front of filename info
	if logLevel == levelLoggerImpl {
		// set to emergency to ensure all log will be print out correctly
		logLevel = LevelEmergency
	} else {
		msg = "[" + bl.ProcessID + "] " + levelPrefix[logLevel] + msg
	}

	if span != nil {
		msg = " [" + span.Trace + "] " + "[" + span.Span + "] " + msg
	} else {
		msg = " [-] " + "[-] " + msg
	}
	return msg, nil
}

func (bl *BeeLogger) formatJson(when time.Time, span *BeegoTraceSpan, logLevel int, msg string, v ...interface{}) (string, error) {
	if len(v) > 0 {
		msg = fmt.Sprintf(msg, v...)
	}
	h, _ := formatTimeHeader(when)
	msgjson := map[string]interface{}{
		"message":    msg,
		"formattime": string(h),
		"timestamp":  when.UnixNano(),
	}
	if bl.enableFuncCallDepth {
		//_, file, line, ok := runtime.Caller(bl.loggerFuncCallDepth)
		//if !ok {
		//	file = "???"
		//	line = 0
		//}
		//_, filename := path.Split(file)
		//msg = "[" + filename + ":" + strconv.Itoa(line) + "] " + msg
		if logLevel <= LevelWarn {
			CallStack, ShortFile := GetCallStack(4, bl.loggerFuncCallDepth, "")
			msgjson["file"] = ShortFile
			msgjson["stack"] = CallStack
		} else {
			_, ShortFile := GetCallStack(4, bl.loggerFuncCallDepth, "")
			msgjson["file"] = ShortFile
			msgjson["stack"] = ""
		}

	}

	//set level info in front of filename info
	if logLevel == levelLoggerImpl {
		// set to emergency to ensure all log will be print out correctly
		logLevel = LevelEmergency
	} else {
		msgjson["processid"] = bl.ProcessID
		msgjson["level"] = levelPrefix[logLevel]
	}

	if span != nil {
		msgjson["trace_id"] = span.Trace
		msgjson["trace_span"] = span.Span
	} else {
		msgjson["trace_id"] = ""
		msgjson["trace_span"] = ""
	}
	msgbys, err := json.Marshal(msgjson)
	if err != nil {
		return "", err
	}
	return string(msgbys), nil
}

func (bl *BeeLogger) writeMsg(span *BeegoTraceSpan, logLevel int, msg string, v ...interface{}) error {
	if !bl.init {
		bl.lock.Lock()
		bl.setLogger(AdapterConsole)
		bl.lock.Unlock()
	}

	when := time.Now()
	original := false
	if bl.contentType == "application/json" {
		message, err := bl.formatJson(when, span, logLevel, msg, v...)
		if err != nil {
			return err
		}
		msg = message
		original = true
	} else {
		message, err := bl.formatText(when, span, logLevel, msg, v...)
		if err != nil {
			return err
		}
		msg = message
		original = false
	}
	if bl.asynchronous {
		lm := logMsgPool.Get().(*logMsg)
		lm.level = logLevel
		lm.msg = msg
		lm.when = when
		lm.original = original
		bl.msgChan <- lm
	} else {
		bl.writeToLoggers(original, when, msg, logLevel)
	}
	return nil
}

func (bl *BeeLogger) writeBiReport(msg string, logLevel int) error {
	if !bl.init {
		bl.lock.Lock()
		bl.setLogger(AdapterConsole)
		bl.lock.Unlock()
	}

	when := time.Now()
	if bl.asynchronous {
		lm := logMsgPool.Get().(*logMsg)
		lm.level = LevelError
		lm.msg = msg
		lm.original = true
		lm.when = when
		bl.msgChan <- lm
	} else {
		bl.writeToLoggers(true, when, msg, logLevel)
	}
	return nil
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

// SetLevel Set log message level.
// If message level (such as LevelDebug) is higher than logger level (such as LevelWarning),
// log providers will not even be sent the message.
func (bl *BeeLogger) SetLevel(l int) {
	bl.level = l
}

// SetLogFuncCallDepth set log funcCallDepth
func (bl *BeeLogger) SetLogFuncCallDepth(d int) {
	bl.loggerFuncCallDepth = d
}

// GetLogFuncCallDepth return log funcCallDepth for wrapper
func (bl *BeeLogger) GetLogFuncCallDepth() int {
	return bl.loggerFuncCallDepth
}

// EnableFuncCallDepth enable log funcCallDepth
func (bl *BeeLogger) EnableFuncCallDepth(b bool) {
	bl.enableFuncCallDepth = b
}

func (bl *BeeLogger) SetContentType(b string) {
	bl.contentType = b
}

// start logger chan reading.
// when chan is not empty, write logs.
func (bl *BeeLogger) startLogger() {
	gameOver := false
	for {
		select {
		case bm := <-bl.msgChan:
			bl.writeToLoggers(bm.original, bm.when, bm.msg, bm.level)
			logMsgPool.Put(bm)
		case sg := <-bl.signalChan:
			// Now should only send "flush" or "close" to bl.signalChan
			bl.flush()
			if sg == "close" {
				for _, l := range bl.outputs {
					l.Destroy()
				}
				bl.outputs = nil
				gameOver = true
			}
			bl.wg.Done()
		}
		if gameOver {
			break
		}
	}
}

// Emergency Log EMERGENCY level message.
func (bl *BeeLogger) Emergency(span *BeegoTraceSpan, format string, v ...interface{}) {
	if LevelEmergency > bl.level {
		return
	}
	bl.writeMsg(span, LevelEmergency, format, v...)
}

// Alert Log ALERT level message.
func (bl *BeeLogger) Alert(span *BeegoTraceSpan, format string, v ...interface{}) {
	if LevelAlert > bl.level {
		return
	}
	bl.writeMsg(span, LevelAlert, format, v...)
}

// Critical Log CRITICAL level message.
func (bl *BeeLogger) Critical(span *BeegoTraceSpan, format string, v ...interface{}) {
	if LevelCritical > bl.level {
		return
	}
	bl.writeMsg(span, LevelCritical, format, v...)
}

// Error Log ERROR level message.
func (bl *BeeLogger) Error(span *BeegoTraceSpan, format string, v ...interface{}) {
	if LevelError > bl.level {
		return
	}
	bl.writeMsg(span, LevelError, format, v...)
}

// Warning Log WARNING level message.
func (bl *BeeLogger) Warning(span *BeegoTraceSpan, format string, v ...interface{}) {
	if LevelWarn > bl.level {
		return
	}
	bl.writeMsg(span, LevelWarn, format, v...)
}

// Notice Log NOTICE level message.
func (bl *BeeLogger) Notice(span *BeegoTraceSpan, format string, v ...interface{}) {
	if LevelNotice > bl.level {
		return
	}
	bl.writeMsg(span, LevelNotice, format, v...)
}

// Informational Log INFORMATIONAL level message.
func (bl *BeeLogger) Informational(span *BeegoTraceSpan, format string, v ...interface{}) {
	if LevelInfo > bl.level {
		return
	}
	bl.writeMsg(span, LevelInfo, format, v...)
}

// Debug Log DEBUG level message.
func (bl *BeeLogger) Debug(span *BeegoTraceSpan, format string, v ...interface{}) {
	if LevelDebug > bl.level {
		return
	}
	bl.writeMsg(span, LevelDebug, format, v...)
}

// Warn Log WARN level message.
// compatibility alias for Warning()
func (bl *BeeLogger) Warn(span *BeegoTraceSpan, format string, v ...interface{}) {
	if LevelWarn > bl.level {
		return
	}
	bl.writeMsg(span, LevelWarn, format, v...)
}

// Info Log INFO level message.
// compatibility alias for Informational()
func (bl *BeeLogger) Info(span *BeegoTraceSpan, format string, v ...interface{}) {
	if LevelInfo > bl.level {
		return
	}
	bl.writeMsg(span, LevelInfo, format, v...)
}

// Trace Log TRACE level message.
// compatibility alias for Debug()
func (bl *BeeLogger) Trace(span *BeegoTraceSpan, format string, v ...interface{}) {
	if LevelDebug > bl.level {
		return
	}
	bl.writeMsg(span, LevelDebug, format, v...)
}

func (bl *BeeLogger) BiReport(msg string) {
	if LevelEmergency > bl.level {
		return
	}
	bl.writeBiReport(msg, LevelEmergency)
}

// Flush flush all chan data.
func (bl *BeeLogger) Flush() {
	if bl.asynchronous {
		bl.signalChan <- "flush"
		bl.wg.Wait()
		bl.wg.Add(1)
		return
	}
	bl.flush()
}

// Close close logger, flush all chan data and destroy all adapters in BeeLogger.
func (bl *BeeLogger) Close() {
	if bl.asynchronous {
		bl.signalChan <- "close"
		bl.wg.Wait()
		close(bl.msgChan)
	} else {
		bl.flush()
		for _, l := range bl.outputs {
			l.Destroy()
		}
		bl.outputs = nil
	}
	close(bl.signalChan)
}

// Reset close all outputs, and set bl.outputs to nil
func (bl *BeeLogger) Reset() {
	bl.Flush()
	for _, l := range bl.outputs {
		l.Destroy()
	}
	bl.outputs = nil
}

func (bl *BeeLogger) flush() {
	if bl.asynchronous {
		for {
			if len(bl.msgChan) > 0 {
				bm := <-bl.msgChan
				bl.writeToLoggers(bm.original, bm.when, bm.msg, bm.level)
				logMsgPool.Put(bm)
				continue
			}
			break
		}
	}
	for _, l := range bl.outputs {
		l.Flush()
	}
}

// beeLogger references the used application logger.
var beeLogger = NewLogger()

// GetBeeLogger returns the default BeeLogger
func GetBeeLogger() *BeeLogger {
	return beeLogger
}

var beeLoggerMap = struct {
	sync.RWMutex
	logs map[string]*log.Logger
}{
	logs: map[string]*log.Logger{},
}

// GetLogger returns the default BeeLogger
func GetLogger(prefixes ...string) *log.Logger {
	prefix := append(prefixes, "")[0]
	if prefix != "" {
		prefix = fmt.Sprintf(`[%s] `, strings.ToUpper(prefix))
	}
	beeLoggerMap.RLock()
	l, ok := beeLoggerMap.logs[prefix]
	if ok {
		beeLoggerMap.RUnlock()
		return l
	}
	beeLoggerMap.RUnlock()
	beeLoggerMap.Lock()
	defer beeLoggerMap.Unlock()
	l, ok = beeLoggerMap.logs[prefix]
	if !ok {
		l = log.New(beeLogger, prefix, 0)
		beeLoggerMap.logs[prefix] = l
	}
	return l
}

// Reset will remove all the adapter
func Reset() {
	beeLogger.Reset()
}

// Async set the beelogger with Async mode and hold msglen messages
func Async(msgLen ...int64) *BeeLogger {
	return beeLogger.Async(msgLen...)
}

// SetLevel sets the global log level used by the simple logger.
func SetLevel(l int) {
	beeLogger.SetLevel(l)
}

// EnableFuncCallDepth enable log funcCallDepth
func EnableFuncCallDepth(b bool) {
	beeLogger.enableFuncCallDepth = b
}

// SetLogFuncCall set the CallDepth, default is 4
func SetLogFuncCall(b bool) {
	beeLogger.EnableFuncCallDepth(b)
	beeLogger.SetLogFuncCallDepth(4)
}

// SetLogFuncCallDepth set log funcCallDepth
func SetLogFuncCallDepth(d int) {
	beeLogger.loggerFuncCallDepth = d
}

// SetLogger sets a new logger.
func SetLogger(adapter string, config ...string) error {
	return beeLogger.SetLogger(adapter, config...)
}

//
// Emergency logs a message at emergency level.
func Emergency(f interface{}, v ...interface{}) {
	beeLogger.Emergency(nil, formatLog(f, v...))
}

// Alert logs a message at alert level.
func Alert(f interface{}, v ...interface{}) {
	beeLogger.Alert(nil, formatLog(f, v...))
}

// Critical logs a message at critical level.
func Critical(f interface{}, v ...interface{}) {
	beeLogger.Critical(nil, formatLog(f, v...))
}

// Error logs a message at error level.
func Error(f interface{}, v ...interface{}) {
	beeLogger.Error(nil, formatLog(f, v...))
}

// Warning logs a message at warning level.
func Warning(f interface{}, v ...interface{}) {
	beeLogger.Warn(nil, formatLog(f, v...))
}

// Warn compatibility alias for Warning()
func Warn(f interface{}, v ...interface{}) {
	beeLogger.Warn(nil, formatLog(f, v...))
}

// Notice logs a message at notice level.
func Notice(f interface{}, v ...interface{}) {
	beeLogger.Notice(nil, formatLog(f, v...))
}

// Informational logs a message at info level.
func Informational(f interface{}, v ...interface{}) {
	beeLogger.Info(nil, formatLog(f, v...))
}

// Info compatibility alias for Warning()
func Info(f interface{}, v ...interface{}) {
	beeLogger.Info(nil, formatLog(f, v...))
}

// Debug logs a message at debug level.
func Debug(f interface{}, v ...interface{}) {
	beeLogger.Debug(nil, formatLog(f, v...))
}

// Trace logs a message at trace level.
// compatibility alias for Warning()
func Trace(f interface{}, v ...interface{}) {
	beeLogger.Trace(nil, formatLog(f, v...))
}

func formatLog(f interface{}, v ...interface{}) string {
	var msg string
	switch f.(type) {
	case string:
		msg = f.(string)
		if len(v) == 0 {
			return msg
		}
		if strings.Contains(msg, "%") && !strings.Contains(msg, "%%") {
			//format string
		} else {
			//do not contain format char
			msg += strings.Repeat(" %v", len(v))
		}
	default:
		msg = fmt.Sprint(f)
		if len(v) == 0 {
			return msg
		}
		msg += strings.Repeat(" %v", len(v))
	}
	return fmt.Sprintf(msg, v...)
}
