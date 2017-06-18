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
package basegate

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/liangdas/mqant/module"
	"github.com/liangdas/mqant/gate"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/liangdas/mqant/log"
)

type sessionagent struct {
	app       module.App
	session   *session
	span	  opentracing.Span
}

func NewSession(app module.App, data []byte) (gate.Session,error) {
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

func NewSessionByMap(app module.App, data map[string]interface{}) (gate.Session,error) {
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

func (this *sessionagent) GetIP() string {
	return this.session.GetIP()
}

func (this *sessionagent) GetNetwork() string {
	return this.session.GetNetwork()
}

func (this *sessionagent) GetUserid() string {
	return this.session.GetUserid()
}

func (this *sessionagent) GetSessionid() string {
	return this.session.GetSessionid()
}

func (this *sessionagent) GetServerid() string {
	return this.session.GetServerid()
}

func (this *sessionagent) GetSettings() map[string]string {
	return this.session.GetSettings()
}


func (this *sessionagent)SetIP(ip string){
	this.session.IP=ip
}
func (this *sessionagent)SetNetwork(network string){
	this.session.Network=network
}
func (this *sessionagent)SetUserid(userid string){
	this.session.Userid=userid
}
func (this *sessionagent)SetSessionid(sessionid string){
	this.session.Sessionid=sessionid
}
func (this *sessionagent)SetServerid(serverid string){
	this.session.Serverid=serverid
}
func (this *sessionagent)SetSettings(settings map[string]string){
	this.session.Settings=settings
}

func (this *sessionagent) updateMap(s map[string]interface{})error {
	Userid := s["Userid"]
	if Userid != nil {
		this.session.Userid = Userid.(string)
	}
	IP := s["IP"]
	if IP != nil {
		this.session.IP = IP.(string)
	}
	Network := s["Network"]
	if Network != nil {
		this.session.Network = Network.(string)
	}
	Sessionid := s["Sessionid"]
	if Sessionid != nil {
		this.session.Sessionid = Sessionid.(string)
	}
	Serverid := s["Serverid"]
	if Serverid != nil {
		this.session.Serverid = Serverid.(string)
	}
	Settings := s["Settings"]
	if Settings != nil {
		this.session.Settings = Settings.(map[string]string)
	}
	return nil
}

func (this *sessionagent) update(s gate.Session)error {
	Userid := s.GetUserid()
	this.session.Userid = Userid
	IP := s.GetIP()
	this.session.IP = IP
	Network := s.GetNetwork()
	this.session.Network = Network
	Sessionid := s.GetSessionid()
	this.session.Sessionid = Sessionid
	Serverid := s.GetServerid()
	this.session.Serverid = Serverid
	Settings := s.GetSettings()
	this.session.Settings = Settings
	return nil
}

func (this *sessionagent)Serializable()([]byte,error){
	data, err := proto.Marshal(this.session)
	if err != nil {
		return nil,err
	}    // 进行解码
	return data,nil
}


func (this *sessionagent) Update() (err string) {
	if this.app == nil {
		err = fmt.Sprintf("Module.App is nil")
		return
	}
	server, e := this.app.GetServersById(this.session.Serverid)
	if e != nil {
		err = fmt.Sprintf("Service not found id(%s)", this.session.Serverid)
		return
	}
	result, err := server.Call("Update", this.session.Sessionid)
	if err == "" {
		if result != nil {
			//绑定成功,重新更新当前Session
			this.update(result.(gate.Session))
		}
	}
	return
}

func (this *sessionagent) Bind(Userid string) (err string) {
	if this.app == nil {
		err = fmt.Sprintf("Module.App is nil")
		return
	}
	server, e := this.app.GetServersById(this.session.Serverid)
	if e != nil {
		err = fmt.Sprintf("Service not found id(%s)", this.session.Serverid)
		return
	}
	result, err := server.Call("Bind", this.session.Sessionid, Userid)
	if err == "" {
		if result != nil {
			//绑定成功,重新更新当前Session
			this.update(result.(gate.Session))
		}
	}
	return
}

func (this *sessionagent) UnBind() (err string) {
	if this.app == nil {
		err = fmt.Sprintf("Module.App is nil")
		return
	}
	server, e := this.app.GetServersById(this.session.Serverid)
	if e != nil {
		err = fmt.Sprintf("Service not found id(%s)", this.session.Serverid)
		return
	}
	result, err := server.Call("UnBind", this.session.Sessionid)
	if err == "" {
		if result != nil {
			//绑定成功,重新更新当前Session
			this.update(result.(gate.Session))
		}
	}
	return
}

func (this *sessionagent) Push() (err string) {
	if this.app == nil {
		err = fmt.Sprintf("Module.App is nil")
		return
	}
	server, e := this.app.GetServersById(this.session.Serverid)
	if e != nil {
		err = fmt.Sprintf("Service not found id(%s)", this.session.Serverid)
		return
	}
	result, err := server.Call("Push", this.session.Sessionid, this.session.Settings)
	if err == "" {
		if result != nil {
			//绑定成功,重新更新当前Session
			this.update(result.(gate.Session))
		}
	}
	return
}

func (this *sessionagent) Set(key string, value string) (err string) {
	if this.app == nil {
		err = fmt.Sprintf("Module.App is nil")
		return
	}
	if this.session.Settings == nil {
		this.session.Settings=map[string]string{}
	}
	this.session.Settings[key] = value
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

func (this *sessionagent) Get(key string) (result string) {
	if this.session.Settings == nil {
		return
	}
	result = this.session.Settings[key]
	return
}

func (this *sessionagent) Remove(key string) (err string) {
	if this.app == nil {
		err = fmt.Sprintf("Module.App is nil")
		return
	}
	if this.session.Settings == nil {
		this.session.Settings=map[string]string{}
	}
	delete(this.session.Settings, key)
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
func (this *sessionagent) Send(topic string, body []byte) (string) {
	if this.app == nil {
		return fmt.Sprintf("Module.App is nil")
	}
	server, e := this.app.GetServersById(this.session.Serverid)
	if e != nil {
		return fmt.Sprintf("Service not found id(%s)", this.session.Serverid)
	}
	_, err := server.Call("Send", this.session.Sessionid, topic, body)
	//span:=this.ExtractSpan(topic)
	//if span!=nil{
	//	span.LogEventWithPayload("SendToClient",map[string]string{
	//		"topic":topic,
	//		"err":err,
	//	})
	//	span.Finish()
	//}
	return	err
}

func (this *sessionagent) SendNR(topic string, body []byte) (string) {
	if this.app == nil {
		return fmt.Sprintf("Module.App is nil")
	}
	server, e := this.app.GetServersById(this.session.Serverid)
	if e != nil {
		return fmt.Sprintf("Service not found id(%s)", this.session.Serverid)
	}
	e = server.CallNR("Send", this.session.Sessionid, topic, body)
	if e != nil {
		return e.Error()
	}
	//span:=this.ExtractSpan(topic)
	//if span!=nil{
	//	span.LogEventWithPayload("SendToClient",map[string]string{
	//		"topic":topic,
	//	})
	//	span.Finish()
	//}
	return ""
}

func (this *sessionagent) Close() (err string) {
	if this.app == nil {
		err = fmt.Sprintf("Module.App is nil")
		return
	}
	server, e := this.app.GetServersById(this.session.Serverid)
	if e != nil {
		err = fmt.Sprintf("Service not found id(%s)", this.session.Serverid)
		return
	}
	_, err = server.Call("Close", this.session.Sessionid)
	return
}

/**
每次rpc调用都拷贝一份新的Session进行传输
 */
func (this *sessionagent) Clone()gate.Session{
	agent:=&sessionagent{
		app:this.app,
		span:this.Span(),
	}
	se := &session{
		IP        :this.session.IP,
		Network   :this.session.Network,
		Userid    :this.session.Userid,
		Sessionid :this.session.Sessionid,
		Serverid  :this.session.Serverid,
		Settings  :this.session.Settings,
	}
	//这个要换成本次RPC调用的新Span
	se.Carrier=this.inject()

	agent.session=se
	return agent
}

func (this *sessionagent)inject()map[string]string{
	if this.app.GetTracer()==nil{
		return nil
	}
	if this.Span()==nil{
		return nil
	}
	carrier := &opentracing.TextMapCarrier{}
	err := this.app.GetTracer().Inject(
		this.Span().Context(),
		opentracing.TextMap,
		carrier)
	if err!=nil{
		log.Warning("session.session.Carrier Inject Fail",err.Error())
		return nil
	}else{
		m:=map[string]string{}
		carrier.ForeachKey(func(key, val string) error{
			m[key]=val
			return nil
		})
		return m
	}
}
func (this *sessionagent)extract(gCarrier map[string]string)(opentracing.SpanContext, error){
	carrier := &opentracing.TextMapCarrier{}
	for v,k:=range gCarrier{
		carrier.Set(v,k)
	}
	return this.app.GetTracer().Extract(opentracing.TextMap, carrier)
}
func (this *sessionagent)LoadSpan(operationName string)opentracing.Span{
	if this.app.GetTracer()==nil{
		return nil
	}
	if this.span==nil{
		if this.session.Carrier!=nil{
			//从已有记录恢复
			clientContext, err := this.extract(this.session.Carrier)
			if err == nil {
				this.span = this.app.GetTracer().StartSpan(
					operationName, opentracing.ChildOf(clientContext))
			} else {
				log.Warning("session.session.Carrier Extract Fail",err.Error())
			}
		}
	}
	return this.span
}
func (this *sessionagent)CreateRootSpan(operationName string)opentracing.Span{
	if this.app.GetTracer()==nil{
		return nil
	}
	this.span = this.app.GetTracer().StartSpan(operationName)
	this.session.Carrier=this.inject()
	return this.span
}
func (this *sessionagent)Span()opentracing.Span{
	return this.span
}

func (this *sessionagent)TracCarrier()map[string]string{
	return this.session.Carrier
}
/**
从Session的 Span继承一个新的Span
 */
func (this *sessionagent)ExtractSpan(operationName string)opentracing.Span{
	if this.app.GetTracer()==nil{
		return nil
	}
	if this.Span()!=nil{
		span := this.app.GetTracer().StartSpan(operationName,opentracing.ChildOf(this.Span().Context()))
		return span
	}
	return nil
}


