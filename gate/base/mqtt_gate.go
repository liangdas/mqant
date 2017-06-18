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
	"bufio"
	"github.com/liangdas/mqant/gate"
	"github.com/liangdas/mqant/conf"
	"github.com/liangdas/mqant/network"
	"time"
	"github.com/liangdas/mqant/module"
	"github.com/liangdas/mqant/module/base"
	"fmt"
	"reflect"
	"github.com/liangdas/mqant/log"
)
var RPC_PARAM_SESSION_TYPE="SESSION"
type Gate struct {
	module.RPCSerialize
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
	handler      gate.GateHandler
	agentLearner gate.AgentLearner
	storage      gate.StorageHandler
	tracing      gate.TracingHandler
}

/**
设置Session信息持久化接口
*/
func (this *Gate) SetStorageHandler(storage gate.StorageHandler) error {
	this.storage = storage
	return nil
}

/**
设置Session信息持久化接口
*/
func (this *Gate) SetTracingHandler(tracing gate.TracingHandler) error {
	this.tracing = tracing
	return nil
}

func (this *Gate) GetStorageHandler() (storage gate.StorageHandler) {
	return this.storage
}
func (this *Gate)OnConfChanged(settings *conf.ModuleSettings)  {

}

/**
自定义rpc参数序列化反序列化  Session
 */
func (this *Gate)Serialize(param interface{})(ptype string,p []byte, err error){
	switch v2:=param.(type) {
	case gate.Session:
		bytes,err:=v2.Serializable()
		if err != nil{
			return RPC_PARAM_SESSION_TYPE,nil,err
		}
		return RPC_PARAM_SESSION_TYPE,bytes,nil
	default:
		return "", nil,fmt.Errorf("args [%s] Types not allowed",reflect.TypeOf(param))
	}
}

func (this *Gate)Deserialize(ptype string,b []byte)(param interface{},err error){
	switch ptype {
	case RPC_PARAM_SESSION_TYPE:
		mps,errs:= NewSession(this.App,b)
		if errs!=nil{
			return	nil,errs
		}
		return mps,nil
	default:
		return	nil,fmt.Errorf("args [%s] Types not allowed",ptype)
	}
}

func (this *Gate)GetTypes()([]string){
	return []string{RPC_PARAM_SESSION_TYPE}
}

func (this *Gate) OnInit(subclass module.RPCModule, app module.App, settings *conf.ModuleSettings) {
	this.BaseModule.OnInit(subclass, app, settings) //这是必须的

	//添加Session结构体的序列化操作类
	err:=app.AddRPCSerialize("gate",this)
	if err!=nil{
		log.Warning("Adding session structures failed to serialize interfaces",err.Error())
	}

	this.MaxConnNum = int(settings.Settings["MaxConnNum"].(float64))
	this.MaxMsgLen = uint32(settings.Settings["MaxMsgLen"].(float64))
	this.WSAddr = settings.Settings["WSAddr"].(string)
	this.HTTPTimeout = time.Second * time.Duration(settings.Settings["HTTPTimeout"].(float64))
	this.TCPAddr = settings.Settings["TCPAddr"].(string)
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

	this.GetServer().Register("Update", this.handler.Update)
	this.GetServer().Register("Bind", this.handler.Bind)
	this.GetServer().Register("UnBind", this.handler.UnBind)
	this.GetServer().Register("Push", this.handler.Push)
	this.GetServer().Register("Set", this.handler.Set)
	this.GetServer().Register("Remove", this.handler.Remove)
	this.GetServer().Register("Send", this.handler.Send)
	this.GetServer().Register("Close", this.handler.Close)
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
			a := &agent{
				conn:    conn,
				gate:    this,
				r:       bufio.NewReader(conn),
				w:       bufio.NewWriter(conn),
				isclose: false,
				rev_num:0,
				send_num:0,
			}
			return a
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
			a := &agent{
				conn:    conn,
				gate:    this,
				r:       bufio.NewReader(conn),
				w:       bufio.NewWriter(conn),
				isclose: false,
				rev_num:0,
				send_num:0,
			}
			return a
		}
	}

	if wsServer != nil {
		wsServer.Start()
	}
	if tcpServer != nil {
		tcpServer.Start()
	}
	<-closeSig
	if this.handler!=nil{
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
