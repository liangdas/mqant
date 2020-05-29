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

// Package basemodule 服务节点实例定义
package basemodule

import (
	"context"
	"github.com/liangdas/mqant/module"
	"github.com/liangdas/mqant/registry"
	"github.com/liangdas/mqant/rpc"
	"github.com/liangdas/mqant/rpc/base"
)

// NewServerSession 创建一个节点实例
func NewServerSession(app module.App, name string, node *registry.Node) (module.ServerSession, error) {
	session := &serverSession{
		name: name,
		node: node,
		app:  app,
	}
	rpc, err := defaultrpc.NewRPCClient(app, session)
	if err != nil {
		return nil, err
	}
	session.rpc = rpc
	return session, err
}

type serverSession struct {
	node *registry.Node
	name string
	rpc  mqrpc.RPCClient
	app  module.App
}

func (c *serverSession) GetID() string {
	return c.node.Id
}

// Deprecated: 因为命名规范问题函数将废弃,请用GetID代替
func (c *serverSession) GetId() string {
	return c.node.Id
}
func (c *serverSession) GetName() string {
	return c.name
}
func (c *serverSession) GetRPC() mqrpc.RPCClient {
	return c.rpc
}

// Deprecated: 因为命名规范问题函数将废弃,请用GetRPC代替
func (c *serverSession) GetRpc() mqrpc.RPCClient {
	return c.rpc
}

func (c *serverSession) GetApp() module.App {
	return c.app
}
func (c *serverSession) GetNode() *registry.Node {
	return c.node
}

func (c *serverSession) SetNode(node *registry.Node) (err error) {
	c.node = node
	return
}

/**
消息请求 需要回复
*/
func (c *serverSession) Call(ctx context.Context, _func string, params ...interface{}) (interface{}, string) {
	return c.rpc.Call(ctx, _func, params...)
}

/**
消息请求 不需要回复
*/
func (c *serverSession) CallNR(_func string, params ...interface{}) (err error) {
	return c.rpc.CallNR(_func, params...)
}

/**
消息请求 需要回复
*/
func (c *serverSession) CallArgs(ctx context.Context, _func string, ArgsType []string, args [][]byte) (interface{}, string) {
	return c.rpc.CallArgs(ctx, _func, ArgsType, args)
}

/**
消息请求 不需要回复
*/
func (c *serverSession) CallNRArgs(_func string, ArgsType []string, args [][]byte) (err error) {
	return c.rpc.CallNRArgs(_func, ArgsType, args)
}
