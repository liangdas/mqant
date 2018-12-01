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
	"github.com/liangdas/mqant/gate"
	"github.com/liangdas/mqant/log"
	"github.com/liangdas/mqant/module"
	"github.com/liangdas/mqant/utils"
	"strconv"
)

type sessionagent struct {
	app        module.App
	session    *SessionImp
	judgeGuest func(session gate.Session) bool
}

func NewSession(app module.App, data []byte) (gate.Session, error) {
	agent := &sessionagent{
		app: app,
	}
	se := &SessionImp{}
	err := proto.Unmarshal(data, se)
	if err != nil {
		return nil, err
	} // 测试结果
	agent.session = se
	return agent, nil
}

func NewSessionByMap(app module.App, data map[string]interface{}) (gate.Session, error) {
	agent := &sessionagent{
		app:     app,
		session: new(SessionImp),
	}
	err := agent.updateMap(data)
	if err != nil {
		return nil, err
	}
	return agent, nil
}

func (this *sessionagent) GetIP() string {
	return this.session.GetIP()
}

func (this *sessionagent) GetNetwork() string {
	return this.session.GetNetwork()
}

func (this *sessionagent) GetUserId() string {
	return this.session.GetUserId()
}

func (this *sessionagent) GetUserIdInt64() int64 {
	uid64, err := strconv.ParseInt(this.session.GetUserId(), 10, 64)
	if err != nil {
		return -1
	}
	return uid64
}

func (this *sessionagent) GetSessionId() string {
	return this.session.GetSessionId()
}

func (this *sessionagent) GetServerId() string {
	return this.session.GetServerId()
}

func (this *sessionagent) GetSettings() map[string]string {
	return this.session.GetSettings()
}

func (this *sessionagent) SetIP(ip string) {
	this.session.IP = ip
}
func (this *sessionagent) SetNetwork(network string) {
	this.session.Network = network
}
func (this *sessionagent) SetUserId(userid string) {
	this.session.UserId = userid
}
func (this *sessionagent) SetSessionId(sessionid string) {
	this.session.SessionId = sessionid
}
func (this *sessionagent) SetServerId(serverid string) {
	this.session.ServerId = serverid
}
func (this *sessionagent) SetSettings(settings map[string]string) {
	this.session.Settings = settings
}

func (this *sessionagent) updateMap(s map[string]interface{}) error {
	Userid := s["Userid"]
	if Userid != nil {
		this.session.UserId = Userid.(string)
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
		this.session.SessionId = Sessionid.(string)
	}
	Serverid := s["Serverid"]
	if Serverid != nil {
		this.session.ServerId = Serverid.(string)
	}
	Settings := s["Settings"]
	if Settings != nil {
		this.session.Settings = Settings.(map[string]string)
	}
	return nil
}

func (this *sessionagent) update(s gate.Session) error {
	Userid := s.GetUserId()
	this.session.UserId = Userid
	IP := s.GetIP()
	this.session.IP = IP
	Network := s.GetNetwork()
	this.session.Network = Network
	Sessionid := s.GetSessionId()
	this.session.SessionId = Sessionid
	Serverid := s.GetServerId()
	this.session.ServerId = Serverid
	Settings := s.GetSettings()
	this.session.Settings = Settings
	return nil
}

func (this *sessionagent) Serializable() ([]byte, error) {
	data, err := proto.Marshal(this.session)
	if err != nil {
		return nil, err
	} // 进行解码
	return data, nil
}

func (this *sessionagent) Update() (err string) {
	if this.app == nil {
		err = fmt.Sprintf("Module.App is nil")
		return
	}
	server, e := this.app.GetServerById(this.session.ServerId)
	if e != nil {
		err = fmt.Sprintf("Service not found id(%s)", this.session.ServerId)
		return
	}
	result, err := server.Call("Update", log.CreateTrace(this.TraceId(), this.SpanId()), this.session.SessionId)
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
	server, e := this.app.GetServerById(this.session.ServerId)
	if e != nil {
		err = fmt.Sprintf("Service not found id(%s)", this.session.ServerId)
		return
	}
	result, err := server.Call("Bind", log.CreateTrace(this.TraceId(), this.SpanId()), this.session.SessionId, Userid)
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
	server, e := this.app.GetServerById(this.session.ServerId)
	if e != nil {
		err = fmt.Sprintf("Service not found id(%s)", this.session.ServerId)
		return
	}
	result, err := server.Call("UnBind", log.CreateTrace(this.TraceId(), this.SpanId()), this.session.SessionId)
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
	server, e := this.app.GetServerById(this.session.ServerId)
	if e != nil {
		err = fmt.Sprintf("Service not found id(%s)", this.session.ServerId)
		return
	}
	result, err := server.Call("Push", log.CreateTrace(this.TraceId(), this.SpanId()), this.session.SessionId, this.session.Settings)
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
		this.session.Settings = map[string]string{}
	}
	this.session.Settings[key] = value
	return
}
func (this *sessionagent) SetPush(key string, value string) (err string) {
	if this.app == nil {
		err = fmt.Sprintf("Module.App is nil")
		return
	}
	if this.session.Settings == nil {
		this.session.Settings = map[string]string{}
	}
	this.session.Settings[key] = value
	return this.Push()
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
		this.session.Settings = map[string]string{}
	}
	delete(this.session.Settings, key)
	return
}
func (this *sessionagent) Send(topic string, body []byte) string {
	if this.app == nil {
		return fmt.Sprintf("Module.App is nil")
	}
	server, e := this.app.GetServerById(this.session.ServerId)
	if e != nil {
		return fmt.Sprintf("Service not found id(%s)", this.session.ServerId)
	}
	_, err := server.Call("Send", log.CreateTrace(this.TraceId(), this.SpanId()), this.session.SessionId, topic, body)
	return err
}

func (this *sessionagent) SendBatch(Sessionids string, topic string, body []byte) (int64, string) {
	if this.app == nil {
		return 0, fmt.Sprintf("Module.App is nil")
	}
	server, e := this.app.GetServerById(this.session.ServerId)
	if e != nil {
		return 0, fmt.Sprintf("Service not found id(%s)", this.session.ServerId)
	}
	count, err := server.Call("SendBatch", log.CreateTrace(this.TraceId(), this.SpanId()), Sessionids, topic, body)
	if err != "" {
		return 0, err
	}
	return count.(int64), err
}

func (this *sessionagent) IsConnect(userId string) (bool, string) {
	if this.app == nil {
		return false, fmt.Sprintf("Module.App is nil")
	}
	server, e := this.app.GetServerById(this.session.ServerId)
	if e != nil {
		return false, fmt.Sprintf("Service not found id(%s)", this.session.ServerId)
	}
	result, err := server.Call("IsConnect", log.CreateTrace(this.TraceId(), this.SpanId()), this.session.SessionId, userId)
	return result.(bool), err
}

func (this *sessionagent) SendNR(topic string, body []byte) string {
	if this.app == nil {
		return fmt.Sprintf("Module.App is nil")
	}
	server, e := this.app.GetServerById(this.session.ServerId)
	if e != nil {
		return fmt.Sprintf("Service not found id(%s)", this.session.ServerId)
	}
	e = server.CallNR("Send", log.CreateTrace(this.TraceId(), this.SpanId()), this.session.SessionId, topic, body)
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
	server, e := this.app.GetServerById(this.session.ServerId)
	if e != nil {
		err = fmt.Sprintf("Service not found id(%s)", this.session.ServerId)
		return
	}
	_, err = server.Call("Close", log.CreateTrace(this.TraceId(), this.SpanId()), this.session.SessionId)
	return
}

/**
每次rpc调用都拷贝一份新的Session进行传输
*/
func (this *sessionagent) Clone() gate.Session {
	agent := &sessionagent{
		app: this.app,
	}
	se := &SessionImp{
		IP:        this.session.IP,
		Network:   this.session.Network,
		UserId:    this.session.UserId,
		SessionId: this.session.SessionId,
		ServerId:  this.session.ServerId,
		TraceId:   this.session.TraceId,
		SpanId:    utils.GenerateID().String(),
		Settings:  this.session.Settings,
	}
	agent.session = se
	return agent
}

func (this *sessionagent) CreateTrace() {
	this.session.TraceId = utils.GenerateID().String()
	this.session.SpanId = utils.GenerateID().String()
}

func (this *sessionagent) TraceId() string {
	return this.session.TraceId
}

func (this *sessionagent) SpanId() string {
	return this.session.SpanId
}

func (this *sessionagent) ExtractSpan() log.TraceSpan {
	agent := &sessionagent{
		app: this.app,
	}
	se := &SessionImp{
		IP:        this.session.IP,
		Network:   this.session.Network,
		UserId:    this.session.UserId,
		SessionId: this.session.SessionId,
		ServerId:  this.session.ServerId,
		TraceId:   this.session.TraceId,
		SpanId:    utils.GenerateID().String(),
		Settings:  this.session.Settings,
	}
	agent.session = se
	return agent
}

//是否是访客(未登录) ,默认判断规则为 userId==""代表访客
func (this *sessionagent) IsGuest() bool {
	if this.judgeGuest != nil {
		return this.judgeGuest(this)
	}
	if this.GetUserId() == "" {
		return true
	} else {
		return false
	}
}

//设置自动的访客判断函数,记得一定要在gate模块设置这个值,以免部分模块因为未设置这个判断函数造成错误的判断
func (this *sessionagent) JudgeGuest(judgeGuest func(session gate.Session) bool) {
	this.judgeGuest = judgeGuest
}
