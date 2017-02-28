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
	"github.com/liangdas/mqant/network"
	"time"
	"bufio"
	"github.com/liangdas/mqant/module"
	"github.com/liangdas/mqant/conf"
)

type Gate struct {
	module.BaseModule
	MaxConnNum      int
	MaxMsgLen       uint32
	MinStorageHeartbeat int64 //Session持久化最短心跳包

	// websocket
	WSAddr      	string
	HTTPTimeout 	time.Duration

	// tcp
	TCPAddr      	string

	//tls
	Tls		bool
	CertFile	string
	KeyFile		string
	//
	handler		GateHandler
	agentLearner	AgentLearner
	storage		StorageHandler
}
/**
设置Session信息持久化接口
 */
func (gate *Gate) SetStorageHandler(storage StorageHandler)(error){
	gate.storage=storage
	return nil
}

func (gate *Gate) GetStorageHandler()(storage StorageHandler){
	return gate.storage
}

func (gate *Gate) OnInit(subclass module.Module,app module.App,settings *conf.ModuleSettings) {
	gate.BaseModule.OnInit(subclass,app,settings) //这是必须的


	gate.MaxConnNum=int(settings.Settings["MaxConnNum"].(float64))
	gate.MaxMsgLen=uint32(settings.Settings["MaxMsgLen"].(float64))
	gate.WSAddr=settings.Settings["WSAddr"].(string)
	gate.HTTPTimeout=time.Second*time.Duration(settings.Settings["HTTPTimeout"].(float64))
	gate.TCPAddr=settings.Settings["TCPAddr"].(string)
	if Tls,ok:=settings.Settings["Tls"];ok{
		gate.Tls=Tls.(bool)
	}else{
		gate.Tls= false
	}
	if CertFile,ok:=settings.Settings["CertFile"];ok{
		gate.CertFile=CertFile.(string)
	}else{
		gate.CertFile= ""
	}
	if KeyFile,ok:=settings.Settings["KeyFile"];ok{
		gate.KeyFile=KeyFile.(string)
	}else{
		gate.KeyFile= ""
	}


	if MinHBStorage, ok := settings.Settings["MinHBStorage"]; ok {
		gate.MinStorageHeartbeat=int64(MinHBStorage.(float64))
	}else{
		gate.MinStorageHeartbeat=60
	}


	handler:=NewGateHandler(gate)

	gate.agentLearner=handler
	gate.handler=handler


	gate.GetServer().RegisterGO("Update",gate.handler.Update)
	gate.GetServer().RegisterGO("Bind",gate.handler.Bind)
	gate.GetServer().RegisterGO("UnBind",gate.handler.UnBind)
	gate.GetServer().RegisterGO("Push",gate.handler.Push)
	gate.GetServer().RegisterGO("Set",gate.handler.Set)
	gate.GetServer().RegisterGO("Remove",gate.handler.Remove)
	gate.GetServer().RegisterGO("Send",gate.handler.Send)
	gate.GetServer().RegisterGO("Close",gate.handler.Close)
}

func (gate *Gate) Run(closeSig chan bool) {
	var wsServer *network.WSServer
	if gate.WSAddr != "" {
		wsServer = new(network.WSServer)
		wsServer.Addr = gate.WSAddr
		wsServer.MaxConnNum = gate.MaxConnNum
		wsServer.MaxMsgLen = gate.MaxMsgLen
		wsServer.HTTPTimeout = gate.HTTPTimeout
		wsServer.Tls=gate.Tls
		wsServer.CertFile=gate.CertFile
		wsServer.KeyFile=gate.KeyFile
		wsServer.NewAgent = func(conn *network.WSConn) network.Agent {
			a := &agent{
				conn: conn,
				gate: gate,
				r:   bufio.NewReader(conn),
				w:   bufio.NewWriter(conn),
				isclose:false,
			}
			return a
		}
	}

	var tcpServer *network.TCPServer
	if gate.TCPAddr != "" {
		tcpServer = new(network.TCPServer)
		tcpServer.Addr = gate.TCPAddr
		tcpServer.MaxConnNum = gate.MaxConnNum
		tcpServer.Tls=gate.Tls
		tcpServer.CertFile=gate.CertFile
		tcpServer.KeyFile=gate.KeyFile
		tcpServer.NewAgent = func(conn *network.TCPConn) network.Agent {
			a := &agent{
				conn: conn,
				gate: gate,
				r:   bufio.NewReader(conn),
				w:   bufio.NewWriter(conn),
				isclose:false,
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
	if wsServer != nil {
		wsServer.Close()
	}
	if tcpServer != nil {
		tcpServer.Close()
	}
}

func (gate *Gate) OnDestroy() {
	gate.BaseModule.OnDestroy() //这是必须的
}

