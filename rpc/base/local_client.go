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
	"fmt"
	"github.com/liangdas/mqant/log"
	"github.com/liangdas/mqant/rpc"
	"github.com/liangdas/mqant/rpc/pb"
	"github.com/liangdas/mqant/rpc/util"
	"github.com/liangdas/mqant/utils"
	"sync"
	"time"
)

type LocalClient struct {
	//callinfos map[string]*ClinetCallInfo
	callinfos    utils.ConcurrentMap
	cmutex       sync.Mutex //操作callinfos的锁
	local_server mqrpc.LocalServer
	result_chan  chan rpcpb.ResultInfo
	done         chan error
	isClose      bool
}

func NewLocalClient(server mqrpc.LocalServer) (*LocalClient, error) {
	client := new(LocalClient)
	//client.callinfos=make(map[string]*ClinetCallInfo)
	client.callinfos = utils.New()
	client.local_server = server
	client.done = make(chan error)
	client.isClose = false
	client.result_chan = make(chan rpcpb.ResultInfo, 1)
	go client.on_response_handle(client.result_chan, client.done)
	//go client.on_timeout_handle(nil) //处理超时请求的协程
	return client, nil
	//log.Printf("shutting down")
	//
	//if err := c.Shutdown(); err != nil {
	//	log.Fatalf("error during shutdown: %s", err)
	//}
}

func (c *LocalClient) Done() error {
	//关闭消息回复通道
	c.isClose = true
	c.done <- nil
	close(c.result_chan)
	//清理 callinfos 列表
	for tuple := range c.callinfos.Iter() {
		//关闭管道
		close(tuple.Val.(*ClinetCallInfo).call)
		//从Map中删除
		c.callinfos.Remove(tuple.Key)
	}
	return nil
}

/**
消息请求
*/
func (c *LocalClient) Call(callInfo mqrpc.CallInfo, callback chan rpcpb.ResultInfo) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf(r.(string))
		}
	}()

	if c.isClose {
		return fmt.Errorf("MQClient is closed")
	}

	var correlation_id = callInfo.RpcInfo.Cid

	clinetCallInfo := &ClinetCallInfo{
		correlation_id: correlation_id,
		call:           callback,
		timeout:        callInfo.RpcInfo.Expired,
	}
	c.callinfos.Set(correlation_id, clinetCallInfo)
	callInfo.Props = map[string]interface{}{
		"reply_to": c.result_chan,
	}
	//发送消息
	c.local_server.Write(callInfo)

	return nil
}

/**
消息请求 不需要回复
*/
func (c *LocalClient) CallNR(callInfo mqrpc.CallInfo) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf(r.(string))
		}
	}()
	//发送消息
	c.local_server.Write(callInfo)

	return nil
}

func (c *LocalClient) on_timeout_handle(args interface{}) {
	//处理超时的请求
	for tuple := range c.callinfos.Iter() {
		var clinetCallInfo = tuple.Val.(*ClinetCallInfo)
		if clinetCallInfo.timeout < (time.Now().UnixNano() / 1000000) {
			//已经超时了
			resultInfo := &rpcpb.ResultInfo{
				Result:     nil,
				ResultType: argsutil.NULL,
				Error:      "timeout: This is Call",
			}
			//发送一个超时的消息
			clinetCallInfo.call <- *resultInfo
			//关闭管道
			close(clinetCallInfo.call)
			//从Map中删除
			c.callinfos.Remove(tuple.Key)
		}
	}
}

/**
接收应答信息
*/
func (c *LocalClient) on_response_handle(deliveries <-chan rpcpb.ResultInfo, done chan error) {
	timeout := time.NewTimer(time.Second * 1)
	for {
		select {
		case <-done:
			timeout.Stop()
			goto ForEnd
		case resultInfo, ok := <-deliveries:
			if !ok {
				deliveries = nil
			} else {
				correlation_id := resultInfo.Cid
				clinetCallInfo, ok := c.callinfos.Get(correlation_id)
				if ok {
					c.callinfos.Remove(correlation_id)
					clinetCallInfo.(*ClinetCallInfo).call <- resultInfo
					close(clinetCallInfo.(*ClinetCallInfo).call)
				} else {
					log.Warning("rpc callback no found : [%s]", correlation_id)
				}
			}
		case <-timeout.C:
			timeout.Reset(time.Second * 1)
			c.on_timeout_handle(nil)

		}

		if deliveries == nil {
			timeout.Stop()
			goto ForEnd
		}
	}
ForEnd:
}
