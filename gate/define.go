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

/**
net代理服务 处理器
 */
type GateHandler interface {
	Bind(Sessionid	string,Userid string)(result interface{},err string) //Bind the session with the the Userid.
	UnBind(Sessionid string)(result interface{},err string) //UnBind the session with the the Userid.
	Set(Sessionid string,key string, value interface{})(result interface{},err string) //Set values (one or many) for the session.
	Remove(Sessionid string,key string)(result interface{},err string) //Remove value from the session.
	Push(Sessionid string,Settings map[string]interface{})(result interface{},err string)	//推送信息给Session
	Send(Sessionid string,topic  string,body []byte)(result interface{},err string) //Send message to the session.
	Close(Sessionid string)(result interface{},err string)	//主动关闭连接
	Update(Sessionid string)(result interface{},err string) //更新整个Session 通常是其他模块拉取最新数据
}

/**
Session信息持久化
 */
type StorageHandler interface {
	/**
	存储用户的Session信息
	Session Bind Userid以后每次设置 settings都会调用一次Storage
	 */
	Storage(Userid string,settings map[string]interface{})(err error)
	/**
	强制删除Session信息
	 */
	Delete(Userid string)(err error)
	/**
	获取用户Session信息
	用户登录以后会调用Query获取最新信息
	 */
	Query(Userid string)(settings map[string]interface{},err error)
	/**
	用户心跳,一般用户在线时1s发送一次
	可以用来延长Session信息过期时间
	 */
	Heartbeat(Userid string)
}
type AgentLearner interface{
	Connect(a Agent)   //当连接建立  并且MQTT协议握手成功
	DisConnect(a Agent) //当连接关闭	或者客户端主动发送MQTT DisConnect命令
}

type Agent interface {
	WriteMsg(topic  string,body []byte) error
	Close()
	Destroy()
	IsClosed()(bool)
	GetSession()(*Session)
}
