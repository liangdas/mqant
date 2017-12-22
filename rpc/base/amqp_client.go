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
	"github.com/golang/protobuf/proto"
	"github.com/liangdas/mqant/conf"
	"github.com/liangdas/mqant/log"
	"github.com/liangdas/mqant/rpc"
	"github.com/liangdas/mqant/rpc/pb"
	"github.com/liangdas/mqant/rpc/util"
	"github.com/liangdas/mqant/utils"
	"github.com/streadway/amqp"
	"sync"
	"time"
)

type AMQPClient struct {
	//callinfos map[string]*ClinetCallInfo
	callinfos   *utils.BeeMap
	cmutex      sync.Mutex //操作callinfos的锁
	rabbitAgent *RabbitAgent
	done        chan error
}

type ClinetCallInfo struct {
	correlation_id string
	timeout        int64 //超时
	call           chan rpcpb.ResultInfo
}

func NewAMQPClient(info *conf.Rabbitmq) (client *AMQPClient, err error) {
	agent, err := NewRabbitAgent(info, TypeClient)
	if err != nil {
		return nil, fmt.Errorf("rabbit agent: %s", err.Error())
	}
	client = new(AMQPClient)
	client.callinfos = utils.NewBeeMap()
	client.rabbitAgent = agent
	client.done = make(chan error)
	go client.on_response_handle(agent.ReadMsg(), client.done)
	return client, nil
	//log.Printf("shutting down")
	//
	//if err := c.Shutdown(); err != nil {
	//	log.Fatalf("error during shutdown: %s", err)
	//}
}

func (c *AMQPClient) Done() (err error) {
	//关闭amqp链接通道
	err = c.rabbitAgent.Shutdown()
	//close(c.send_chan)
	//c.send_done<-nil
	//c.done<-nil
	//清理 callinfos 列表
	for key, clinetCallInfo := range c.callinfos.Items() {
		if clinetCallInfo != nil {
			//关闭管道
			close(clinetCallInfo.(ClinetCallInfo).call)
			//从Map中删除
			c.callinfos.Delete(key)
		}
	}
	c.callinfos = nil
	return
}

/**
消息请求
*/
func (c *AMQPClient) Call(callInfo mqrpc.CallInfo, callback chan rpcpb.ResultInfo) error {
	//var err error
	if c.callinfos == nil {
		return fmt.Errorf("AMQPClient is closed")
	}
	callback_queue, err := c.rabbitAgent.CallbackQueue()
	if err != nil {
		return err
	}
	callInfo.RpcInfo.ReplyTo = callback_queue
	var correlation_id = callInfo.RpcInfo.Cid

	clinetCallInfo := &ClinetCallInfo{
		correlation_id: correlation_id,
		call:           callback,
		timeout:        callInfo.RpcInfo.Expired,
	}
	c.callinfos.Set(correlation_id, *clinetCallInfo)
	body, err := c.Marshal(&callInfo.RpcInfo)
	if err != nil {
		return err
	}
	return c.rabbitAgent.ClientPublish(body)
}

/**
消息请求 不需要回复
*/
func (c *AMQPClient) CallNR(callInfo mqrpc.CallInfo) error {
	body, err := c.Marshal(&callInfo.RpcInfo)
	if err != nil {
		return err
	}
	return c.rabbitAgent.ClientPublish(body)
}

func (c *AMQPClient) on_timeout_handle(args interface{}) {
	if c.callinfos != nil {
		//处理超时的请求
		for key, clinetCallInfo := range c.callinfos.Items() {
			if clinetCallInfo != nil {
				var clinetCallInfo = clinetCallInfo.(ClinetCallInfo)
				if clinetCallInfo.timeout < (time.Now().UnixNano() / 1000000) {
					//从Map中删除
					c.callinfos.Delete(key)
					//已经超时了
					resultInfo := &rpcpb.ResultInfo{
						Result:     nil,
						Error:      "timeout: This is Call",
						ResultType: argsutil.NULL,
					}
					//发送一个超时的消息
					clinetCallInfo.call <- *resultInfo
					//关闭管道
					close(clinetCallInfo.call)
				}

			}
		}
	}
}

/**
接收应答信息
*/
func (c *AMQPClient) on_response_handle(deliveries <-chan amqp.Delivery, done chan error) {
	timeout := time.NewTimer(time.Second * 1)
	for {
		select {
		case d, ok := <-deliveries:
			if !ok {
				deliveries = nil
			} else {
				d.Ack(false)
				resultInfo, err := c.UnmarshalResult(d.Body)
				if err != nil {
					log.Error("Unmarshal faild", err)
				} else {
					correlation_id := resultInfo.Cid
					clinetCallInfo := c.callinfos.Get(correlation_id)
					//删除
					c.callinfos.Delete(correlation_id)
					if clinetCallInfo != nil {
						clinetCallInfo.(ClinetCallInfo).call <- *resultInfo
						close(clinetCallInfo.(ClinetCallInfo).call)
					} else {
						//可能客户端已超时了，但服务端处理完还给回调了
						log.Warning("rpc callback no found : [%s]", correlation_id)
					}
				}
			}
		case <-timeout.C:
			timeout.Reset(time.Second * 1)
			c.on_timeout_handle(nil)
		case <-done:
			goto LForEnd
		}
		if deliveries == nil {
			goto LForEnd
		}
	}
LForEnd:
	timeout.Stop()
}

func (c *AMQPClient) UnmarshalResult(data []byte) (*rpcpb.ResultInfo, error) {
	//fmt.Println(msg)
	//保存解码后的数据，Value可以为任意数据类型
	var resultInfo rpcpb.ResultInfo
	err := proto.Unmarshal(data, &resultInfo)
	if err != nil {
		return nil, err
	} else {
		return &resultInfo, err
	}
}

func (c *AMQPClient) Unmarshal(data []byte) (*rpcpb.RPCInfo, error) {
	//fmt.Println(msg)
	//保存解码后的数据，Value可以为任意数据类型
	var rpcInfo rpcpb.RPCInfo
	err := proto.Unmarshal(data, &rpcInfo)
	if err != nil {
		return nil, err
	} else {
		return &rpcInfo, err
	}

	panic("bug")
}

// goroutine safe
func (c *AMQPClient) Marshal(rpcInfo *rpcpb.RPCInfo) ([]byte, error) {
	//map2:= structs.Map(callInfo)
	b, err := proto.Marshal(rpcInfo)
	return b, err
}
