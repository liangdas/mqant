// Copyright 2014 mqant Author. All Rights Reserved.
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
package log

import (
	beegolog "github.com/liangdas/mqant/log/beego"
	"github.com/liangdas/mqant/utils"
)

var beego *beegolog.BeeLogger
var bi *beegolog.BeeLogger

func InitLog(debug bool, ProcessID string, Logdir string, settings map[string]interface{}) {
	beego = NewBeegoLogger(debug, ProcessID, Logdir, settings)
}
func InitBI(debug bool, ProcessID string, Logdir string, settings map[string]interface{}) {
	bi = NewBeegoLogger(debug, ProcessID, Logdir, settings)
}
func LogBeego() *beegolog.BeeLogger {
	if beego == nil {
		beego = beegolog.NewLogger()
	}
	return beego
}

func BiBeego() *beegolog.BeeLogger {
	return bi
}

func CreateRootTrace() TraceSpan {
	return &TraceSpanImp{
		Trace: utils.GenerateID().String(),
		Span:  utils.GenerateID().String(),
	}
}

func CreateTrace(trace, span string) TraceSpan {
	return &TraceSpanImp{
		Trace: trace,
		Span:  span,
	}
}

func BiReport(msg string) {
	//gLogger.doPrintf(debugLevel, printDebugLevel, format, a...)
	l := BiBeego()
	if l != nil {
		l.BiReport(msg)
	}
}

func Debug(format string, a ...interface{}) {
	//gLogger.doPrintf(debugLevel, printDebugLevel, format, a...)
	LogBeego().Debug(nil, format, a...)
}
func Info(format string, a ...interface{}) {
	//gLogger.doPrintf(releaseLevel, printReleaseLevel, format, a...)
	LogBeego().Info(nil, format, a...)
}

func Error(format string, a ...interface{}) {
	//gLogger.doPrintf(errorLevel, printErrorLevel, format, a...)
	LogBeego().Error(nil, format, a...)
}

func Warning(format string, a ...interface{}) {
	//gLogger.doPrintf(fatalLevel, printFatalLevel, format, a...)
	LogBeego().Warning(nil, format, a...)
}

func TDebug(span TraceSpan, format string, a ...interface{}) {
	if span != nil {
		LogBeego().Debug(
			&beegolog.BeegoTraceSpan{
				Trace: span.TraceId(),
				Span:  span.SpanId(),
			}, format, a...)
	} else {
		LogBeego().Debug(nil, format, a...)
	}
}
func TInfo(span TraceSpan, format string, a ...interface{}) {
	if span != nil {
		LogBeego().Info(
			&beegolog.BeegoTraceSpan{
				Trace: span.TraceId(),
				Span:  span.SpanId(),
			}, format, a...)
	} else {
		LogBeego().Info(nil, format, a...)
	}
}

func TError(span TraceSpan, format string, a ...interface{}) {
	if span != nil {
		LogBeego().Error(
			&beegolog.BeegoTraceSpan{
				Trace: span.TraceId(),
				Span:  span.SpanId(),
			}, format, a...)
	} else {
		LogBeego().Error(nil, format, a...)
	}
}

func TWarning(span TraceSpan, format string, a ...interface{}) {
	if span != nil {
		LogBeego().Warning(
			&beegolog.BeegoTraceSpan{
				Trace: span.TraceId(),
				Span:  span.SpanId(),
			}, format, a...)
	} else {
		LogBeego().Warning(nil, format, a...)
	}
}

func Close() {
	LogBeego().Close()
}
