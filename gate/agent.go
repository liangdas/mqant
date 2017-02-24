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
	"github.com/liangdas/mqant/network"
	"github.com/liangdas/mqant/mqtt"
	"bufio"
	"runtime"
	"fmt"
	"strings"
	"encoding/json"
	"github.com/liangdas/mqant/conf"
	"github.com/liangdas/mqant/utils/uuid"
)

type resultInfo struct {
	Error   string  //错误结果 如果为nil表示请求正确
	Result  interface{}	//结果
}

type agent struct {
	Agent
	session *Session
	conn    network.Conn
	r  	*bufio.Reader
	w  	*bufio.Writer
	gate    *Gate
	client 	*client
	isclose  bool
}
func (a *agent) IsClosed()(bool){
	return a.isclose
}

func (a *agent) GetSession()(*Session){
	return a.session
}

func (a *agent) Run() (err error){
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
		log.Error("Read login pack error",err)
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
	c := newClient(conf.Conf.Mqtt,a,a.r, a.w, a.conn,info.GetKeepAlive())
	a.client=c
	a.session=NewSession(a.gate.App,map[string]interface{}{
		"Sessionid"	:Get_uuid(),
		"Serverid"    	:a.gate.server.GetId(),
		"Settings"	:make(map[string]interface{}),
	})
	a.gate.agentLearner.Connect(a) //发送连接成功的事件

	//回复客户端 CONNECT
	err = mqtt.WritePack(mqtt.GetConnAckPack(0), a.w)
	if err != nil {
		return
	}

	c.listen_loop()	//开始监听,直到连接中断
	return nil
}

func (a *agent) OnClose() error{
	a.isclose=true
	a.gate.agentLearner.DisConnect(a) //发送连接断开的事件
	return nil
}
func (a *agent) OnRecover(pack *mqtt.Pack){
	toResult:= func(a *agent,Topic string,Result interface{},Error string) (err error){
		r:=&resultInfo{
			Error:Error,
			Result:Result,
		}
		b,err:=json.Marshal(r)
		if err==nil{
			a.WriteMsg(Topic,b)
		}
		return
	}
	//路由服务
	switch pack.GetType() {
	case mqtt.PUBLISH:
		pub := pack.GetVariable().(*mqtt.Publish)
		topics:=strings.Split(*pub.GetTopic(),"/")
		var msgid string
		if len(topics)<2{
			log.Error("Topic must be [serverType]/[handler]|[serverType]/[handler]/[msgid]")
			return
		}else if len(topics)==3{
			msgid=topics[2]
		}

		var obj interface{} // var obj map[string]interface{}
		err:=json.Unmarshal(pub.GetMsg(), &obj)
		if err!=nil{
			if msgid!=""{
				toResult(a,*pub.GetTopic(),nil,"body must be JSON format")
			}
			return
		}
		serverSession,err:=a.gate.App.GetRouteServersByType(topics[0])
		if err!=nil{
			if msgid!=""{
				toResult(a,*pub.GetTopic(),nil,fmt.Sprintf("Service(type:%s) not found",topics[0]))
			}
			return
		}
		startsWith := strings.HasPrefix(topics[1], "HD_")
		if !startsWith{
			if msgid!=""{
				toResult(a,*pub.GetTopic(),nil,fmt.Sprintf("Method(%s) must begin with 'HD_'",topics[1]))
			}
			return
		}
		result,e:=serverSession.Call(topics[1],a.GetSession().ExportMap(),obj)
		toResult(a,*pub.GetTopic(),result,e)
	}
}

func (a *agent) WriteMsg(topic  string,body []byte) error{
	return a.client.WriteMsg(topic,body)
}


func (a *agent) Close() {
	a.conn.Close()
}

func (a *agent) Destroy() {
	a.conn.Destroy()
}


func Get_uuid() string {
	return uuid.Rand().Hex()
}
