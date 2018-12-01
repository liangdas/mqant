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
	"reflect"
	"time"

	"github.com/liangdas/mqant/conf"
	"github.com/liangdas/mqant/gate"
	"github.com/liangdas/mqant/log"
	"github.com/liangdas/mqant/module"
	"github.com/liangdas/mqant/module/base"
	"github.com/liangdas/mqant/network"
)

var RPC_PARAM_SESSION_TYPE = gate.RPC_PARAM_SESSION_TYPE
var RPC_PARAM_ProtocolMarshal_TYPE = gate.RPC_PARAM_ProtocolMarshal_TYPE

type Gate struct {
	//module.RPCSerialize
	basemodule.BaseModule
	MaxConnNum          int
	MaxMsgLen           uint32
	MinStorageHeartbeat int64 //Session持久化最短心跳包

	// websocket
	WSAddr      string
	HTTPTimeout time.Duration

	// tcp
	TCPAddr string

	//tls
	Tls      bool
	CertFile string
	KeyFile  string
	//
	handler        gate.GateHandler
	agentLearner   gate.AgentLearner
	sessionLearner gate.SessionLearner
	storage        gate.StorageHandler
	tracing        gate.TracingHandler
	router         gate.RouteHandler
	judgeGuest     func(session gate.Session) bool

	createAgent func() gate.Agent
}

func (this *Gate) defaultCreateAgentd() gate.Agent {
	a := NewMqttAgent(this.GetModule())
	return a
}

func (this *Gate) SetJudgeGuest(judgeGuest func(session gate.Session) bool) error {
	this.judgeGuest = judgeGuest
	return nil
}

/**
设置Session信息持久化接口
*/
func (this *Gate) SetRouteHandler(router gate.RouteHandler) error {
	this.router = router
	return nil
}

/**
设置Session信息持久化接口
*/
func (this *Gate) SetStorageHandler(storage gate.StorageHandler) error {
	this.storage = storage
	return nil
}

/**
设置客户端连接和断开的监听器
*/
func (this *Gate) SetSessionLearner(sessionLearner gate.SessionLearner) error {
	this.sessionLearner = sessionLearner
	return nil
}

/**
设置Session信息持久化接口
*/
func (this *Gate) SetTracingHandler(tracing gate.TracingHandler) error {
	this.tracing = tracing
	return nil
}

/**
设置创建客户端Agent的函数
*/
func (this *Gate) SetCreateAgent(cfunc func() gate.Agent) error {
	this.createAgent = cfunc
	return nil
}

func (this *Gate) GetStorageHandler() (storage gate.StorageHandler) {
	return this.storage
}
func (this *Gate) GetMinStorageHeartbeat() int64 {
	return this.MinStorageHeartbeat
}
func (this *Gate) GetGateHandler() gate.GateHandler {
	return this.handler
}
func (this *Gate) GetAgentLearner() gate.AgentLearner {
	return this.agentLearner
}
func (this *Gate) GetSessionLearner() gate.SessionLearner {
	return this.sessionLearner
}
func (this *Gate) GetTracingHandler() gate.TracingHandler {
	return this.tracing
}
func (this *Gate) GetRouteHandler() gate.RouteHandler {
	return this.router
}
func (this *Gate) GetJudgeGuest() func(session gate.Session) bool {
	return this.judgeGuest
}
func (this *Gate) GetModule() module.RPCModule {
	return this.GetSubclass()
}

func (this *Gate) NewSession(data []byte) (gate.Session, error) {
	return NewSession(this.App, data)
}
func (this *Gate) NewSessionByMap(data map[string]interface{}) (gate.Session, error) {
	return NewSessionByMap(this.App, data)
}

func (this *Gate) OnConfChanged(settings *conf.ModuleSettings) {

}

/**
自定义rpc参数序列化反序列化  Session
*/
func (this *Gate) Serialize(param interface{}) (ptype string, p []byte, err error) {
	switch v2 := param.(type) {
	case gate.Session:
		bytes, err := v2.Serializable()
		if err != nil {
			return RPC_PARAM_SESSION_TYPE, nil, err
		}
		return RPC_PARAM_SESSION_TYPE, bytes, nil
	case module.ProtocolMarshal:
		bytes := v2.GetData()
		return RPC_PARAM_ProtocolMarshal_TYPE, bytes, nil
	default:
		return "", nil, fmt.Errorf("args [%s] Types not allowed", reflect.TypeOf(param))
	}
}

func (this *Gate) Deserialize(ptype string, b []byte) (param interface{}, err error) {
	switch ptype {
	case RPC_PARAM_SESSION_TYPE:
		mps, errs := NewSession(this.App, b)
		if errs != nil {
			return nil, errs
		}
		return mps, nil
	case RPC_PARAM_ProtocolMarshal_TYPE:
		return this.App.NewProtocolMarshal(b), nil
	default:
		return nil, fmt.Errorf("args [%s] Types not allowed", ptype)
	}
}

func (this *Gate) GetTypes() []string {
	return []string{RPC_PARAM_SESSION_TYPE}
}
func (this *Gate) OnAppConfigurationLoaded(app module.App) {
	//添加Session结构体的序列化操作类
	this.BaseModule.OnAppConfigurationLoaded(app) //这是必须的
	err := app.AddRPCSerialize("gate", this)
	if err != nil {
		log.Warning("Adding session structures failed to serialize interfaces %s", err.Error())
	}
}
func (this *Gate) OnInit(subclass module.RPCModule, app module.App, settings *conf.ModuleSettings) {
	this.BaseModule.OnInit(subclass, app, settings) //这是必须的

	this.MaxConnNum = int(settings.Settings["MaxConnNum"].(float64))
	this.MaxMsgLen = uint32(settings.Settings["MaxMsgLen"].(float64))
	if WSAddr, ok := settings.Settings["WSAddr"]; ok {
		this.WSAddr = WSAddr.(string)
	}
	this.HTTPTimeout = time.Second * time.Duration(settings.Settings["HTTPTimeout"].(float64))
	if TCPAddr, ok := settings.Settings["TCPAddr"]; ok {
		this.TCPAddr = TCPAddr.(string)
	}
	if Tls, ok := settings.Settings["Tls"]; ok {
		this.Tls = Tls.(bool)
	} else {
		this.Tls = false
	}
	if CertFile, ok := settings.Settings["CertFile"]; ok {
		this.CertFile = CertFile.(string)
	} else {
		this.CertFile = ""
	}
	if KeyFile, ok := settings.Settings["KeyFile"]; ok {
		this.KeyFile = KeyFile.(string)
	} else {
		this.KeyFile = ""
	}

	if MinHBStorage, ok := settings.Settings["MinHBStorage"]; ok {
		this.MinStorageHeartbeat = int64(MinHBStorage.(float64))
	} else {
		this.MinStorageHeartbeat = 60
	}

	handler := NewGateHandler(this)

	this.agentLearner = handler
	this.handler = handler
	this.GetServer().RegisterGO("Update", this.handler.Update)
	this.GetServer().RegisterGO("Bind", this.handler.Bind)
	this.GetServer().RegisterGO("UnBind", this.handler.UnBind)
	this.GetServer().RegisterGO("Push", this.handler.Push)
	this.GetServer().RegisterGO("Set", this.handler.Set)
	this.GetServer().RegisterGO("Remove", this.handler.Remove)
	this.GetServer().RegisterGO("Send", this.handler.Send)
	this.GetServer().RegisterGO("SendBatch", this.handler.SendBatch)
	this.GetServer().RegisterGO("BroadCast", this.handler.BroadCast)
	this.GetServer().RegisterGO("IsConnect", this.handler.IsConnect)
	this.GetServer().RegisterGO("Close", this.handler.Close)
}

func (this *Gate) Run(closeSig chan bool) {
	var wsServer *network.WSServer
	if this.WSAddr != "" {
		wsServer = new(network.WSServer)
		wsServer.Addr = this.WSAddr
		wsServer.MaxConnNum = this.MaxConnNum
		wsServer.MaxMsgLen = this.MaxMsgLen
		wsServer.HTTPTimeout = this.HTTPTimeout
		wsServer.Tls = this.Tls
		wsServer.CertFile = this.CertFile
		wsServer.KeyFile = this.KeyFile
		wsServer.NewAgent = func(conn *network.WSConn) network.Agent {
			if this.createAgent == nil {
				this.createAgent = this.defaultCreateAgentd
			}
			agent := this.createAgent()
			agent.OnInit(this, conn)
			return agent
		}
	}

	var tcpServer *network.TCPServer
	if this.TCPAddr != "" {
		tcpServer = new(network.TCPServer)
		tcpServer.Addr = this.TCPAddr
		tcpServer.MaxConnNum = this.MaxConnNum
		tcpServer.Tls = this.Tls
		tcpServer.CertFile = this.CertFile
		tcpServer.KeyFile = this.KeyFile
		tcpServer.NewAgent = func(conn *network.TCPConn) network.Agent {
			if this.createAgent == nil {
				this.createAgent = this.defaultCreateAgentd
			}
			agent := this.createAgent()
			agent.OnInit(this, conn)
			return agent
		}
	}

	if wsServer != nil {
		wsServer.Start()
	}
	if tcpServer != nil {
		tcpServer.Start()
	}
	<-closeSig
	if this.handler != nil {
		this.handler.OnDestroy()
	}
	if wsServer != nil {
		wsServer.Close()
	}
	if tcpServer != nil {
		tcpServer.Close()
	}
}

func (this *Gate) OnDestroy() {
	this.BaseModule.OnDestroy() //这是必须的
}
