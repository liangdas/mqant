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

// Package basegate handler
package basegate

import (
	"fmt"
	"runtime"

	"github.com/liangdas/mqant/gate"
	"github.com/liangdas/mqant/log"
	"github.com/pkg/errors"
	"strings"
	"sync"
)

type handler struct {
	//gate.AgentLearner
	//gate.GateHandler
	lock     sync.RWMutex
	gate     gate.Gate
	sessions sync.Map //连接列表
	agentNum int
}

// NewGateHandler NewGateHandler
func NewGateHandler(gate gate.Gate) *handler {
	handler := &handler{
		gate: gate,
	}
	return handler
}

//当连接建立  并且MQTT协议握手成功
func (h *handler) Connect(a gate.Agent) {
	defer func() {
		if err := recover(); err != nil {
			buff := make([]byte, 1024)
			runtime.Stack(buff, false)
			log.Error("handler Connect panic(%v)\n info:%s", err, string(buff))
		}
	}()
	if a.GetSession() != nil {
		h.sessions.Store(a.GetSession().GetSessionID(), a)
		//已经建联成功的才计算
		if a.ProtocolOK() {
			h.lock.Lock()
			h.agentNum++
			h.lock.Unlock()
		}
	}
	if h.gate.GetSessionLearner() != nil {
		go func() {
			h.gate.GetSessionLearner().Connect(a.GetSession())
		}()
	}
}

//当连接关闭	或者客户端主动发送MQTT DisConnect命令
func (h *handler) DisConnect(a gate.Agent) {
	defer func() {
		if err := recover(); err != nil {
			buff := make([]byte, 1024)
			runtime.Stack(buff, false)
			log.Error("handler DisConnect panic(%v)\n info:%s", err, string(buff))
		}
		if a.GetSession() != nil {
			h.sessions.Delete(a.GetSession().GetSessionID())
			//已经建联成功的才计算
			if a.ProtocolOK() {
				h.lock.Lock()
				h.agentNum--
				h.lock.Unlock()
			}
		}
	}()
	if h.gate.GetSessionLearner() != nil {
		if a.GetSession() != nil {
			//没有session的就不返回了
			h.gate.GetSessionLearner().DisConnect(a.GetSession())
		}
	}
}

func (h *handler) OnDestroy() {
	h.sessions.Range(func(key, value interface{}) bool {
		value.(gate.Agent).Close()
		h.sessions.Delete(key)
		return true
	})
}

func (h *handler) GetAgentNum() int {
	num := 0
	h.lock.RLock()
	num = h.agentNum
	h.lock.RUnlock()
	return num
}

/**
 *更新整个Session 通常是其他模块拉取最新数据
 */
func (h *handler) GetAgent(Sessionid string) (gate.Agent, error) {
	agent, ok := h.sessions.Load(Sessionid)
	if !ok || agent == nil {
		return nil, errors.New("No Sesssion found")
	}
	return agent.(gate.Agent), nil
}

/**
 *更新整个Session 通常是其他模块拉取最新数据
 */
func (h *handler) Update(span log.TraceSpan, Sessionid string) (result gate.Session, err string) {
	agent, ok := h.sessions.Load(Sessionid)
	if !ok || agent == nil {
		err = "No Sesssion found"
		return
	}
	result = agent.(gate.Agent).GetSession()
	return
}

/**
 *Bind the session with the the Userid.
 */
func (h *handler) Bind(span log.TraceSpan, Sessionid string, Userid string) (result gate.Session, err string) {
	agent, ok := h.sessions.Load(Sessionid)
	if !ok || agent == nil {
		err = "No Sesssion found"
		return
	}
	agent.(gate.Agent).GetSession().SetUserID(Userid)

	if h.gate.GetStorageHandler() != nil && agent.(gate.Agent).GetSession().GetUserID() != "" {
		//可以持久化
		data, err := h.gate.GetStorageHandler().Query(Userid)
		if err == nil && data != nil {
			//有已持久化的数据,可能是上一次连接保存的
			impSession, err := h.gate.NewSession(data)
			if err == nil {
				if agent.(gate.Agent).GetSession() == nil {
					agent.(gate.Agent).GetSession().SetSettings(impSession.CloneSettings())
				} else {
					//合并两个map 并且以 agent.(Agent).GetSession().Settings 已有的优先
					settings := impSession.CloneSettings()
					_ = agent.(gate.Agent).GetSession().ImportSettings(settings)
				}
			} else {
				//解析持久化数据失败
				log.Warning("Sesssion Resolve fail %s", err.Error())
			}
		}
		//数据持久化
		_ = h.gate.GetStorageHandler().Storage(agent.(gate.Agent).GetSession())
	}

	result = agent.(gate.Agent).GetSession()
	return
}

/**
 *查询某一个userId是否连接中，这里只是查询这一个网关里面是否有userId客户端连接，如果有多个网关就需要遍历了
 */
func (h *handler) IsConnect(span log.TraceSpan, Sessionid string, Userid string) (bool, string) {
	isconnect := false
	found := false
	h.sessions.Range(func(key, agent interface{}) bool {
		if agent.(gate.Agent).GetSession().GetUserID() == Userid {
			isconnect = !agent.(gate.Agent).IsClosed()
			found = true
			return false
		}
		return true
	})
	if !found {
		return false, fmt.Sprintf("The gateway did not find the corresponding userId 【%s】", Userid)
	}
	return isconnect, ""
}

/**
 *UnBind the session with the the Userid.
 */
func (h *handler) UnBind(span log.TraceSpan, Sessionid string) (result gate.Session, err string) {
	agent, ok := h.sessions.Load(Sessionid)
	if !ok || agent == nil {
		err = "No Sesssion found"
		return
	}
	agent.(gate.Agent).GetSession().SetUserID("")
	result = agent.(gate.Agent).GetSession()
	return
}

/**
 *Push the session with the the Userid.
 */
func (h *handler) Push(span log.TraceSpan, Sessionid string, Settings map[string]string) (result gate.Session, err string) {
	agent, ok := h.sessions.Load(Sessionid)
	if !ok || agent == nil {
		err = "No Sesssion found"
		return
	}
	//覆盖当前map对应的key-value
	for key, value := range Settings {
		_ = agent.(gate.Agent).GetSession().SetLocalKV(key, value)
	}
	result = agent.(gate.Agent).GetSession()
	if h.gate.GetStorageHandler() != nil && agent.(gate.Agent).GetSession().GetUserID() != "" {
		err := h.gate.GetStorageHandler().Storage(agent.(gate.Agent).GetSession())
		if err != nil {
			log.Warning("gate session storage failure : %s", err.Error())
		}
	}

	return
}

/**
 *Set values (one or many) for the session.
 */
func (h *handler) Set(span log.TraceSpan, Sessionid string, key string, value string) (result gate.Session, err string) {
	agent, ok := h.sessions.Load(Sessionid)
	if !ok || agent == nil {
		err = "No Sesssion found"
		return
	}
	_ = agent.(gate.Agent).GetSession().SetLocalKV(key, value)
	result = agent.(gate.Agent).GetSession()

	if h.gate.GetStorageHandler() != nil && agent.(gate.Agent).GetSession().GetUserID() != "" {
		err := h.gate.GetStorageHandler().Storage(agent.(gate.Agent).GetSession())
		if err != nil {
			log.Error("gate session storage failure : %s", err.Error())
		}
	}

	return
}

/**
 *Remove value from the session.
 */
func (h *handler) Remove(span log.TraceSpan, Sessionid string, key string) (result interface{}, err string) {
	agent, ok := h.sessions.Load(Sessionid)
	if !ok || agent == nil {
		err = "No Sesssion found"
		return
	}
	_ = agent.(gate.Agent).GetSession().RemoveLocalKV(key)
	result = agent.(gate.Agent).GetSession()

	if h.gate.GetStorageHandler() != nil && agent.(gate.Agent).GetSession().GetUserID() != "" {
		err := h.gate.GetStorageHandler().Storage(agent.(gate.Agent).GetSession())
		if err != nil {
			log.Error("gate session storage failure :%s", err.Error())
		}
	}

	return
}

/**
 *Send message to the session.
 */
func (h *handler) Send(span log.TraceSpan, Sessionid string, topic string, body []byte) (result interface{}, err string) {
	agent, ok := h.sessions.Load(Sessionid)
	if !ok || agent == nil {
		err = "No Sesssion found"
		return
	}
	e := agent.(gate.Agent).WriteMsg(topic, body)
	if e != nil {
		err = e.Error()
	} else {
		result = "success"
	}
	return
}

/**
 *批量发送消息,sessionid之间用,分割
 */
func (h *handler) SendBatch(span log.TraceSpan, SessionidStr string, topic string, body []byte) (int64, string) {
	sessionids := strings.Split(SessionidStr, ",")
	var count int64 = 0
	for _, sessionid := range sessionids {
		agent, ok := h.sessions.Load(sessionid)
		if !ok || agent == nil {
			continue
		}
		e := agent.(gate.Agent).WriteMsg(topic, body)
		if e != nil {
			log.Warning("WriteMsg error: %v", e.Error())
		} else {
			count++
		}
	}
	return count, ""
}
func (h *handler) BroadCast(span log.TraceSpan, topic string, body []byte) (int64, string) {
	var count int64 = 0
	h.sessions.Range(func(key, agent interface{}) bool {
		e := agent.(gate.Agent).WriteMsg(topic, body)
		if e != nil {
			log.Warning("WriteMsg error:", e.Error())
		} else {
			count++
		}
		return true
	})
	return count, ""
}

/**
 *主动关闭连接
 */
func (h *handler) Close(span log.TraceSpan, Sessionid string) (result interface{}, err string) {
	agent, ok := h.sessions.Load(Sessionid)
	if !ok || agent == nil {
		err = "No Sesssion found"
		return
	}
	agent.(gate.Agent).Close()
	return
}
