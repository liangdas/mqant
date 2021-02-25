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

// Package gate 长连接网关定义
package gate

import (
	"github.com/liangdas/mqant/log"
	"github.com/liangdas/mqant/network"
	"time"
)

// RPCParamSessionType gate.session 类型
var RPCParamSessionType = "SESSION"

// RPCParamProtocolMarshalType ProtocolMarshal类型
var RPCParamProtocolMarshalType = "ProtocolMarshal"

// JudgeGuest 判断是否为游客
var JudgeGuest func(session Session) bool

// GateHandler net代理服务处理器
type GateHandler interface {
	GetAgent(Sessionid string) (Agent, error)
	GetAgentNum() int
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

//Session session代表一个客户端连接,不是线程安全的
type Session interface {
	GetIP() string
	GetTopic() string
	GetNetwork() string
	// Deprecated: 因为命名规范问题函数将废弃,请用GetUserID代替
	GetUserId() string
	GetUserID() string
	GetUserIdInt64() int64
	// Deprecated: 因为命名规范问题函数将废弃,请用GetUserIDInt64代替
	GetUserIDInt64() int64
	// Deprecated: 因为命名规范问题函数将废弃,请用GetSessionID代替
	GetSessionId() string
	GetSessionID() string
	// Deprecated: 因为命名规范问题函数将废弃,请用GetServerID代替
	GetServerId() string
	GetServerID() string
	//SettingsRange 配合一个回调函数进行遍历操作，通过回调函数返回内部遍历出来的值。Range 参数中的回调函数的返回值功能是：需要继续迭代遍历时，返回 true；终止迭代遍历时，返回 false。
	SettingsRange(func(k, v string) bool)
	// 合并两个map 并且以 agent.(Agent).GetSession().Settings 已有的优先
	ImportSettings(map[string]string) error
	//网关本地的额外数据,不会再rpc中传递
	LocalUserData() interface{}
	SetIP(ip string)
	SetTopic(topic string)
	SetNetwork(network string)
	// Deprecated: 因为命名规范问题函数将废弃,请用SetUserID代替
	SetUserId(userid string)
	SetUserID(userid string)
	// Deprecated: 因为命名规范问题函数将废弃,请用SetSessionID代替
	SetSessionId(sessionid string)
	SetSessionID(sessionid string)
	// Deprecated: 因为命名规范问题函数将废弃,请用SetServerId代替
	SetServerId(serverid string)
	SetServerID(serverid string)
	SetSettings(settings map[string]string)
	//CloneSettings
	CloneSettings() map[string]string
	SetLocalKV(key, value string) error
	RemoveLocalKV(key string) error
	//网关本地的额外数据,不会再rpc中传递
	SetLocalUserData(data interface{}) error
	Serializable() ([]byte, error)
	Update() (err string)
	Bind(UserID string) (err string)
	UnBind() (err string)
	Push() (err string)
	Set(key string, value string) (err string)
	SetPush(key string, value string) (err string)    //设置值以后立即推送到gate网关,跟Set功能相同
	SetBatch(settings map[string]string) (err string) //批量设置settings,跟当前已存在的settings合并,如果跟当前已存在的key重复则会被新value覆盖
	Get(key string) (result string)
	//Load 跟Get方法类似，但如果key不存在则 ok会返回false
	Load(key string) (result string, ok bool)
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
	// Deprecated: 因为命名规范问题函数将废弃,请用TraceID代替
	TraceId() string
	TraceID() string

	// Span is an ID that probabilistically uniquely identifies this
	// span.
	// Deprecated: 因为命名规范问题函数将废弃,请用SpanID代替
	SpanId() string
	SpanID() string

	ExtractSpan() log.TraceSpan
}

// StorageHandler Session信息持久化
type StorageHandler interface {
	/**
	存储用户的Session信息
	Session Bind Userid以后每次设置 settings都会调用一次Storage
	*/
	Storage(session Session) (err error)
	/**
	强制删除Session信息
	*/
	Delete(session Session) (err error)
	/**
	获取用户Session信息
	Bind Userid时会调用Query获取最新信息
	*/
	Query(Userid string) (data []byte, err error)
	/**
	用户心跳,一般用户在线时1s发送一次
	可以用来延长Session信息过期时间
	*/
	Heartbeat(session Session)
}

// RouteHandler 路由器
type RouteHandler interface {
	/**
	是否需要对本次客户端请求转发规则进行hook
	*/
	OnRoute(session Session, topic string, msg []byte) (bool, interface{}, error)
}

// SendMessageHook 给客户端下发消息拦截器
type SendMessageHook func(session Session, topic string, msg []byte) ([]byte, error)

// AgentLearner 连接代理
type AgentLearner interface {
	Connect(a Agent)    //当连接建立  并且MQTT协议握手成功
	DisConnect(a Agent) //当连接关闭	或者客户端主动发送MQTT DisConnect命令
}

// SessionLearner 客户端代理
type SessionLearner interface {
	Connect(a Session)    //当连接建立  并且MQTT协议握手成功
	DisConnect(a Session) //当连接关闭	或者客户端主动发送MQTT DisConnect命令
}

// Agent 客户端代理定义
type Agent interface {
	OnInit(gate Gate, conn network.Conn) error
	WriteMsg(topic string, body []byte) error
	Close()
	Run() (err error)
	OnClose() error
	Destroy()
	ConnTime() time.Time
	RevNum() int64
	SendNum() int64
	IsClosed() bool
	ProtocolOK() bool
	GetError() error //连接断开的错误日志
	GetSession() Session
}

// Gate 网关代理定义
type Gate interface {
	Options() Options
	GetGateHandler() GateHandler
	GetAgentLearner() AgentLearner
	GetSessionLearner() SessionLearner
	GetStorageHandler() StorageHandler
	GetRouteHandler() RouteHandler
	GetJudgeGuest() func(session Session) bool
	NewSession(data []byte) (Session, error)
	NewSessionByMap(data map[string]interface{}) (Session, error)
}
