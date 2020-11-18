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

//Package module 模块定义
package module

import (
	"context"
	"github.com/liangdas/mqant/conf"
	"github.com/liangdas/mqant/registry"
	"github.com/liangdas/mqant/rpc"
	"github.com/liangdas/mqant/selector"
	"github.com/nats-io/nats.go"
)

// ProtocolMarshal 数据包装
type ProtocolMarshal interface {
	GetData() []byte
}

// ServerSession 服务代理
type ServerSession interface {
	// Deprecated: 因为命名规范问题函数将废弃,请用GetID代替
	GetId() string
	GetID() string
	GetName() string
	GetRpc() mqrpc.RPCClient
	// Deprecated: 因为命名规范问题函数将废弃,请用GetRPC代替
	GetRPC() mqrpc.RPCClient
	GetApp() App
	GetNode() *registry.Node
	SetNode(node *registry.Node) (err error)
	Call(ctx context.Context, _func string, params ...interface{}) (interface{}, string)
	CallNR(_func string, params ...interface{}) (err error)
	CallArgs(ctx context.Context, _func string, ArgsType []string, args [][]byte) (interface{}, string)
	CallNRArgs(_func string, ArgsType []string, args [][]byte) (err error)
}

//App mqant应用定义
type App interface {
	UpdateOptions(opts ...Option) error
	Run(mods ...Module) error
	SetMapRoute(fn func(app App, route string) string) error
	Configure(settings conf.Config) error
	OnInit(settings conf.Config) error
	OnDestroy() error
	Options() Options
	Transport() *nats.Conn
	Registry() registry.Registry
	// Deprecated: 因为命名规范问题函数将废弃,请用GetServerByID代替
	GetServerById(id string) (ServerSession, error)
	GetServerByID(id string) (ServerSession, error)
	/**
	filter		 调用者服务类型    moduleType|moduleType@moduleID
	Type	   	想要调用的服务类型
	*/
	GetRouteServer(filter string, opts ...selector.SelectOption) (ServerSession, error) //获取经过筛选过的服务
	GetServersByType(Type string) []ServerSession
	GetSettings() conf.Config //获取配置信息

	// Deprecated: 因为命名规范问题函数将废弃,请用Invoke代替
	RpcInvoke(module RPCModule, moduleType string, _func string, params ...interface{}) (interface{}, string)
	// Deprecated: 因为命名规范问题函数将废弃,请用InvokeNR代替
	RpcInvokeNR(module RPCModule, moduleType string, _func string, params ...interface{}) error
	// Deprecated: 因为命名规范问题函数将废弃,请用Call代替
	RpcCall(ctx context.Context, moduleType, _func string, param mqrpc.ParamOption, opts ...selector.SelectOption) (interface{}, string)

	Invoke(module RPCModule, moduleType string, _func string, params ...interface{}) (interface{}, string)
	InvokeNR(module RPCModule, moduleType string, _func string, params ...interface{}) error
	Call(ctx context.Context, moduleType, _func string, param mqrpc.ParamOption, opts ...selector.SelectOption) (interface{}, string)

	/**
	添加一个 自定义参数序列化接口
	gate,system 关键词一被占用请使用其他名称
	*/
	AddRPCSerialize(name string, Interface RPCSerialize) error

	GetRPCSerialize() map[string]RPCSerialize

	GetModuleInited() func(app App, module Module)

	OnConfigurationLoaded(func(app App)) error
	OnModuleInited(func(app App, module Module)) error
	OnStartup(func(app App)) error

	SetProtocolMarshal(protocolMarshal func(Trace string, Result interface{}, Error string) (ProtocolMarshal, string)) error
	/**
	与客户端通信的协议包接口
	*/
	ProtocolMarshal(Trace string, Result interface{}, Error string) (ProtocolMarshal, string)
	NewProtocolMarshal(data []byte) ProtocolMarshal
	GetProcessID() string
	WorkDir() string
}

// Module 基本模块定义
type Module interface {
	Version() string                             //模块版本
	GetType() string                             //模块类型
	OnAppConfigurationLoaded(app App)            //当App初始化时调用，这个接口不管这个模块是否在这个进程运行都会调用
	OnConfChanged(settings *conf.ModuleSettings) //为以后动态服务发现做准备
	OnInit(app App, settings *conf.ModuleSettings)
	OnDestroy()
	GetApp() App
	Run(closeSig chan bool)
}

// RPCModule RPC模块定义
type RPCModule interface {
	context.Context
	Module
	// Deprecated: 因为命名规范问题函数将废弃,请用GetServerID代替
	GetServerId() string //模块类型
	GetServerID() string //模块类型
	// Deprecated: 因为命名规范问题函数将废弃,请用Invoke代替
	RpcInvoke(moduleType string, _func string, params ...interface{}) (interface{}, string)
	Invoke(moduleType string, _func string, params ...interface{}) (interface{}, string)
	// Deprecated: 因为命名规范问题函数将废弃,请用InvokeNR代替
	RpcInvokeNR(moduleType string, _func string, params ...interface{}) error
	InvokeNR(moduleType string, _func string, params ...interface{}) error
	// Deprecated: 因为命名规范问题函数将废弃,请用InvokeArgs代替
	RpcInvokeArgs(moduleType string, _func string, ArgsType []string, args [][]byte) (interface{}, string)
	InvokeArgs(moduleType string, _func string, ArgsType []string, args [][]byte) (interface{}, string)
	// Deprecated: 因为命名规范问题函数将废弃,请用InvokeNRArgs代替
	RpcInvokeNRArgs(moduleType string, _func string, ArgsType []string, args [][]byte) error
	InvokeNRArgs(moduleType string, _func string, ArgsType []string, args [][]byte) error

	// Deprecated: 因为命名规范问题函数将废弃,请用Call代替
	RpcCall(ctx context.Context, moduleType, _func string, param mqrpc.ParamOption, opts ...selector.SelectOption) (interface{}, string)
	//	Call 通用RPC调度函数
	//	ctx 		context.Context 			上下文,可以设置这次请求的超时时间
	//	moduleType	string 						服务名称
	//	_func		string						需要调度的服务方法
	//	param 		mqrpc.ParamOption			方法传参
	//	opts ...selector.SelectOption			服务发现模块过滤，可以用来选择调用哪个服务节点
	Call(ctx context.Context, moduleType, _func string, param mqrpc.ParamOption, opts ...selector.SelectOption) (interface{}, string)
	GetModuleSettings() (settings *conf.ModuleSettings)
	/**
	filter		 调用者服务类型    moduleType|moduleType@moduleID
	Type	   	想要调用的服务类型
	*/
	GetRouteServer(filter string, opts ...selector.SelectOption) (ServerSession, error)
	GetExecuting() int64
}

//RPCSerialize 自定义参数序列化接口
type RPCSerialize interface {
	/**
	序列化 结构体-->[]byte
	param 需要序列化的参数值
	@return ptype 当能够序列化这个值,并且正确解析为[]byte时 返回改值正确的类型,否则返回 ""即可
	@return p 解析成功得到的数据, 如果无法解析该类型,或者解析失败 返回nil即可
	@return err 无法解析该类型,或者解析失败 返回错误信息
	*/
	Serialize(param interface{}) (ptype string, p []byte, err error)
	/**
	反序列化 []byte-->结构体
	ptype 参数类型 与Serialize函数中ptype 对应
	b   参数的字节流
	@return param 解析成功得到的数据结构
	@return err 无法解析该类型,或者解析失败 返回错误信息
	*/
	Deserialize(ptype string, b []byte) (param interface{}, err error)
	/**
	返回这个接口能够处理的所有类型
	*/
	GetTypes() []string
}
