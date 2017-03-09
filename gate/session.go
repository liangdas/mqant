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
package gate

import (
	"fmt"
	"github.com/liangdas/mqant/module"
)

type Session struct {
	app       module.App
	IP        string //客户端IP
	Network   string //网络类型 TCP UDP websocket
	Userid    string
	Sessionid string
	Serverid  string
	Settings  map[string]interface{}
}

func NewSession(app module.App, s map[string]interface{}) *Session {
	se := &Session{
		app: app,
	}
	se.update(s)
	return se
}

func (session *Session) update(s map[string]interface{}) {
	Userid := s["Userid"]
	if Userid != nil {
		session.Userid = Userid.(string)
	}
	IP := s["IP"]
	if IP != nil {
		session.IP = IP.(string)
	}
	Network := s["Network"]
	if Network != nil {
		session.Network = Network.(string)
	}
	Sessionid := s["Sessionid"]
	if Sessionid != nil {
		session.Sessionid = Sessionid.(string)
	}
	Serverid := s["Serverid"]
	if Serverid != nil {
		session.Serverid = Serverid.(string)
	}
	Settings := s["Settings"]
	if Settings != nil {
		session.Settings = Settings.(map[string]interface{})
	}
}

func (session *Session) ExportMap() map[string]interface{} {
	s := map[string]interface{}{}
	if session.Userid != "" {
		s["Userid"] = session.Userid
	}
	if session.IP != "" {
		s["IP"] = session.IP
	}
	if session.Network != "" {
		s["Network"] = session.Network
	}
	if session.Sessionid != "" {
		s["Sessionid"] = session.Sessionid
	}
	if session.Serverid != "" {
		s["Serverid"] = session.Serverid
	}
	if session.Settings != nil {
		s["Settings"] = session.Settings
	}
	return s
}

func (session *Session) Update() (err string) {
	if session.app == nil {
		err = fmt.Sprintf("Module.App is nil")
		return
	}
	server, e := session.app.GetServersById(session.Serverid)
	if e != nil {
		err = fmt.Sprintf("Service not found id(%s)", session.Serverid)
		return
	}
	result, err := server.Call("Update", session.Sessionid)
	if err == "" {
		if result != nil {
			//绑定成功,重新更新当前Session
			session.update(result.(map[string]interface{}))
		}
	}
	return
}

func (session *Session) Bind(Userid string) (err string) {
	if session.app == nil {
		err = fmt.Sprintf("Module.App is nil")
		return
	}
	server, e := session.app.GetServersById(session.Serverid)
	if e != nil {
		err = fmt.Sprintf("Service not found id(%s)", session.Serverid)
		return
	}
	result, err := server.Call("Bind", session.Sessionid, Userid)
	if err == "" {
		if result != nil {
			//绑定成功,重新更新当前Session
			session.update(result.(map[string]interface{}))
		}
	}
	return
}

func (session *Session) UnBind() (err string) {
	if session.app == nil {
		err = fmt.Sprintf("Module.App is nil")
		return
	}
	server, e := session.app.GetServersById(session.Serverid)
	if e != nil {
		err = fmt.Sprintf("Service not found id(%s)", session.Serverid)
		return
	}
	result, err := server.Call("UnBind", session.Sessionid)
	if err == "" {
		if result != nil {
			//绑定成功,重新更新当前Session
			session.update(result.(map[string]interface{}))
		}
	}
	return
}

func (session *Session) Push() (err string) {
	if session.app == nil {
		err = fmt.Sprintf("Module.App is nil")
		return
	}
	server, e := session.app.GetServersById(session.Serverid)
	if e != nil {
		err = fmt.Sprintf("Service not found id(%s)", session.Serverid)
		return
	}
	result, err := server.Call("Push", session.Sessionid, session.Settings)
	if err == "" {
		if result != nil {
			//绑定成功,重新更新当前Session
			session.update(result.(map[string]interface{}))
		}
	}
	return
}

func (session *Session) Set(key string, value interface{}) (err string) {
	if session.app == nil {
		err = fmt.Sprintf("Module.App is nil")
		return
	}
	if session.Settings == nil {
		err = fmt.Sprintf("Session.Settings is nil")
		return
	}
	session.Settings[key] = value
	//server,e:=session.app.GetServersById(session.Serverid)
	//if e!=nil{
	//	err=fmt.Sprintf("Service not found id(%s)",session.Serverid)
	//	return
	//}
	//result,err:=server.Call("Set",session.Sessionid,key,value)
	//if err==""{
	//	if result!=nil{
	//		//绑定成功,重新更新当前Session
	//		session.update(result.(map[string]interface {}))
	//	}
	//}
	return
}

func (session *Session) Get(key string) (result interface{}) {
	if session.Settings == nil {
		return
	}
	result = session.Settings[key]
	return
}

func (session *Session) Remove(key string) (err string) {
	if session.app == nil {
		err = fmt.Sprintf("Module.App is nil")
		return
	}
	if session.Settings == nil {
		err = fmt.Sprintf("Session.Settings is nil")
		return
	}
	delete(session.Settings, key)
	//server,e:=session.app.GetServersById(session.Serverid)
	//if e!=nil{
	//	err=fmt.Sprintf("Service not found id(%s)",session.Serverid)
	//	return
	//}
	//result,err:=server.Call("Remove",session.Sessionid,key)
	//if err==""{
	//	if result!=nil{
	//		//绑定成功,重新更新当前Session
	//		session.update(result.(map[string]interface {}))
	//	}
	//}
	return
}
func (session *Session) Send(topic string, body []byte) (err string) {
	if session.app == nil {
		err = fmt.Sprintf("Module.App is nil")
		return
	}
	server, e := session.app.GetServersById(session.Serverid)
	if e != nil {
		err = fmt.Sprintf("Service not found id(%s)", session.Serverid)
		return
	}
	_, err = server.Call("Send", session.Sessionid, topic, body)
	return
}

func (session *Session) SendNR(topic string, body []byte) (err string) {
	if session.app == nil {
		err = fmt.Sprintf("Module.App is nil")
		return
	}
	server, e := session.app.GetServersById(session.Serverid)
	if e != nil {
		err = fmt.Sprintf("Service not found id(%s)", session.Serverid)
		return
	}
	e = server.CallNR("Send", session.Sessionid, topic, body)
	if e != nil {
		err = e.Error()
	}
	return ""
}

func (session *Session) Close() (err string) {
	if session.app == nil {
		err = fmt.Sprintf("Module.App is nil")
		return
	}
	server, e := session.app.GetServersById(session.Serverid)
	if e != nil {
		err = fmt.Sprintf("Service not found id(%s)", session.Serverid)
		return
	}
	_, err = server.Call("Close", session.Sessionid)
	return
}
