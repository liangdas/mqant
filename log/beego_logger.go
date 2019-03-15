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
	"encoding/json"
	"fmt"
	"github.com/liangdas/mqant/log/beego"
)

func NewBeegoLogger(debug bool, ProcessID string, Logdir string, settings map[string]interface{}) *logs.BeeLogger {
	log := logs.NewLogger()
	log.ProcessID = ProcessID
	log.EnableFuncCallDepth(true)
	log.Async(1024)	//同步打印,可能影响性能
	log.SetLogFuncCallDepth(4)
	if debug {
		//控制台
		log.SetLogger(logs.AdapterConsole)
	}
	if contenttype, ok := settings["contenttype"]; ok {
		log.SetContentType(contenttype.(string))
	}
	if f, ok := settings["file"]; ok {
		ff := f.(map[string]interface{})
		Prefix := ""
		if prefix, ok := ff["prefix"]; ok {
			Prefix = prefix.(string)
		}
		Suffix := ".log"
		if suffix, ok := ff["suffix"]; ok {
			Suffix = suffix.(string)
		}
		ff["filename"] = fmt.Sprintf("%s/%v%s%s", Logdir, Prefix, ProcessID, Suffix)
		config, err := json.Marshal(ff)
		if err != nil {
			logs.Error(err)
		}
		log.SetLogger(logs.AdapterFile, string(config))
	}
	if f, ok := settings["multifile"]; ok {
		multifile := f.(map[string]interface{})
		Prefix := ""
		if prefix, ok := multifile["prefix"]; ok {
			Prefix = prefix.(string)
		}
		Suffix := ".log"
		if suffix, ok := multifile["suffix"]; ok {
			Suffix = suffix.(string)
		}
		multifile["filename"] = fmt.Sprintf("%s/%v%s%s", Logdir, Prefix, ProcessID, Suffix)
		config, err := json.Marshal(multifile)
		if err != nil {
			logs.Error(err)
		}
		log.SetLogger(logs.AdapterMultiFile, string(config))
	}
	if dingtalk, ok := settings["dingtalk"]; ok {
		config, err := json.Marshal(dingtalk)
		if err != nil {
			logs.Error(err)
		}
		log.SetLogger(logs.AdapterDingtalk, string(config))
	}
	if slack, ok := settings["slack"]; ok {
		config, err := json.Marshal(slack)
		if err != nil {
			logs.Error(err)
		}
		log.SetLogger(logs.AdapterSlack, string(config))
	}
	if jianliao, ok := settings["jianliao"]; ok {
		config, err := json.Marshal(jianliao)
		if err != nil {
			logs.Error(err)
		}
		log.SetLogger(logs.AdapterJianLiao, string(config))
	}
	if conn, ok := settings["conn"]; ok {
		config, err := json.Marshal(conn)
		if err != nil {
			logs.Error(err)
		}
		log.SetLogger(logs.AdapterConn, string(config))
	}
	if smtp, ok := settings["smtp"]; ok {
		config, err := json.Marshal(smtp)
		if err != nil {
			logs.Error(err)
		}
		log.SetLogger(logs.AdapterMail, string(config))
	}
	if es, ok := settings["es"]; ok {
		config, err := json.Marshal(es)
		if err != nil {
			logs.Error(err)
		}
		log.SetLogger(logs.AdapterEs, string(config))
	}
	return log
}
