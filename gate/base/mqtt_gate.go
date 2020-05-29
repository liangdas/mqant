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

var RPCParamSessionType = gate.RPCParamSessionType
var RPCParamProtocolMarshalType = gate.RPCParamProtocolMarshalType

type Gate struct {
	//module.RPCSerialize
	basemodule.BaseModule
	opts       gate.Options
	judgeGuest func(session gate.Session) bool

	createAgent func() gate.Agent
}

func (gt *Gate) defaultCreateAgentd() gate.Agent {
	a := NewMqttAgent(gt.GetModule())
	return a
}

func (gt *Gate) SetJudgeGuest(judgeGuest func(session gate.Session) bool) error {
	gt.judgeGuest = judgeGuest
	return nil
}

/**
设置Session信息持久化接口
*/
func (gt *Gate) SetRouteHandler(router gate.RouteHandler) error {
	gt.opts.RouteHandler = router
	return nil
}

/**
设置Session信息持久化接口
*/
func (gt *Gate) SetStorageHandler(storage gate.StorageHandler) error {
	gt.opts.StorageHandler = storage
	return nil
}

/**
设置客户端连接和断开的监听器
*/
func (gt *Gate) SetSessionLearner(sessionLearner gate.SessionLearner) error {
	gt.opts.SessionLearner = sessionLearner
	return nil
}

/**
设置创建客户端Agent的函数
*/
func (gt *Gate) SetCreateAgent(cfunc func() gate.Agent) error {
	gt.createAgent = cfunc
	return nil
}
func (gt *Gate) Options() gate.Options {
	return gt.opts
}
func (gt *Gate) GetStorageHandler() (storage gate.StorageHandler) {
	return gt.opts.StorageHandler
}
func (gt *Gate) GetGateHandler() gate.GateHandler {
	return gt.opts.GateHandler
}
func (gt *Gate) GetAgentLearner() gate.AgentLearner {
	return gt.opts.AgentLearner
}
func (gt *Gate) GetSessionLearner() gate.SessionLearner {
	return gt.opts.SessionLearner
}
func (gt *Gate) GetRouteHandler() gate.RouteHandler {
	return gt.opts.RouteHandler
}
func (gt *Gate) GetJudgeGuest() func(session gate.Session) bool {
	return gt.judgeGuest
}
func (gt *Gate) GetModule() module.RPCModule {
	return gt.GetSubclass()
}

func (gt *Gate) NewSession(data []byte) (gate.Session, error) {
	return NewSession(gt.App, data)
}
func (gt *Gate) NewSessionByMap(data map[string]interface{}) (gate.Session, error) {
	return NewSessionByMap(gt.App, data)
}

func (gt *Gate) OnConfChanged(settings *conf.ModuleSettings) {

}

/**
自定义rpc参数序列化反序列化  Session
*/
func (gt *Gate) Serialize(param interface{}) (ptype string, p []byte, err error) {
	rv := reflect.ValueOf(param)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		//不是指针
		return "", nil, fmt.Errorf("Serialize [%v ] or not pointer type", rv.Type())
	}
	switch v2 := param.(type) {
	case gate.Session:
		bytes, err := v2.Serializable()
		if err != nil {
			return RPCParamSessionType, nil, err
		}
		return RPCParamSessionType, bytes, nil
	case module.ProtocolMarshal:
		bytes := v2.GetData()
		return RPCParamProtocolMarshalType, bytes, nil
	default:
		return "", nil, fmt.Errorf("args [%s] Types not allowed", reflect.TypeOf(param))
	}
}

func (gt *Gate) Deserialize(ptype string, b []byte) (param interface{}, err error) {
	switch ptype {
	case RPCParamSessionType:
		mps, errs := NewSession(gt.App, b)
		if errs != nil {
			return nil, errs
		}
		return mps.Clone(), nil
	case RPCParamProtocolMarshalType:
		return gt.App.NewProtocolMarshal(b), nil
	default:
		return nil, fmt.Errorf("args [%s] Types not allowed", ptype)
	}
}

func (gt *Gate) GetTypes() []string {
	return []string{RPCParamSessionType}
}
func (gt *Gate) OnAppConfigurationLoaded(app module.App) {
	//添加Session结构体的序列化操作类
	gt.BaseModule.OnAppConfigurationLoaded(app) //这是必须的
	err := app.AddRPCSerialize("gate", gt)
	if err != nil {
		log.Warning("Adding session structures failed to serialize interfaces %s", err.Error())
	}
}
func (gt *Gate) OnInit(subclass module.RPCModule, app module.App, settings *conf.ModuleSettings, opts ...gate.Option) {
	gt.opts = gate.NewOptions(opts...)
	gt.BaseModule.OnInit(subclass, app, settings, gt.opts.Opts...) //这是必须的
	if gt.opts.WsAddr == "" {
		if WSAddr, ok := settings.Settings["WSAddr"]; ok {
			gt.opts.WsAddr = WSAddr.(string)
		}
	}
	if gt.opts.TCPAddr == "" {
		if TCPAddr, ok := settings.Settings["TCPAddr"]; ok {
			gt.opts.TCPAddr = TCPAddr.(string)
		}
	}

	if gt.opts.TLS == false {
		if tls, ok := settings.Settings["TLS"]; ok {
			gt.opts.TLS = tls.(bool)
		} else {
			gt.opts.TLS = false
		}
	}

	if gt.opts.CertFile == "" {
		if CertFile, ok := settings.Settings["CertFile"]; ok {
			gt.opts.CertFile = CertFile.(string)
		} else {
			gt.opts.CertFile = ""
		}
	}

	if gt.opts.KeyFile == "" {
		if KeyFile, ok := settings.Settings["KeyFile"]; ok {
			gt.opts.KeyFile = KeyFile.(string)
		} else {
			gt.opts.KeyFile = ""
		}
	}

	handler := NewGateHandler(gt)

	gt.opts.AgentLearner = handler
	gt.opts.GateHandler = handler
	gt.GetServer().RegisterGO("Update", gt.opts.GateHandler.Update)
	gt.GetServer().RegisterGO("Bind", gt.opts.GateHandler.Bind)
	gt.GetServer().RegisterGO("UnBind", gt.opts.GateHandler.UnBind)
	gt.GetServer().RegisterGO("Push", gt.opts.GateHandler.Push)
	gt.GetServer().RegisterGO("Set", gt.opts.GateHandler.Set)
	gt.GetServer().RegisterGO("Remove", gt.opts.GateHandler.Remove)
	gt.GetServer().RegisterGO("Send", gt.opts.GateHandler.Send)
	gt.GetServer().RegisterGO("SendBatch", gt.opts.GateHandler.SendBatch)
	gt.GetServer().RegisterGO("BroadCast", gt.opts.GateHandler.BroadCast)
	gt.GetServer().RegisterGO("IsConnect", gt.opts.GateHandler.IsConnect)
	gt.GetServer().RegisterGO("Close", gt.opts.GateHandler.Close)
}

func (gt *Gate) Run(closeSig chan bool) {
	var wsServer *network.WSServer
	if gt.opts.WsAddr != "" {
		wsServer = new(network.WSServer)
		wsServer.Addr = gt.opts.WsAddr
		wsServer.HTTPTimeout = 30 * time.Second
		wsServer.TLS = gt.opts.TLS
		wsServer.CertFile = gt.opts.CertFile
		wsServer.KeyFile = gt.opts.KeyFile
		wsServer.NewAgent = func(conn *network.WSConn) network.Agent {
			if gt.createAgent == nil {
				gt.createAgent = gt.defaultCreateAgentd
			}
			agent := gt.createAgent()
			agent.OnInit(gt, conn)
			return agent
		}
	}

	var tcpServer *network.TCPServer
	if gt.opts.TCPAddr != "" {
		tcpServer = new(network.TCPServer)
		tcpServer.Addr = gt.opts.TCPAddr
		tcpServer.TLS = gt.opts.TLS
		tcpServer.CertFile = gt.opts.CertFile
		tcpServer.KeyFile = gt.opts.KeyFile
		tcpServer.NewAgent = func(conn *network.TCPConn) network.Agent {
			if gt.createAgent == nil {
				gt.createAgent = gt.defaultCreateAgentd
			}
			agent := gt.createAgent()
			agent.OnInit(gt, conn)
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
	if gt.opts.GateHandler != nil {
		gt.opts.GateHandler.OnDestroy()
	}
	if wsServer != nil {
		wsServer.Close()
	}
	if tcpServer != nil {
		tcpServer.Close()
	}
}

func (gt *Gate) OnDestroy() {
	gt.BaseModule.OnDestroy() //这是必须的
}
