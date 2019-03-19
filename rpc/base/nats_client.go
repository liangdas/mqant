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
	"sync"
	"github.com/liangdas/mqant/rpc"
	"time"
	"github.com/liangdas/mqant/rpc/util"
	"github.com/liangdas/mqant/utils"
	"github.com/liangdas/mqant/rpc/pb"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/nats-io/go-nats"
	"github.com/liangdas/mqant/log"
	"runtime"
	"github.com/liangdas/mqant/conf"
)

type NatsClient struct {
			       //callinfos map[string]*ClinetCallInfo
	callinfos   *utils.BeeMap
	cmutex      sync.Mutex //操作callinfos的锁
	nc 		*nats.Conn
	callbackqueueName string
	subs 		*nats.Subscription
	done        chan error
	conf *conf.ModuleSettings
	timeout_done chan error
}

func NewNatsClient(conf *conf.ModuleSettings) (client *NatsClient, err error) {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		return nil, fmt.Errorf("nats agent: %s", err.Error())
	}
	client = new(NatsClient)
	client.conf=conf
	client.callinfos = utils.NewBeeMap()
	client.callbackqueueName = createQueueName()
	client.nc = nc
	client.done = make(chan error)
	client.timeout_done = make(chan error)
	subs,err:=nc.Subscribe(client.callbackqueueName, client.on_response_handle)
	if err != nil {
		return nil, fmt.Errorf("nats agent: %s", err.Error())
	}
	client.subs=subs
	go client.on_timeout_handle(client.timeout_done) //处理超时请求的协程
	return client, nil
}

func (c *NatsClient) Done() (err error) {
	//关闭amqp链接通道
	c.nc.Close()
	c.timeout_done <- nil
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
func (c *NatsClient) Call(callInfo mqrpc.CallInfo, callback chan rpcpb.ResultInfo) error {
	//var err error
	if c.callinfos == nil {
		return fmt.Errorf("AMQPClient is closed")
	}
	callInfo.RpcInfo.ReplyTo = c.callbackqueueName
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
	return c.nc.Publish(c.conf.Id,body)
}

/**
消息请求 不需要回复
*/
func (c *NatsClient) CallNR(callInfo mqrpc.CallInfo) error {
	body, err := c.Marshal(&callInfo.RpcInfo)
	if err != nil {
		return err
	}
	return c.nc.Publish(c.conf.Id,body)
}

func (c *NatsClient) on_timeout_handle(done chan error) {
	timeout := time.NewTimer(time.Second * 1)
	for {
		select {
		case <-timeout.C:
			timeout.Reset(time.Second * 1)
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
		case <-done:
			timeout.Stop()
			goto LLForEnd

		}
	}
	LLForEnd:
}

/**
接收应答信息
*/
func (c *NatsClient) on_response_handle(m *nats.Msg) {
	defer func() {
		if r := recover(); r != nil {
			var rn = ""
			switch r.(type) {

			case string:
				rn = r.(string)
			case error:
				rn = r.(error).Error()
			}
			buf := make([]byte, 1024)
			l := runtime.Stack(buf, false)
			errstr := string(buf[:l])
			log.Error("%s\n ----Stack----\n%s", rn, errstr)
		}
	}()
	resultInfo, err := c.UnmarshalResult(m.Data)
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

func (c *NatsClient) UnmarshalResult(data []byte) (*rpcpb.ResultInfo, error) {
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

func (c *NatsClient) Unmarshal(data []byte) (*rpcpb.RPCInfo, error) {
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
func (c *NatsClient) Marshal(rpcInfo *rpcpb.RPCInfo) ([]byte, error) {
	//map2:= structs.Map(callInfo)
	b, err := proto.Marshal(rpcInfo)
	return b, err
}

