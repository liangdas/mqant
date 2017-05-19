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
	"github.com/liangdas/mqant/logger/ozzo-log"
)
var mqlog *log.Logger
var defaultLogger *log.Logger
func Init(debug bool, ProcessID string, Logdir string) {
	mqlog = NewMqantLog(debug, ProcessID, Logdir)
}

func Log()(*log.Logger){
	if mqlog==nil{
		if defaultLogger==nil{
			defaultLogger=NewDefaultLogger()
		}
		return defaultLogger
	}
	return mqlog
}

func Debug(format string, a ...interface{}) {
	//gLogger.doPrintf(debugLevel, printDebugLevel, format, a...)
	Log().Debug(format, a...)
}
func Info(format string, a ...interface{}) {
	//gLogger.doPrintf(releaseLevel, printReleaseLevel, format, a...)
	Log().Info(format, a...)
}

func Error(format string, a ...interface{}) {
	//gLogger.doPrintf(errorLevel, printErrorLevel, format, a...)
	Log().Error(format, a...)
}

func Warning(format string, a ...interface{}) {
	//gLogger.doPrintf(fatalLevel, printFatalLevel, format, a...)
	Log().Warning(format, a...)
}

func Close() {
	Log().Close()
}
