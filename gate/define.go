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
)

var RPC_PARAM_SESSION_TYPE = "SESSION"
var RPC_PARAM_ProtocolMarshal_TYPE = "ProtocolMarshal"

/**
net代理服务 处理器
*/
type GateHandler interface {
	Bind(span log.TraceSpan, Sessionid string, Userid string) (result Session, err string)                 //Bind the session with the the Userid.
	UnBind(span log.TraceSpan, Sessionid string) (result Session, err string)                              //UnBind the session with the the Userid.
	Set(span log.TraceSpan, Sessionid string, key string, value string) (result Session, err string)       //Set values (one or many) for the session.
	Remove(span log.TraceSpan, Sessionid string, key string) (result interface{}, err string)              //Remove value from the session.
	Push(span log.TraceSpan, Sessionid string, Settings map[string]string) (result Session, err string)    //推送信息给Session
	Send(span log.TraceSpan, Sessionid string, topic string, body []byte) (result interface{}, err string) //Send message
	SendBatch(span log.TraceSpan, Sessionids string, topic string, body []byte) (int64, string)            //批量发送
	BroadCast(span log.TraceSpan, topic string, body []byte) (int64, string)                               //广播消息给网关所有在连客户端
	//查询某一个userId是否连接中，这里只是查询这一个网关里面是否有userId客户端连接，如果有多个网关就需要遍历了
	IsConnect(span log.TraceSpan, Sessionid string, Userid string) (result bool, err string)
	Close(span log.TraceSpan, Sessionid string) (result interface{}, err string) //主动关闭连接
	Update(span log.TraceSpan, Sessionid string) (result Session, err string)    //更新整个Session 通常是其他模块拉取最新数据
	OnDestroy()                                                                  //退出事件,主动关闭所有的连接
}

type Session interface {
	GetIP() string
	GetNetwork() string
	GetUserId() string
	GetUserIdInt64() int64
	GetSessionId() string
	GetServerId() string
	GetSettings() map[string]string
	SetIP(ip string)
	SetNetwork(network string)
	SetUserId(userid string)
	SetSessionId(sessionid string)
	SetServerId(serverid string)
	SetSettings(settings map[string]string)
	Serializable() ([]byte, error)
	Update() (err string)
	Bind(UserId string) (err string)
	UnBind() (err string)
	Push() (err string)
	Set(key string, value string) (err string)
	SetPush(key string, value string) (err string) //设置值以后立即推送到gate网关
	Get(key string) (result string)
	Remove(key string) (err string)
	Send(topic string, body []byte) (err string)
	SendNR(topic string, body []byte) (err string)
	SendBatch(Sessionids string, topic string, body []byte) (int64, string) //想该客户端的网关批量发送消息
	//查询某一个userId是否连接中，这里只是查询这一个网关里面是否有userId客户端连接，如果有多个网关就需要遍历了
	IsConnect(Userid string) (result bool, err string)
	//是否是访客(未登录) ,默认判断规则为 userId==""代表访客
	IsGuest() bool
	//设置自动的访客判断函数,记得一定要在全局的时候设置这个值,以免部分模块因为未设置这个判断函数造成错误的判断
	JudgeGuest(judgeGuest func(session Session) bool)
	Close() (err string)
	Clone() Session

	CreateTrace()

	TraceId() string

	// Span is an ID that probabilistically uniquely identifies this
	// span.
	SpanId() string

	ExtractSpan() log.TraceSpan
}

/**
Session信息持久化
*/
type StorageHandler interface {
	/**
	存储用户的Session信息
	Session Bind Userid以后每次设置 settings都会调用一次Storage
	*/
	Storage(Userid string, session Session) (err error)
	/**
	强制删除Session信息
	*/
	Delete(Userid string) (err error)
	/**
	获取用户Session信息
	Bind Userid时会调用Query获取最新信息
	*/
	Query(Userid string) (data []byte, err error)
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
	OnRequestTracing(session Session, topic string, msg []byte) bool
}

type RouteHandler interface {
	/**
	是否需要对本次客户端请求转发规则进行hook
	*/
	OnRoute(session Session, topic string, msg []byte) (bool, interface{}, error)
}

type AgentLearner interface {
	Connect(a Agent)    //当连接建立  并且MQTT协议握手成功
	DisConnect(a Agent) //当连接关闭	或者客户端主动发送MQTT DisConnect命令
}

type SessionLearner interface {
	Connect(a Session)    //当连接建立  并且MQTT协议握手成功
	DisConnect(a Session) //当连接关闭	或者客户端主动发送MQTT DisConnect命令
}

type Agent interface {
	OnInit(gate Gate, conn network.Conn) error
	WriteMsg(topic string, body []byte) error
	Close()
	Run() (err error)
	OnClose() error
	Destroy()
	RevNum() int64
	SendNum() int64
	IsClosed() bool
	GetSession() Session
}

type Gate interface {
	GetMinStorageHeartbeat() int64
	GetGateHandler() GateHandler
	GetAgentLearner() AgentLearner
	GetSessionLearner() SessionLearner
	GetStorageHandler() StorageHandler
	GetTracingHandler() TracingHandler
	GetRouteHandler() RouteHandler
	GetJudgeGuest() func(session Session) bool
	NewSession(data []byte) (Session, error)
	NewSessionByMap(data map[string]interface{}) (Session, error)
}
