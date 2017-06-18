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
	"github.com/liangdas/mqant/gate"
	"github.com/liangdas/mqant/log"
	"github.com/liangdas/mqant/utils"
)

type handler struct {
	gate.AgentLearner
	gate.GateHandler
	gate     *Gate
	sessions *utils.BeeMap //连接列表
}

func NewGateHandler(gate *Gate) *handler {
	handler := &handler{
		gate:     gate,
		sessions: utils.NewBeeMap(),
	}
	return handler
}

//当连接建立  并且MQTT协议握手成功
func (h *handler) Connect(a gate.Agent) {
	if a.GetSession() != nil {
		h.sessions.Set(a.GetSession().GetSessionid(), a)
	}
}

//当连接关闭	或者客户端主动发送MQTT DisConnect命令
func (h *handler) DisConnect(a gate.Agent) {
	if a.GetSession() != nil {
		h.sessions.Delete(a.GetSession().GetSessionid())
	}
}

func (h *handler)OnDestroy(){
	for _,v:=range h.sessions.Items(){
		v.(gate.Agent).Close()
	}
	h.sessions.DeleteAll()
}

/**
 *更新整个Session 通常是其他模块拉取最新数据
 */
func (h *handler) Update(Sessionid string) (result gate.Session, err string) {
	agent := h.sessions.Get(Sessionid)
	if agent == nil {
		err = "No Sesssion found"
		return
	}
	result = agent.(gate.Agent).GetSession()
	return
}

/**
 *Bind the session with the the Userid.
 */
func (h *handler) Bind(Sessionid string, Userid string) (result gate.Session, err string) {
	agent := h.sessions.Get(Sessionid)
	if agent == nil {
		err = "No Sesssion found"
		return
	}
	agent.(gate.Agent).GetSession().SetUserid(Userid)

	if h.gate.storage != nil && agent.(gate.Agent).GetSession().GetUserid() != "" {
		//可以持久化
		settings, err := h.gate.storage.Query(Userid)
		if err == nil && settings != nil {
			//有已持久化的数据,可能是上一次连接保存的
			if agent.(gate.Agent).GetSession().GetSettings() == nil {
				agent.(gate.Agent).GetSession().SetSettings(settings)
			} else {
				//合并两个map 并且以 agent.(Agent).GetSession().Settings 已有的优先
				for k, v := range settings {
					if _, ok := agent.(gate.Agent).GetSession().GetSettings()[k]; ok {
						//不用替换
					} else {
						agent.(gate.Agent).GetSession().GetSettings()[k] = v
					}
				}
				//数据持久化
				h.gate.storage.Storage(Userid, agent.(gate.Agent).GetSession().GetSettings())

			}
		}
	}

	result = agent.(gate.Agent).GetSession()
	return
}

/**
 *UnBind the session with the the Userid.
 */
func (h *handler) UnBind(Sessionid string) (result gate.Session, err string) {
	agent := h.sessions.Get(Sessionid)
	if agent == nil {
		err = "No Sesssion found"
		return
	}
	agent.(gate.Agent).GetSession().SetUserid("")
	result = agent.(gate.Agent).GetSession()
	return
}

/**
 *Push the session with the the Userid.
 */
func (h *handler) Push(Sessionid string, Settings map[string]string) (result gate.Session, err string) {
	agent := h.sessions.Get(Sessionid)
	if agent == nil {
		err = "No Sesssion found"
		return
	}
	agent.(gate.Agent).GetSession().SetSettings(Settings)
	result = agent.(gate.Agent).GetSession()
	if h.gate.storage != nil && agent.(gate.Agent).GetSession().GetUserid() != "" {
		err := h.gate.storage.Storage(agent.(gate.Agent).GetSession().GetUserid(), agent.(gate.Agent).GetSession().GetSettings())
		if err != nil {
			log.Error("gate session storage failure")
		}
	}

	return
}

/**
 *Set values (one or many) for the session.
 */
func (h *handler) Set(Sessionid string, key string, value string) (result gate.Session, err string) {
	agent := h.sessions.Get(Sessionid)
	if agent == nil {
		err = "No Sesssion found"
		return
	}
	agent.(gate.Agent).GetSession().GetSettings()[key] = value
	result = agent.(gate.Agent).GetSession()

	if h.gate.storage != nil && agent.(gate.Agent).GetSession().GetUserid() != "" {
		err := h.gate.storage.Storage(agent.(gate.Agent).GetSession().GetUserid(), agent.(gate.Agent).GetSession().GetSettings())
		if err != nil {
			log.Error("gate session storage failure")
		}
	}

	return
}

/**
 *Remove value from the session.
 */
func (h *handler) Remove(Sessionid string, key string) (result interface{}, err string) {
	agent := h.sessions.Get(Sessionid)
	if agent == nil {
		err = "No Sesssion found"
		return
	}
	delete(agent.(gate.Agent).GetSession().GetSettings(), key)
	result = agent.(gate.Agent).GetSession()

	if h.gate.storage != nil && agent.(gate.Agent).GetSession().GetUserid() != "" {
		err := h.gate.storage.Storage(agent.(gate.Agent).GetSession().GetUserid(), agent.(gate.Agent).GetSession().GetSettings())
		if err != nil {
			log.Error("gate session storage failure")
		}
	}

	return
}

/**
 *Send message to the session.
 */
func (h *handler) Send(Sessionid string, topic string, body []byte) (result interface{}, err string) {
	agent := h.sessions.Get(Sessionid)
	if agent == nil {
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
 *主动关闭连接
 */
func (h *handler) Close(Sessionid string) (result interface{}, err string) {
	agent := h.sessions.Get(Sessionid)
	if agent == nil {
		err = "No Sesssion found"
		return
	}
	agent.(gate.Agent).Close()
	return
}
