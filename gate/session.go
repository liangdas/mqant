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
	"github.com/golang/protobuf/proto"
	"github.com/liangdas/mqant/module"
)

type sessionagent struct {
	app       module.App
	session   *session
}

func NewSession(app module.App, data []byte) (Session,error) {
	agent:=&sessionagent{
		app:app,
	}
	se := &session{}
	err := proto.Unmarshal(data, se)
	if err != nil {
		return nil,err
	}    // 测试结果
	agent.session=se
	return agent,nil
}

func NewSessionByMap(app module.App, data map[string]interface{}) (Session,error) {
	agent:=&sessionagent{
		app:app,
		session:new(session),
	}
	err:=agent.updateMap(data)
	if err!=nil{
		return nil,err
	}
	return agent,nil
}

func (session *sessionagent) GetIP() string {
	return session.session.GetIP()
}

func (session *sessionagent) GetNetwork() string {
	return session.session.GetNetwork()
}

func (session *sessionagent) GetUserid() string {
	return session.session.GetUserid()
}

func (session *sessionagent) GetSessionid() string {
	return session.session.GetSessionid()
}

func (session *sessionagent) GetServerid() string {
	return session.session.GetServerid()
}

func (session *sessionagent) GetSettings() map[string]string {
	return session.session.GetSettings()
}


func (session *sessionagent)SetIP(ip string){
	session.session.IP=ip
}
func (session *sessionagent)SetNetwork(network string){
	session.session.Network=network
}
func (session *sessionagent)SetUserid(userid string){
	session.session.Userid=userid
}
func (session *sessionagent)SetSessionid(sessionid string){
	session.session.Sessionid=sessionid
}
func (session *sessionagent)SetServerid(serverid string){
	session.session.Serverid=serverid
}
func (session *sessionagent)SetSettings(settings map[string]string){
	session.session.Settings=settings
}

func (session *sessionagent) updateMap(s map[string]interface{})error {
	Userid := s["Userid"]
	if Userid != nil {
		session.session.Userid = Userid.(string)
	}
	IP := s["IP"]
	if IP != nil {
		session.session.IP = IP.(string)
	}
	Network := s["Network"]
	if Network != nil {
		session.session.Network = Network.(string)
	}
	Sessionid := s["Sessionid"]
	if Sessionid != nil {
		session.session.Sessionid = Sessionid.(string)
	}
	Serverid := s["Serverid"]
	if Serverid != nil {
		session.session.Serverid = Serverid.(string)
	}
	Settings := s["Settings"]
	if Settings != nil {
		session.session.Settings = Settings.(map[string]string)
	}
	return nil
}

func (session *sessionagent) update(s Session)error {
	Userid := s.GetUserid()
	session.session.Userid = Userid
	IP := s.GetIP()
	session.session.IP = IP
	Network := s.GetNetwork()
	session.session.Network = Network
	Sessionid := s.GetSessionid()
	session.session.Sessionid = Sessionid
	Serverid := s.GetServerid()
	session.session.Serverid = Serverid
	Settings := s.GetSettings()
	session.session.Settings = Settings
	return nil
}

func (session *sessionagent)Serializable()([]byte,error){
	data, err := proto.Marshal(session.session)
	if err != nil {
		return nil,err
	}    // 进行解码
	return data,nil
}


func (session *sessionagent) Update() (err string) {
	if session.app == nil {
		err = fmt.Sprintf("Module.App is nil")
		return
	}
	server, e := session.app.GetServersById(session.session.Serverid)
	if e != nil {
		err = fmt.Sprintf("Service not found id(%s)", session.session.Serverid)
		return
	}
	result, err := server.Call("Update", session.session.Sessionid)
	if err == "" {
		if result != nil {
			//绑定成功,重新更新当前Session
			session.update(result.(Session))
		}
	}
	return
}

func (session *sessionagent) Bind(Userid string) (err string) {
	if session.app == nil {
		err = fmt.Sprintf("Module.App is nil")
		return
	}
	server, e := session.app.GetServersById(session.session.Serverid)
	if e != nil {
		err = fmt.Sprintf("Service not found id(%s)", session.session.Serverid)
		return
	}
	result, err := server.Call("Bind", session.session.Sessionid, Userid)
	if err == "" {
		if result != nil {
			//绑定成功,重新更新当前Session
			session.update(result.(Session))
		}
	}
	return
}

func (session *sessionagent) UnBind() (err string) {
	if session.app == nil {
		err = fmt.Sprintf("Module.App is nil")
		return
	}
	server, e := session.app.GetServersById(session.session.Serverid)
	if e != nil {
		err = fmt.Sprintf("Service not found id(%s)", session.session.Serverid)
		return
	}
	result, err := server.Call("UnBind", session.session.Sessionid)
	if err == "" {
		if result != nil {
			//绑定成功,重新更新当前Session
			session.update(result.(Session))
		}
	}
	return
}

func (session *sessionagent) Push() (err string) {
	if session.app == nil {
		err = fmt.Sprintf("Module.App is nil")
		return
	}
	server, e := session.app.GetServersById(session.session.Serverid)
	if e != nil {
		err = fmt.Sprintf("Service not found id(%s)", session.session.Serverid)
		return
	}
	result, err := server.Call("Push", session.session.Sessionid, session.session.Settings)
	if err == "" {
		if result != nil {
			//绑定成功,重新更新当前Session
			session.update(result.(Session))
		}
	}
	return
}

func (session *sessionagent) Set(key string, value string) (err string) {
	if session.app == nil {
		err = fmt.Sprintf("Module.App is nil")
		return
	}
	if session.session.Settings == nil {
		session.session.Settings=map[string]string{}
	}
	session.session.Settings[key] = value
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

func (session *sessionagent) Get(key string) (result string) {
	if session.session.Settings == nil {
		return
	}
	result = session.session.Settings[key]
	return
}

func (session *sessionagent) Remove(key string) (err string) {
	if session.app == nil {
		err = fmt.Sprintf("Module.App is nil")
		return
	}
	if session.session.Settings == nil {
		session.session.Settings=map[string]string{}
	}
	delete(session.session.Settings, key)
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
func (session *sessionagent) Send(topic string, body []byte) (err string) {
	if session.app == nil {
		err = fmt.Sprintf("Module.App is nil")
		return
	}
	server, e := session.app.GetServersById(session.session.Serverid)
	if e != nil {
		err = fmt.Sprintf("Service not found id(%s)", session.session.Serverid)
		return
	}
	_, err = server.Call("Send", session.session.Sessionid, topic, body)
	return
}

func (session *sessionagent) SendNR(topic string, body []byte) (err string) {
	if session.app == nil {
		err = fmt.Sprintf("Module.App is nil")
		return
	}
	server, e := session.app.GetServersById(session.session.Serverid)
	if e != nil {
		err = fmt.Sprintf("Service not found id(%s)", session.session.Serverid)
		return
	}
	e = server.CallNR("Send", session.session.Sessionid, topic, body)
	if e != nil {
		err = e.Error()
	}
	return ""
}

func (session *sessionagent) Close() (err string) {
	if session.app == nil {
		err = fmt.Sprintf("Module.App is nil")
		return
	}
	server, e := session.app.GetServersById(session.session.Serverid)
	if e != nil {
		err = fmt.Sprintf("Service not found id(%s)", session.session.Serverid)
		return
	}
	_, err = server.Call("Close", session.session.Sessionid)
	return
}
