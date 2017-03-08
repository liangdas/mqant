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
	"github.com/liangdas/mqant/log"
	"github.com/liangdas/mqant/utils"
)

type handler struct {
	AgentLearner
	GateHandler
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
func (h *handler) Connect(a Agent) {
	if a.GetSession() != nil {
		h.sessions.Set(a.GetSession().Sessionid, a)
	}
}

//当连接关闭	或者客户端主动发送MQTT DisConnect命令
func (h *handler) DisConnect(a Agent) {
	if a.GetSession() != nil {
		h.sessions.Delete(a.GetSession().Sessionid)
	}
}

/**
 *更新整个Session 通常是其他模块拉取最新数据
 */
func (h *handler) Update(Sessionid string) (result interface{}, err string) {
	agent := h.sessions.Get(Sessionid)
	if agent == nil {
		err = "No Sesssion found"
		return
	}
	result = agent.(Agent).GetSession().ExportMap()
	return
}

/**
 *Bind the session with the the Userid.
 */
func (h *handler) Bind(Sessionid string, Userid string) (result interface{}, err string) {
	agent := h.sessions.Get(Sessionid)
	if agent == nil {
		err = "No Sesssion found"
		return
	}
	agent.(Agent).GetSession().Userid = Userid

	if h.gate.storage != nil && agent.(Agent).GetSession().Userid != "" {
		//可以持久化
		settings, err := h.gate.storage.Query(Userid)
		if err == nil && settings != nil {
			//有已持久化的数据,可能是上一次连接保存的
			if agent.(Agent).GetSession().Settings == nil {
				agent.(Agent).GetSession().Settings = settings
			} else {
				//合并两个map 并且以 agent.(Agent).GetSession().Settings 已有的优先
				for k, v := range settings {
					if _, ok := agent.(Agent).GetSession().Settings[k]; ok {
						//不用替换
					} else {
						agent.(Agent).GetSession().Settings[k] = v
					}
				}
				//数据持久化
				h.gate.storage.Storage(Userid, agent.(Agent).GetSession().Settings)

			}
		}
	}

	result = agent.(Agent).GetSession().ExportMap()
	return
}

/**
 *UnBind the session with the the Userid.
 */
func (h *handler) UnBind(Sessionid string) (result interface{}, err string) {
	agent := h.sessions.Get(Sessionid)
	if agent == nil {
		err = "No Sesssion found"
		return
	}
	agent.(Agent).GetSession().Userid = ""
	result = agent.(Agent).GetSession().ExportMap()
	return
}

/**
 *Push the session with the the Userid.
 */
func (h *handler) Push(Sessionid string, Settings map[string]interface{}) (result interface{}, err string) {
	agent := h.sessions.Get(Sessionid)
	if agent == nil {
		err = "No Sesssion found"
		return
	}
	agent.(Agent).GetSession().Settings = Settings
	result = agent.(Agent).GetSession().ExportMap()
	if h.gate.storage != nil && agent.(Agent).GetSession().Userid != "" {
		err := h.gate.storage.Storage(agent.(Agent).GetSession().Userid, agent.(Agent).GetSession().Settings)
		if err != nil {
			log.Error("gate session storage failure")
		}
	}

	return
}

/**
 *Set values (one or many) for the session.
 */
func (h *handler) Set(Sessionid string, key string, value interface{}) (result interface{}, err string) {
	agent := h.sessions.Get(Sessionid)
	if agent == nil {
		err = "No Sesssion found"
		return
	}
	agent.(Agent).GetSession().Settings[key] = value
	result = agent.(Agent).GetSession().ExportMap()

	if h.gate.storage != nil && agent.(Agent).GetSession().Userid != "" {
		err := h.gate.storage.Storage(agent.(Agent).GetSession().Userid, agent.(Agent).GetSession().Settings)
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
	delete(agent.(Agent).GetSession().Settings, key)
	result = agent.(Agent).GetSession().ExportMap()

	if h.gate.storage != nil && agent.(Agent).GetSession().Userid != "" {
		err := h.gate.storage.Storage(agent.(Agent).GetSession().Userid, agent.(Agent).GetSession().Settings)
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
	e := agent.(Agent).WriteMsg(topic, body)
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
	agent.(Agent).Close()
	return
}
