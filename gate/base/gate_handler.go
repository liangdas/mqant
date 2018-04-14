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

	"github.com/liangdas/mqant/gate"
	"github.com/liangdas/mqant/log"
	"github.com/liangdas/mqant/utils"
	"strings"
)

type handler struct {
	//gate.AgentLearner
	//gate.GateHandler
	gate     gate.Gate
	sessions *utils.BeeMap //连接列表
}

func NewGateHandler(gate gate.Gate) *handler {
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
	if h.gate.GetSessionLearner() != nil {
		h.gate.GetSessionLearner().Connect(a.GetSession())
	}
}

//当连接关闭	或者客户端主动发送MQTT DisConnect命令
func (h *handler) DisConnect(a gate.Agent) {
	if a.GetSession() != nil {
		h.sessions.Delete(a.GetSession().GetSessionid())
	}
	if h.gate.GetSessionLearner() != nil {
		h.gate.GetSessionLearner().DisConnect(a.GetSession())
	}
}

func (h *handler) OnDestroy() {
	for _, v := range h.sessions.Items() {
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

	if h.gate.GetStorageHandler() != nil && agent.(gate.Agent).GetSession().GetUserid() != "" {
		//可以持久化
		data, err := h.gate.GetStorageHandler().Query(Userid)
		if err == nil && data != nil {
			//有已持久化的数据,可能是上一次连接保存的
			impSession, err := h.gate.NewSession(data)
			if err == nil {
				if agent.(gate.Agent).GetSession().GetSettings() == nil {
					agent.(gate.Agent).GetSession().SetSettings(impSession.GetSettings())
				} else {
					//合并两个map 并且以 agent.(Agent).GetSession().Settings 已有的优先
					settings := impSession.GetSettings()
					if settings != nil {
						for k, v := range settings {
							if _, ok := agent.(gate.Agent).GetSession().GetSettings()[k]; ok {
								//不用替换
							} else {
								agent.(gate.Agent).GetSession().GetSettings()[k] = v
							}
						}
					}
					//数据持久化
					h.gate.GetStorageHandler().Storage(Userid, agent.(gate.Agent).GetSession())
				}
			} else {
				//解析持久化数据失败
				log.Warning("Sesssion Resolve fail %s", err.Error())
			}
		}
	}

	result = agent.(gate.Agent).GetSession()
	return
}

/**
 *查询某一个userId是否连接中，这里只是查询这一个网关里面是否有userId客户端连接，如果有多个网关就需要遍历了
 */
func (h *handler) IsConnect(Sessionid string, Userid string) (bool, string) {

	for _, agent := range h.sessions.Items() {
		if agent.(gate.Agent).GetSession().GetUserid() == Userid {
			return !agent.(gate.Agent).IsClosed(), ""
		}
	}

	return false, fmt.Sprintf("The gateway did not find the corresponding userId 【%s】", Userid)
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
	if h.gate.GetStorageHandler() != nil && agent.(gate.Agent).GetSession().GetUserid() != "" {
		err := h.gate.GetStorageHandler().Storage(agent.(gate.Agent).GetSession().GetUserid(), agent.(gate.Agent).GetSession())
		if err != nil {
			log.Warning("gate session storage failure : %s", err.Error())
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

	if h.gate.GetStorageHandler() != nil && agent.(gate.Agent).GetSession().GetUserid() != "" {
		err := h.gate.GetStorageHandler().Storage(agent.(gate.Agent).GetSession().GetUserid(), agent.(gate.Agent).GetSession())
		if err != nil {
			log.Error("gate session storage failure : %s", err.Error())
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

	if h.gate.GetStorageHandler() != nil && agent.(gate.Agent).GetSession().GetUserid() != "" {
		err := h.gate.GetStorageHandler().Storage(agent.(gate.Agent).GetSession().GetUserid(), agent.(gate.Agent).GetSession())
		if err != nil {
			log.Error("gate session storage failure :%s", err.Error())
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
 *批量发送消息,sessionid之间用,分割
 */
func (h *handler) SendBatch(SessionidStr string, topic string, body []byte) (int64, string) {
	sessionids := strings.Split(SessionidStr, ",")
	var count int64 = 0
	for _, sessionid := range sessionids {
		agent := h.sessions.Get(sessionid)
		if agent == nil {
			//log.Warning("No Sesssion found")
			continue
		}
		e := agent.(gate.Agent).WriteMsg(topic, body)
		if e != nil {
			log.Warning("WriteMsg error:", e.Error())
		} else {
			count++
		}
	}
	return count, ""
}
func (h *handler) BroadCast(topic string, body []byte) (int64, string) {
	var count int64 = 0
	for _, agent := range h.sessions.Items() {
		e := agent.(gate.Agent).WriteMsg(topic, body)
		if e != nil {
			log.Warning("WriteMsg error:", e.Error())
		} else {
			count++
		}
	}
	return count, ""
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
