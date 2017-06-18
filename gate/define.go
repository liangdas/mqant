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
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/liangdas/mqant/gate/base/mqtt"
)
/**
net代理服务 处理器
*/
type GateHandler interface {
	Bind(Sessionid string, Userid string) (result Session, err string)                   //Bind the session with the the Userid.
	UnBind(Sessionid string) (result Session, err string)                                //UnBind the session with the the Userid.
	Set(Sessionid string, key string, value string) (result Session, err string)    //Set values (one or many) for the session.
	Remove(Sessionid string, key string) (result interface{}, err string)                    //Remove value from the session.
	Push(Sessionid string, Settings map[string]string) (result Session, err string) //推送信息给Session
	Send(Sessionid string, topic string, body []byte) (result interface{}, err string)       //Send message to the session.
	Close(Sessionid string) (result interface{}, err string)                                 //主动关闭连接
	Update(Sessionid string) (result Session, err string)                                //更新整个Session 通常是其他模块拉取最新数据
	OnDestroy()	//退出事件,主动关闭所有的连接
}

type Session interface {
	GetIP() string
	GetNetwork() string
	GetUserid() string
	GetSessionid() string
	GetServerid() string
	GetSettings() map[string]string
	SetIP(ip string)
	SetNetwork(network string)
	SetUserid(userid string)
	SetSessionid(sessionid string)
	SetServerid(serverid string)
	SetSettings(settings map[string]string)
	Serializable()([]byte,error)
	Update() (err string)
	Bind(Userid string) (err string)
	UnBind() (err string)
	Push() (err string)
	Set(key string, value string) (err string)
	Get(key string) (result string)
	Remove(key string) (err string)
	Send(topic string, body []byte) (err string)
	SendNR(topic string, body []byte) (err string)
	Close() (err string)
	Clone()Session
	/**
	通过Carrier数据构造本次rpc调用的tracing Span,如果没有就创建一个新的
	 */
	CreateRootSpan(operationName string)opentracing.Span
	/**
	通过Carrier数据构造本次rpc调用的tracing Span,如果没有就返回nil
	 */
	LoadSpan(operationName string)opentracing.Span
	/**
	获取本次rpc调用的tracing Span
	 */
	Span()opentracing.Span
	/**
	从Session的 Span继承一个新的Span
	 */
	ExtractSpan(operationName string)opentracing.Span
	/**
	获取Tracing的Carrier 可能为nil
	 */
	TracCarrier()map[string]string
}

/**
Session信息持久化
*/
type StorageHandler interface {
	/**
	存储用户的Session信息
	Session Bind Userid以后每次设置 settings都会调用一次Storage
	*/
	Storage(Userid string, settings map[string]string) (err error)
	/**
	强制删除Session信息
	*/
	Delete(Userid string) (err error)
	/**
	获取用户Session信息
	Bind Userid时会调用Query获取最新信息
	*/
	Query(Userid string) (settings map[string]string, err error)
	/**
	用户心跳,一般用户在线时1s发送一次
	可以用来延长Session信息过期时间
	*/
	Heartbeat(Userid string)
}

type TracingHandler interface {
	/**
	是否需要对本次客户端请求进行跟踪
	 */
	OnRequestTracing(session Session,msg *mqtt.Publish)bool
}

type AgentLearner interface {
	Connect(a Agent)    //当连接建立  并且MQTT协议握手成功
	DisConnect(a Agent) //当连接关闭	或者客户端主动发送MQTT DisConnect命令
}

type Agent interface {
	WriteMsg(topic string, body []byte) error
	Close()
	Destroy()
	RevNum() int64
	SendNum() int64
	IsClosed() bool
	GetSession() Session
}
