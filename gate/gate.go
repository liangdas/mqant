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

var Module = func() (module.Module){
	gate := new(Gate)
	return gate
}

type Gate struct {
	App 		module.App
	server 		*module.Server
	MaxConnNum      int
	MaxMsgLen       uint32

	// websocket
	WSAddr      	string
	HTTPTimeout 	time.Duration

	// tcp
	TCPAddr      	string
	Tls		string
	Cert		string
	Key		string
	//
	handler		GateHandler
	agentLearner	AgentLearner
}
func (gate *Gate) GetServer() (*module.Server){
	if gate.server==nil{
		gate.server = new(module.Server)
	}
	return gate.server
}
func (gate *Gate) GetType()(string){
	return "Gate"
}
func (gate *Gate) OnInit(app module.App,settings *conf.ModuleSettings) {
	gate.MaxConnNum=int(settings.Settings["MaxConnNum"].(float64))
	gate.MaxMsgLen=uint32(settings.Settings["MaxMsgLen"].(float64))
	gate.WSAddr=settings.Settings["WSAddr"].(string)
	gate.HTTPTimeout=time.Second*time.Duration(settings.Settings["HTTPTimeout"].(float64))
	gate.TCPAddr=settings.Settings["TCPAddr"].(string)
	gate.App=app

	handler:=NewGateHandler(gate)

	gate.agentLearner=handler
	gate.handler=handler


	gate.GetServer().OnInit(app,settings) //初始化net代理服务的RPC服务
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
	gate.GetServer().OnDestroy()
}

