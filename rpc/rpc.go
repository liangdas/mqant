// Copyright 2014 loolgame Author. All Rights Reserved.
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
package mqrpc

import (
	"github.com/liangdas/mqant/conf"
	"github.com/liangdas/mqant/gate"
	"github.com/liangdas/mqant/rpc/pb"
)

type FunctionInfo struct {
	Function  interface{}
	Goroutine bool
}

type MQServer interface {
	Callback(callinfo CallInfo) error
}

type CallInfo struct {
	RpcInfo rpcpb.RPCInfo
	Result  rpcpb.ResultInfo
	Props   map[string]interface{}
	Agent   MQServer //代理者  AMQPServer / LocalServer 都继承 Callback(callinfo CallInfo)(error) 方法
}
type RPCListener interface {
	/**
	BeforeHandle会对请求做一些前置处理，如：检查当前玩家是否已登录，打印统计日志等。
	@session  可能为nil
	return error  当error不为nil时将直接返回改错误信息而不会再执行后续调用
	*/
	BeforeHandle(fn string, session gate.Session, callInfo *CallInfo) error
	OnTimeOut(fn string, Expired int64)
	OnError(fn string, callInfo *CallInfo, err error)
	/**
	fn 		方法名
	params		参数
	result		执行结果
	exec_time 	方法执行时间 单位为 Nano 纳秒  1000000纳秒等于1毫秒
	*/
	OnComplete(fn string, callInfo *CallInfo, result *rpcpb.ResultInfo, exec_time int64)
}

type GoroutineControl interface {
	Wait() error
	Finish()
}

type RPCServer interface {
	NewRabbitmqRPCServer(info *conf.Rabbitmq) (err error)
	NewRedisRPCServer(info *conf.Redis) (err error)
	NewUdpRPCServer(info *conf.UDP) (err error)
	SetListener(listener RPCListener)
	SetGoroutineControl(control GoroutineControl)
	GetExecuting() int64
	GetLocalServer() LocalServer
	Register(id string, f interface{})
	RegisterGO(id string, f interface{})
	Done() (err error)
}

type RPCClient interface {
	NewRabbitmqClient(info *conf.Rabbitmq) (err error)
	NewRedisClient(info *conf.Redis) (err error)
	NewUdpClient(info *conf.UDP) (err error)
	NewLocalClient(server RPCServer) (err error)
	Done() (err error)
	CallArgs(_func string, ArgsType []string, args [][]byte) (interface{}, string)
	CallNRArgs(_func string, ArgsType []string, args [][]byte) (err error)
	Call(_func string, params ...interface{}) (interface{}, string)
	CallNR(_func string, params ...interface{}) (err error)
	//不可靠的RPC传输,底层基于udp协议
	CallArgsUnreliable(_func string, ArgsType []string, args [][]byte) (interface{}, string)
	//不可靠的RPC传输,底层基于udp协议
	CallUnreliable(_func string, params ...interface{}) (interface{}, string)
}

type LocalClient interface {
	Done() error
	Call(callInfo CallInfo, callback chan rpcpb.ResultInfo) (err error)
	CallNR(callInfo CallInfo) (err error)
}
type LocalServer interface {
	IsClose() bool
	Write(callInfo CallInfo) error
	StopConsume() error
	Shutdown() (err error)
	Callback(callinfo CallInfo) error
}
