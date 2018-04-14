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
package basemodule

import (
	"github.com/liangdas/mqant/module"
	"github.com/liangdas/mqant/rpc"
)

func NewServerSession(Id string, Stype string, Rpc mqrpc.RPCClient) module.ServerSession {
	session := &serverSession{
		Id:    Id,
		Stype: Stype,
		Rpc:   Rpc,
	}
	return session
}

type serverSession struct {
	Id    string
	Stype string
	Rpc   mqrpc.RPCClient
}

func (c *serverSession) GetId() string {
	return c.Id
}
func (c *serverSession) GetType() string {
	return c.Stype
}
func (c *serverSession) GetRpc() mqrpc.RPCClient {
	return c.Rpc
}

/**
消息请求 需要回复
*/
func (c *serverSession) Call(_func string, params ...interface{}) (interface{}, string) {
	return c.Rpc.Call(_func, params...)
}

/**
使用不可靠的udp rpc传输通道
*/
func (c *serverSession) CallUnreliable(_func string, params ...interface{}) (interface{}, string) {
	return c.Rpc.CallUnreliable(_func, params...)
}

/**
消息请求 不需要回复
*/
func (c *serverSession) CallNR(_func string, params ...interface{}) (err error) {
	return c.Rpc.CallNR(_func, params...)
}

/**
消息请求 需要回复
*/
func (c *serverSession) CallArgs(_func string, ArgsType []string, args [][]byte) (interface{}, string) {
	return c.Rpc.CallArgs(_func, ArgsType, args)
}

/**
消息请求 不需要回复
*/
func (c *serverSession) CallNRArgs(_func string, ArgsType []string, args [][]byte) (err error) {
	return c.Rpc.CallNRArgs(_func, ArgsType, args)
}

/**
使用不可靠的udp rpc传输通道
消息请求 需要回复
*/
func (c *serverSession) CallArgsUnreliable(_func string, ArgsType []string, args [][]byte) (interface{}, string) {
	return c.Rpc.CallArgsUnreliable(_func, ArgsType, args)
}
