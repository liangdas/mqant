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

// Package basegate gate.Session
package basegate

import (
	"fmt"
	"strconv"
	"sync"

	"google.golang.org/protobuf/proto"
	"github.com/liangdas/mqant/gate"
	"github.com/liangdas/mqant/log"
	"github.com/liangdas/mqant/module"
	"github.com/liangdas/mqant/rpc"
	"github.com/liangdas/mqant/utils"
)

type sessionagent struct {
	app        module.App
	session    *SessionImp
	lock       *sync.RWMutex
	userdata   interface{}
	judgeGuest func(session gate.Session) bool
}

// NewSession NewSession
func NewSession(app module.App, data []byte) (gate.Session, error) {
	agent := &sessionagent{
		app:  app,
		lock: new(sync.RWMutex),
	}
	se := &SessionImp{}
	err := proto.Unmarshal(data, se)
	if err != nil {
		return nil, err
	} // 测试结果
	agent.session = se
	if app != nil {
		agent.judgeGuest = gate.JudgeGuest
	}
	if agent.session.GetSettings() == nil {
		agent.session.Settings = make(map[string]string)
	}
	return agent, nil
}

// NewSessionByMap NewSessionByMap
func NewSessionByMap(app module.App, data map[string]interface{}) (gate.Session, error) {
	agent := &sessionagent{
		app:     app,
		session: new(SessionImp),
		lock:    new(sync.RWMutex),
	}
	err := agent.updateMap(data)
	if err != nil {
		return nil, err
	}
	if agent.session.GetSettings() == nil {
		agent.session.Settings = make(map[string]string)
	}
	if app != nil {
		agent.judgeGuest = gate.JudgeGuest
	}
	return agent, nil
}

func (sesid *sessionagent) GetIP() string {
	return sesid.session.GetIP()
}

func (sesid *sessionagent) GetTopic() string {
	return sesid.session.GetTopic()
}

func (sesid *sessionagent) GetNetwork() string {
	return sesid.session.GetNetwork()
}

func (sesid *sessionagent) GetUserId() string {
	return sesid.GetUserID()
}

func (sesid *sessionagent) GetUserID() string {
	return sesid.session.GetUserId()
}

func (sesid *sessionagent) GetUserIDInt64() int64 {
	uid64, err := strconv.ParseInt(sesid.session.GetUserId(), 10, 64)
	if err != nil {
		return -1
	}
	return uid64
}

func (sesid *sessionagent) GetUserIdInt64() int64 {
	return sesid.GetUserIDInt64()
}

func (sesid *sessionagent) GetSessionID() string {
	return sesid.session.GetSessionId()
}

func (sesid *sessionagent) GetSessionId() string {
	return sesid.GetSessionID()
}

func (sesid *sessionagent) GetServerID() string {
	return sesid.session.GetServerId()
}

func (sesid *sessionagent) GetServerId() string {
	return sesid.GetServerID()
}

func (sesid *sessionagent) SettingsRange(f func(k, v string) bool) {
	sesid.lock.Lock()
	defer sesid.lock.Unlock()
	if sesid.session.GetSettings() == nil {
		return
	}
	for k, v := range sesid.session.GetSettings() {
		c := f(k, v)
		if c == false {
			return
		}
	}
}

//ImportSettings 合并两个map 并且以 agent.(Agent).GetSession().Settings 已有的优先
func (sesid *sessionagent) ImportSettings(settings map[string]string) error {
	sesid.lock.Lock()
	if sesid.session.GetSettings() == nil {
		sesid.session.Settings = settings
	} else {
		for k, v := range settings {
			if _, ok := sesid.session.GetSettings()[k]; ok {
				//不用替换
			} else {
				sesid.session.GetSettings()[k] = v
			}
		}
	}
	sesid.lock.Unlock()
	return nil
}

func (sesid *sessionagent) LocalUserData() interface{} {
	return sesid.userdata
}

func (sesid *sessionagent) SetIP(ip string) {
	sesid.session.IP = ip
}
func (sesid *sessionagent) SetTopic(topic string) {
	sesid.session.Topic = topic
}
func (sesid *sessionagent) SetNetwork(network string) {
	sesid.session.Network = network
}
func (sesid *sessionagent) SetUserID(userid string) {
	sesid.lock.Lock()
	sesid.session.UserId = userid
	sesid.lock.Unlock()
}
func (sesid *sessionagent) SetUserId(userid string) {
	sesid.SetUserID(userid)
}
func (sesid *sessionagent) SetSessionID(sessionid string) {
	sesid.session.SessionId = sessionid
}
func (sesid *sessionagent) SetSessionId(sessionid string) {
	sesid.SetSessionID(sessionid)
}
func (sesid *sessionagent) SetServerID(serverid string) {
	sesid.session.ServerId = serverid
}
func (sesid *sessionagent) SetServerId(serverid string) {
	sesid.SetServerID(serverid)
}
func (sesid *sessionagent) SetSettings(settings map[string]string) {
	sesid.lock.Lock()
	sesid.session.Settings = settings
	sesid.lock.Unlock()
}
func (sesid *sessionagent) CloneSettings() map[string]string {
	sesid.lock.Lock()
	defer sesid.lock.Unlock()
	tmp := map[string]string{}
	for k, v := range sesid.session.Settings {
		tmp[k] = v
	}
	return tmp
}
func (sesid *sessionagent) SetLocalKV(key, value string) error {
	sesid.lock.Lock()
	sesid.session.GetSettings()[key] = value
	sesid.lock.Unlock()
	return nil
}
func (sesid *sessionagent) RemoveLocalKV(key string) error {
	sesid.lock.Lock()
	delete(sesid.session.GetSettings(), key)
	sesid.lock.Unlock()
	return nil
}
func (sesid *sessionagent) SetLocalUserData(data interface{}) error {
	sesid.userdata = data
	return nil
}

func (sesid *sessionagent) updateMap(s map[string]interface{}) error {
	Userid := s["Userid"]
	if Userid != nil {
		sesid.session.UserId = Userid.(string)
	}
	IP := s["IP"]
	if IP != nil {
		sesid.session.IP = IP.(string)
	}
	if topic, ok := s["Topic"]; ok {
		sesid.session.Topic = topic.(string)
	}
	Network := s["Network"]
	if Network != nil {
		sesid.session.Network = Network.(string)
	}
	Sessionid := s["Sessionid"]
	if Sessionid != nil {
		sesid.session.SessionId = Sessionid.(string)
	}
	Serverid := s["Serverid"]
	if Serverid != nil {
		sesid.session.ServerId = Serverid.(string)
	}
	Settings := s["Settings"]
	if Settings != nil {
		sesid.lock.Lock()
		sesid.session.Settings = Settings.(map[string]string)
		sesid.lock.Unlock()
	}
	return nil
}

func (sesid *sessionagent) update(s gate.Session) error {
	Userid := s.GetUserID()
	sesid.session.UserId = Userid
	IP := s.GetIP()
	sesid.session.IP = IP
	sesid.session.Topic = s.GetTopic()
	Network := s.GetNetwork()
	sesid.session.Network = Network
	Sessionid := s.GetSessionID()
	sesid.session.SessionId = Sessionid
	Serverid := s.GetServerID()
	sesid.session.ServerId = Serverid
	Settings := map[string]string{}
	s.SettingsRange(func(k, v string) bool {
		Settings[k] = v
		return true
	})
	sesid.lock.Lock()
	sesid.session.Settings = Settings
	sesid.lock.Unlock()
	return nil
}

func (sesid *sessionagent) Serializable() ([]byte, error) {
	sesid.lock.RLock()
	data, err := proto.Marshal(sesid.session)
	sesid.lock.RUnlock()
	if err != nil {
		return nil, err
	} // 进行解码
	return data, nil
}

func (sesid *sessionagent) Marshal() ([]byte, error) {
	sesid.lock.RLock()
	data, err := proto.Marshal(sesid.session)
	sesid.lock.RUnlock()
	if err != nil {
		return nil, err
	} // 进行解码
	return data, nil
}
func (sesid *sessionagent) Unmarshal(data []byte) error {
	se := &SessionImp{}
	err := proto.Unmarshal(data, se)
	if err != nil {
		return err
	} // 测试结果
	sesid.session = se
	if sesid.session.GetSettings() == nil {
		sesid.lock.Lock()
		sesid.session.Settings = make(map[string]string)
		sesid.lock.Unlock()
	}
	return nil
}
func (sesid *sessionagent) String() string {
	return "gate.Session"
}

func (sesid *sessionagent) Update() (err string) {
	if sesid.app == nil {
		err = fmt.Sprintf("Module.App is nil")
		return
	}
	server, e := sesid.app.GetServerByID(sesid.session.ServerId)
	if e != nil {
		err = fmt.Sprintf("Service not found id(%s)", sesid.session.ServerId)
		return
	}
	result, err := server.Call(nil, "Update", log.CreateTrace(sesid.TraceId(), sesid.SpanId()), sesid.session.SessionId)
	if err == "" {
		if result != nil {
			//绑定成功,重新更新当前Session
			sesid.update(result.(gate.Session))
		}
	}
	return
}

func (sesid *sessionagent) Bind(Userid string) (err string) {
	if sesid.app == nil {
		err = fmt.Sprintf("Module.App is nil")
		return
	}
	server, e := sesid.app.GetServerByID(sesid.session.ServerId)
	if e != nil {
		err = fmt.Sprintf("Service not found id(%s)", sesid.session.ServerId)
		return
	}
	result, err := server.Call(nil, "Bind", log.CreateTrace(sesid.TraceId(), sesid.SpanId()), sesid.session.SessionId, Userid)
	if err == "" {
		if result != nil {
			//绑定成功,重新更新当前Session
			sesid.update(result.(gate.Session))
		}
	}
	return
}

func (sesid *sessionagent) UnBind() (err string) {
	if sesid.app == nil {
		err = fmt.Sprintf("Module.App is nil")
		return
	}
	server, e := sesid.app.GetServerByID(sesid.session.ServerId)
	if e != nil {
		err = fmt.Sprintf("Service not found id(%s)", sesid.session.ServerId)
		return
	}
	result, err := server.Call(nil, "UnBind", log.CreateTrace(sesid.TraceId(), sesid.SpanId()), sesid.session.SessionId)
	if err == "" {
		if result != nil {
			//绑定成功,重新更新当前Session
			sesid.update(result.(gate.Session))
		}
	}
	return
}

func (sesid *sessionagent) Push() (err string) {
	if sesid.app == nil {
		err = fmt.Sprintf("Module.App is nil")
		return
	}
	server, e := sesid.app.GetServerByID(sesid.session.ServerId)
	if e != nil {
		err = fmt.Sprintf("Service not found id(%s)", sesid.session.ServerId)
		return
	}
	sesid.lock.Lock()
	tmp := map[string]string{}
	for k, v := range sesid.session.Settings {
		tmp[k] = v
	}
	sesid.lock.Unlock()
	result, err := server.Call(nil, "Push", log.CreateTrace(sesid.TraceId(), sesid.SpanId()), sesid.session.SessionId, tmp)
	if err == "" {
		if result != nil {
			//绑定成功,重新更新当前Session
			sesid.update(result.(gate.Session))
		}
	}
	return
}

func (sesid *sessionagent) Set(key string, value string) (err string) {
	if sesid.app == nil {
		err = fmt.Sprintf("Module.App is nil")
		return
	}
	result, err := sesid.app.Call(nil,
		sesid.session.ServerId,
		"Set",
		mqrpc.Param(
			log.CreateTrace(sesid.TraceId(), sesid.SpanId()),
			sesid.session.SessionId,
			key,
			value,
		),
	)
	if err == "" {
		if result != nil {
			//绑定成功,重新更新当前Session
			sesid.update(result.(gate.Session))
		}
	}
	return
}
func (sesid *sessionagent) SetPush(key string, value string) (err string) {
	if sesid.app == nil {
		err = fmt.Sprintf("Module.App is nil")
		return
	}
	if sesid.session.Settings == nil {
		sesid.session.Settings = map[string]string{}
	}
	sesid.lock.Lock()
	sesid.session.Settings[key] = value
	sesid.lock.Unlock()
	return sesid.Push()
}
func (sesid *sessionagent) SetBatch(settings map[string]string) (err string) {
	if sesid.app == nil {
		err = fmt.Sprintf("Module.App is nil")
		return
	}
	server, e := sesid.app.GetServerByID(sesid.session.ServerId)
	if e != nil {
		err = fmt.Sprintf("Service not found id(%s)", sesid.session.ServerId)
		return
	}
	result, err := server.Call(nil, "Push", log.CreateTrace(sesid.TraceId(), sesid.SpanId()), sesid.session.SessionId, settings)
	if err == "" {
		if result != nil {
			//绑定成功,重新更新当前Session
			sesid.update(result.(gate.Session))
		}
	}
	return
}
func (sesid *sessionagent) Get(key string) (result string) {
	sesid.lock.RLock()
	if sesid.session.Settings == nil {
		sesid.lock.RUnlock()
		return
	}
	result = sesid.session.Settings[key]
	sesid.lock.RUnlock()
	return
}

func (sesid *sessionagent) Load(key string) (result string, ok bool) {
	sesid.lock.RLock()
	defer sesid.lock.RUnlock()
	if sesid.session.Settings == nil {
		return "", false
	}
	if result, ok = sesid.session.Settings[key]; ok {
		return result, ok
	} else {
		return "", false
	}
}

func (sesid *sessionagent) Remove(key string) (errStr string) {
	if sesid.app == nil {
		errStr = fmt.Sprintf("Module.App is nil")
		return
	}
	result, err := sesid.app.Call(nil,
		sesid.session.ServerId,
		"Remove",
		mqrpc.Param(
			log.CreateTrace(sesid.TraceId(), sesid.SpanId()),
			sesid.session.SessionId,
			key,
		),
	)
	if err == "" {
		if result != nil {
			//绑定成功,重新更新当前Session
			sesid.update(result.(gate.Session))
		}
	}
	return
}
func (sesid *sessionagent) Send(topic string, body []byte) string {
	if sesid.app == nil {
		return fmt.Sprintf("Module.App is nil")
	}
	server, e := sesid.app.GetServerByID(sesid.session.ServerId)
	if e != nil {
		return fmt.Sprintf("Service not found id(%s)", sesid.session.ServerId)
	}
	_, err := server.Call(nil, "Send", log.CreateTrace(sesid.TraceId(), sesid.SpanId()), sesid.session.SessionId, topic, body)
	return err
}

func (sesid *sessionagent) SendBatch(Sessionids string, topic string, body []byte) (int64, string) {
	if sesid.app == nil {
		return 0, fmt.Sprintf("Module.App is nil")
	}
	server, e := sesid.app.GetServerByID(sesid.session.ServerId)
	if e != nil {
		return 0, fmt.Sprintf("Service not found id(%s)", sesid.session.ServerId)
	}
	count, err := server.Call(nil, "SendBatch", log.CreateTrace(sesid.TraceId(), sesid.SpanId()), Sessionids, topic, body)
	if err != "" {
		return 0, err
	}
	return count.(int64), err
}

func (sesid *sessionagent) IsConnect(userId string) (bool, string) {
	if sesid.app == nil {
		return false, fmt.Sprintf("Module.App is nil")
	}
	server, e := sesid.app.GetServerByID(sesid.session.ServerId)
	if e != nil {
		return false, fmt.Sprintf("Service not found id(%s)", sesid.session.ServerId)
	}
	result, err := server.Call(nil, "IsConnect", log.CreateTrace(sesid.TraceId(), sesid.SpanId()), sesid.session.SessionId, userId)
	if err != "" {
		return false, err
	}
	return result.(bool), err
}

func (sesid *sessionagent) SendNR(topic string, body []byte) string {
	if sesid.app == nil {
		return fmt.Sprintf("Module.App is nil")
	}
	server, e := sesid.app.GetServerByID(sesid.session.ServerId)
	if e != nil {
		return fmt.Sprintf("Service not found id(%s)", sesid.session.ServerId)
	}
	e = server.CallNR("Send", log.CreateTrace(sesid.TraceId(), sesid.SpanId()), sesid.session.SessionId, topic, body)
	if e != nil {
		return e.Error()
	}
	//span:=sesid.ExtractSpan(topic)
	//if span!=nil{
	//	span.LogEventWithPayload("SendToClient",map[string]string{
	//		"topic":topic,
	//	})
	//	span.Finish()
	//}
	return ""
}

func (sesid *sessionagent) Close() (err string) {
	if sesid.app == nil {
		err = fmt.Sprintf("Module.App is nil")
		return
	}
	server, e := sesid.app.GetServerByID(sesid.session.ServerId)
	if e != nil {
		err = fmt.Sprintf("Service not found id(%s)", sesid.session.ServerId)
		return
	}
	_, err = server.Call(nil, "Close", log.CreateTrace(sesid.TraceId(), sesid.SpanId()), sesid.session.SessionId)
	return
}

/**
每次rpc调用都拷贝一份新的Session进行传输
*/
func (sesid *sessionagent) Clone() gate.Session {
	sesid.lock.Lock()
	tmp := map[string]string{}
	for k, v := range sesid.session.Settings {
		tmp[k] = v
	}
	agent := &sessionagent{
		app:      sesid.app,
		userdata: sesid.userdata,
		lock:     new(sync.RWMutex),
	}
	se := &SessionImp{
		IP:        sesid.session.IP,
		Network:   sesid.session.Network,
		UserId:    sesid.session.UserId,
		SessionId: sesid.session.SessionId,
		ServerId:  sesid.session.ServerId,
		TraceId:   sesid.session.TraceId,
		SpanId:    mqanttools.GenerateID().String(),
		Settings:  tmp,
	}
	agent.session = se
	sesid.lock.Unlock()
	return agent
}

func (sesid *sessionagent) CreateTrace() {
	sesid.session.TraceId = mqanttools.GenerateID().String()
	sesid.session.SpanId = mqanttools.GenerateID().String()
}

func (sesid *sessionagent) TraceID() string {
	return sesid.session.TraceId
}

func (sesid *sessionagent) TraceId() string {
	return sesid.TraceID()
}

func (sesid *sessionagent) SpanID() string {
	return sesid.session.SpanId
}

func (sesid *sessionagent) SpanId() string {
	return sesid.SpanID()
}

func (sesid *sessionagent) ExtractSpan() log.TraceSpan {
	agent := &sessionagent{
		app:      sesid.app,
		userdata: sesid.userdata,
		lock:     new(sync.RWMutex),
	}
	se := &SessionImp{
		IP:        sesid.session.IP,
		Network:   sesid.session.Network,
		UserId:    sesid.session.UserId,
		SessionId: sesid.session.SessionId,
		ServerId:  sesid.session.ServerId,
		TraceId:   sesid.session.TraceId,
		SpanId:    mqanttools.GenerateID().String(),
		Settings:  sesid.session.Settings,
	}
	agent.session = se
	return agent
}

//是否是访客(未登录) ,默认判断规则为 userId==""代表访客
func (sesid *sessionagent) IsGuest() bool {
	if sesid.judgeGuest != nil {
		return sesid.judgeGuest(sesid)
	}
	if sesid.GetUserId() == "" {
		return true
	}
	return false
}

//设置自动的访客判断函数,记得一定要在gate模块设置这个值,以免部分模块因为未设置这个判断函数造成错误的判断
func (sesid *sessionagent) JudgeGuest(judgeGuest func(session gate.Session) bool) {
	sesid.judgeGuest = judgeGuest
}
