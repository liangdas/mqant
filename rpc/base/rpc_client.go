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
package defaultrpc

import (
	"time"
	"fmt"
	"github.com/liangdas/mqant/conf"
	"github.com/liangdas/mqant/log"
	"github.com/liangdas/mqant/utils/uuid"
	"github.com/liangdas/mqant/rpc/pb"
	"github.com/golang/protobuf/proto"
	"github.com/liangdas/mqant/rpc/util"
	"github.com/liangdas/mqant/rpc"
	"github.com/liangdas/mqant/module"
	"github.com/liangdas/mqant/gate"
)

type RPCClient struct {
	app 		module.App
	serverId	string
	remote_client *AMQPClient
	local_client  *LocalClient
}

func NewRPCClient(app 	module.App,serverId string) (mqrpc.RPCClient, error) {
	rpc_client := new(RPCClient)
	rpc_client.serverId=serverId
	rpc_client.app=app
	return rpc_client, nil
}

func (c *RPCClient) NewRemoteClient(info *conf.Rabbitmq) (err error) {
	//创建本地连接
	if info != nil && c.remote_client == nil {
		c.remote_client, err = NewAMQPClient(info)
		if err != nil {
			log.Error("Dial: %s", err)
		}
	}
	return
}

func (c *RPCClient) NewLocalClient(server mqrpc.RPCServer) (err error) {
	//创建本地连接
	if server != nil && server.GetLocalServer() != nil && c.local_client == nil {
		c.local_client, err = NewLocalClient(server.GetLocalServer())
		if err != nil {
			log.Error("Dial: %s", err)
		}
	}
	return
}

func (c *RPCClient) Done() (err error) {
	if c.remote_client != nil {
		err = c.remote_client.Done()
	}
	if c.local_client != nil {
		err = c.local_client.Done()
	}
	return
}


func (c *RPCClient) CallArgs(_func string, ArgsType []string,args [][]byte ) (interface{}, string) {
	var correlation_id = uuid.Rand().Hex()
	rpcInfo := &rpcpb.RPCInfo{
		Fn:      *proto.String(_func),
		Reply:   *proto.Bool(true),
		Expired:   *proto.Int64((time.Now().UTC().Add(time.Second * time.Duration(c.app.GetSettings().Rpc.RpcExpired)).UnixNano()) / 1000000),
		Cid:	*proto.String(correlation_id),
		Args:    args,
		ArgsType:ArgsType,
	}
	callInfo := &mqrpc.CallInfo{
		RpcInfo:      *rpcInfo,
	}
	callback := make(chan rpcpb.ResultInfo, 1)
	var err error


	//优先使用本地rpc
	if c.local_client != nil {
		err = c.local_client.Call(*callInfo, callback)
	} else {
		if c.remote_client != nil {
			err = c.remote_client.Call(*callInfo, callback)
		} else {
			return nil, fmt.Sprintf("rpc service (%s) connection failed",c.serverId)
		}
	}

	if err != nil {
		return nil, err.Error()
	}

	resultInfo, ok := <-callback
	if !ok {
		return nil, "client closed"
	}
	result,err:=argsutil.Bytes2Args(c.app,resultInfo.ResultType,resultInfo.Result)
	return result, resultInfo.Error
}

func (c *RPCClient) CallNRArgs(_func string, ArgsType []string,args [][]byte ) (err error) {
	var correlation_id = uuid.Rand().Hex()
	rpcInfo := &rpcpb.RPCInfo{
		Fn:      *proto.String(_func),
		Reply:   *proto.Bool(false),
		Expired:   *proto.Int64((time.Now().UTC().Add(time.Second * time.Duration(c.app.GetSettings().Rpc.RpcExpired)).UnixNano()) / 1000000),
		Cid:	*proto.String(correlation_id),
		Args:    args,
		ArgsType:ArgsType,
	}
	callInfo := &mqrpc.CallInfo{
		RpcInfo:      *rpcInfo,
	}

	//优先使用本地rpc
	if c.local_client != nil {
		err = c.local_client.CallNR(*callInfo)
	} else {
		if c.remote_client != nil {
			err = c.remote_client.CallNR(*callInfo)
		} else {
			return fmt.Errorf("rpc service (%s) connection failed",c.serverId)
		}
	}

	if err != nil {
		return err
	}
	return nil
}

/**
消息请求 需要回复
*/
func (c *RPCClient) Call(_func string, params ...interface{}) (interface{}, string) {
	var ArgsType []string=make([]string, len(params))
	var args [][]byte=make([][]byte, len(params))
	for k, param := range params {
		var err error=nil
		ArgsType[k],args[k],err=argsutil.ArgsTypeAnd2Bytes(c.app,param)
		if err != nil{
			return nil, fmt.Sprintf( "args[%d] error %s",k,err.Error())
		}

		switch v2:=param.(type) {    //多选语句switch
		case gate.Session:
			//如果参数是这个需要拷贝一份新的再传
			param=v2.Clone()
		}
	}
	return c.CallArgs(_func,ArgsType,args)
}

/**
消息请求 不需要回复
*/
func (c *RPCClient) CallNR(_func string, params ...interface{}) (err error) {
	var ArgsType []string=make([]string, len(params))
	var args [][]byte=make([][]byte, len(params))
	for k, param := range params {
		ArgsType[k],args[k],err=argsutil.ArgsTypeAnd2Bytes(c.app,param)
		if err != nil{
			return  fmt.Errorf( "args[%d] error %s",k,err.Error())
		}

		switch v2:=param.(type) {    //多选语句switch
		case gate.Session:
			//如果参数是这个需要拷贝一份新的再传
			param=v2.Clone()
		}
	}
	return c.CallNRArgs(_func,ArgsType,args)
}
