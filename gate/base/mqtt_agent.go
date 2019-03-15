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
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/liangdas/mqant/conf"
	"github.com/liangdas/mqant/gate"
	"github.com/liangdas/mqant/gate/base/mqtt"
	"github.com/liangdas/mqant/log"
	"github.com/liangdas/mqant/module"
	"github.com/liangdas/mqant/network"
	"github.com/liangdas/mqant/rpc/util"
	"github.com/liangdas/mqant/utils"
	"github.com/Jeffail/tunny"
)

//type resultInfo struct {
//	Error  string      //错误结果 如果为nil表示请求正确
//	Result interface{} //结果
//}

type agent struct {
	gate.Agent
	module                           module.RPCModule
	session                          gate.Session
	conn                             network.Conn
	r                                *bufio.Reader
	w                                *bufio.Writer
	gate                             gate.Gate
	client                           *mqtt.Client
	gpool 				 *tunny.Pool
	isclose                          bool
	last_storage_heartbeat_data_time int64 //上一次发送存储心跳时间
	rev_num                          int64
	send_num                         int64
}

func NewMqttAgent(module module.RPCModule) *agent {
	a := &agent{
		module: module,
	}
	return a
}
func (this *agent) OnInit(gate gate.Gate, conn network.Conn) error {
	this.gpool = tunny.NewFunc(1, func(pack interface{})interface {}{
		this.recoverworker(pack.(*mqtt.Pack))
		return nil
	})
	this.conn = conn
	this.gate = gate
	this.r = bufio.NewReaderSize(conn, 2048)
	this.w = bufio.NewWriterSize(conn, 2048)
	this.isclose = false
	this.rev_num = 0
	this.send_num = 0
	return nil
}
func (a *agent) IsClosed() bool {
	return a.isclose
}

func (a *agent) GetSession() gate.Session {
	return a.session
}

func (a *agent) Run() (err error) {
	defer func() {
		if err := recover(); err != nil {
			buff := make([]byte, 4096)
			runtime.Stack(buff, false)
			log.Error("conn.serve() panic(%v)\n info:%s", err, string(buff))
		}
		a.Close()

	}()
	//握手协议
	var pack *mqtt.Pack
	pack, err = mqtt.ReadPack(a.r)
	if err != nil {
		log.Error("Read login pack error", err)
		return
	}
	if pack.GetType() != mqtt.CONNECT {
		log.Error("Recive login pack's type error:%v \n", pack.GetType())
		return
	}
	info, ok := (pack.GetVariable()).(*mqtt.Connect)
	if !ok {
		log.Error("It's not a mqtt connection package.")
		return
	}
	//id := info.GetUserName()
	//psw := info.GetPassword()
	//log.Debug("Read login pack %s %s %s %s",*id,*psw,info.GetProtocol(),info.GetVersion())
	c := mqtt.NewClient(conf.Conf.Mqtt, a, a.r, a.w, a.conn, info.GetKeepAlive())
	a.client = c
	a.session, err = NewSessionByMap(a.module.GetApp(), map[string]interface{}{
		"Sessionid": utils.GenerateID().String(),
		"Network":   a.conn.RemoteAddr().Network(),
		"IP":        a.conn.RemoteAddr().String(),
		"Serverid":  a.module.GetServerId(),
		"Settings":  make(map[string]string),
	})
	if err != nil {
		log.Error("gate create agent fail", err.Error())
		return
	}
	a.session.JudgeGuest(a.gate.GetJudgeGuest())
	a.session.CreateTrace()             //代码跟踪
	a.gate.GetAgentLearner().Connect(a) //发送连接成功的事件

	//回复客户端 CONNECT
	err = mqtt.WritePack(mqtt.GetConnAckPack(0), a.w)
	if err != nil {
		return
	}

	c.Listen_loop() //开始监听,直到连接中断
	return nil
}

func (a *agent) OnClose() error {
	a.isclose = true
	a.gate.GetAgentLearner().DisConnect(a) //发送连接断开的事件
	return nil
}

func (a *agent) RevNum() int64 {
	return a.rev_num
}
func (a *agent) SendNum() int64 {
	return a.send_num
}
func (a *agent) OnRecover(pack *mqtt.Pack)  {
	if int(a.gpool.QueueLength())>=a.gpool.GetSize(){
		//协成池用满了
		if a.gpool.GetSize()>=5{
			log.TInfo(nil,"QueueLength full >= %v",a.gpool.QueueLength())
		}else{
			a.gpool.SetSize(a.gpool.GetSize()+1)
		}
	}
	a.gpool.Process(pack)
}
func (a *agent) recoverworker(pack *mqtt.Pack) {
	defer func() {
		if r := recover(); r != nil {
			log.Error("Gate  OnRecover error [%s]", r)
		}
	}()

	toResult := func(a *agent, Topic string, Result interface{}, Error string) error {
		switch v2 := Result.(type) {
		case module.ProtocolMarshal:
			return a.WriteMsg(Topic, v2.GetData())
		}
		b, err := a.module.GetApp().ProtocolMarshal(a.session.TraceId(), Result, Error)
		if err == "" {
			return a.WriteMsg(Topic, b.GetData())
		} else {
			log.Error(err)
			br, _ := a.module.GetApp().ProtocolMarshal(a.session.TraceId(), nil, err)
			return a.WriteMsg(Topic, br.GetData())
		}
		return fmt.Errorf(err)
	}
	//路由服务
	switch pack.GetType() {
	case mqtt.PUBLISH:
		a.rev_num = a.rev_num + 1
		pub := pack.GetVariable().(*mqtt.Publish)
		topics := strings.Split(*pub.GetTopic(), "/")
		a.session.CreateTrace()
		if a.gate.GetRouteHandler() != nil {
			needreturn, result, err := a.gate.GetRouteHandler().OnRoute(a.session, *pub.GetTopic(), pub.GetMsg())
			if err != nil {
				if needreturn {
					toResult(a, *pub.GetTopic(), nil, err.Error())
				}
				return
			} else {
				if needreturn {
					toResult(a, *pub.GetTopic(), result, "")
				}
			}
		} else {
			var msgid string
			if len(topics) < 2 {
				errorstr := "Topic must be [moduleType@moduleID]/[handler]|[moduleType@moduleID]/[handler]/[msgid]"
				log.Error(errorstr)
				toResult(a, *pub.GetTopic(), nil, errorstr)
				return
			} else if len(topics) == 3 {
				msgid = topics[2]
			}
			startsWith := strings.HasPrefix(topics[1], "HD_")
			if !startsWith {
				if msgid != "" {
					toResult(a, *pub.GetTopic(), nil, fmt.Sprintf("Method(%s) must begin with 'HD_'", topics[1]))
				}
				return
			}
			var ArgsType []string = make([]string, 2)
			var args [][]byte = make([][]byte, 2)
			hash := ""
			if a.session.GetUserId() != "" {
				hash = a.session.GetUserId()
			} else {
				hash = a.module.GetServerId()
			}
			//if (a.gate.GetTracingHandler() != nil) && a.gate.GetTracingHandler().OnRequestTracing(a.session, *pub.GetTopic(), pub.GetMsg()) {
			//	a.session.CreateRootSpan("gate")
			//}

			serverSession, err := a.module.GetRouteServer(topics[0], hash)
			if err != nil {
				if msgid != "" {
					toResult(a, *pub.GetTopic(), nil, fmt.Sprintf("Service(type:%s) not found", topics[0]))
				}
				return
			}
			if pub.GetMsg()[0] == '{' && pub.GetMsg()[len(pub.GetMsg())-1] == '}' {
				//尝试解析为json为map
				var obj interface{} // var obj map[string]interface{}
				err := json.Unmarshal(pub.GetMsg(), &obj)
				if err != nil {
					if msgid != "" {
						toResult(a, *pub.GetTopic(), nil, "The JSON format is incorrect")
					}
					return
				}
				ArgsType[1] = argsutil.MAP
				args[1] = pub.GetMsg()
			} else {
				ArgsType[1] = argsutil.BYTES
				args[1] = pub.GetMsg()
			}
			if msgid != "" {
				ArgsType[0] = RPC_PARAM_SESSION_TYPE
				b, err := a.GetSession().Serializable()
				if err != nil {
					return
				}
				args[0] = b
				result, e := serverSession.CallArgs(topics[1], ArgsType, args)
				toResult(a, *pub.GetTopic(), result, e)
			} else {
				ArgsType[0] = RPC_PARAM_SESSION_TYPE
				b, err := a.GetSession().Serializable()
				if err != nil {
					return
				}
				args[0] = b

				e := serverSession.CallNRArgs(topics[1], ArgsType, args)
				if e != nil {
					log.Warning("Gate RPC", e.Error())
				}
			}
		}
		if a.GetSession().GetUserId() != "" {
			//这个链接已经绑定Userid
			interval := time.Now().Unix() - a.last_storage_heartbeat_data_time //单位秒
			if interval > a.gate.GetMinStorageHeartbeat() {
				//如果用户信息存储心跳包的时长已经大于一秒
				if a.gate.GetStorageHandler() != nil {
					a.gate.GetStorageHandler().Heartbeat(a.GetSession().GetUserId())
					a.last_storage_heartbeat_data_time = time.Now().Unix()
				}
			}
		}
	case mqtt.PINGREQ:
		//客户端发送的心跳包
		if a.GetSession().GetUserId() != "" {
			//这个链接已经绑定Userid
			interval := time.Now().Unix() - a.last_storage_heartbeat_data_time //单位秒
			if interval > a.gate.GetMinStorageHeartbeat() {
				//如果用户信息存储心跳包的时长已经大于60秒
				if a.gate.GetStorageHandler() != nil {
					a.gate.GetStorageHandler().Heartbeat(a.GetSession().GetUserId())
					a.last_storage_heartbeat_data_time = time.Now().Unix()
				}
			}
		}
	}
}

func (a *agent) WriteMsg(topic string, body []byte) error {
	a.send_num++
	return a.client.WriteMsg(topic, body)
}

func (a *agent) Close() {
	a.conn.Close()
}

func (a *agent) Destroy() {
	a.conn.Destroy()
	a.gpool.Close()
}
