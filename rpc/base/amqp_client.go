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
	"github.com/liangdas/mqant/conf"
	"github.com/liangdas/mqant/log"
	"github.com/liangdas/mqant/module/modules/timer"
	"github.com/golang/protobuf/proto"
	"github.com/liangdas/mqant/utils"
	"github.com/streadway/amqp"
	"sync"
	"time"
	"github.com/liangdas/mqant/rpc/pb"
	"github.com/liangdas/mqant/rpc"
	"github.com/liangdas/mqant/rpc/util"
)

type AMQPClient struct {
	//callinfos map[string]*ClinetCallInfo
	callinfos *utils.BeeMap
	cmutex    sync.Mutex //操作callinfos的锁
	Consumer  *Consumer
	done      chan error
}

type ClinetCallInfo struct {
	correlation_id string
	timeout        int64 //超时
	call           chan rpcpb.ResultInfo
}

func NewAMQPClient(info *conf.Rabbitmq) (client *AMQPClient, err error) {
	c, err := NewConsumer(info, info.Uri, info.Exchange, info.ExchangeType, info.ConsumerTag)
	if err != nil {
		log.Error("AMQPClient connect fail %s", err)
		return
	}
	// 声明回调队列，再次声明的原因是，服务器和客户端可能先后开启，该声明是幂等的，多次声明，但只生效一次
	queue, err := c.channel.QueueDeclare(
		"",    // name of the queue
		false, // durable	持久化
		true,  // delete when unused
		true,  // exclusive
		false, // noWait
		nil,   // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("Queue Declare: %s", err)
	}
	//log.Printf("declared Exchange, declaring Queue %q", queue.Name)
	c.callback_queue = queue.Name //设置callback_queue

	//log.Printf("Queue bound to Exchange, starting Consume (consumer tag %q)", c.tag)
	deliveries, err := c.channel.Consume(
		queue.Name, // name
		c.tag,      // consumerTag,
		false,      // noAck 自动应答
		false,      // exclusive
		false,      // noLocal
		false,      // noWait
		nil,        // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("Queue Consume: %s", err)
	}
	client = new(AMQPClient)
	client.callinfos = utils.NewBeeMap()
	client.Consumer = c
	client.done = make(chan error)
	go client.on_response_handle(deliveries, client.done)
	client.on_timeout_handle(nil) //处理超时请求的协程
	return client, nil
	//log.Printf("shutting down")
	//
	//if err := c.Shutdown(); err != nil {
	//	log.Fatalf("error during shutdown: %s", err)
	//}
}

func (c *AMQPClient) Done() (err error) {
	//关闭amqp链接通道
	err = c.Consumer.Shutdown()
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
	var err error
	if c.callinfos == nil {
		return fmt.Errorf("AMQPClient is closed")
	}
	callInfo.RpcInfo.ReplyTo=c.Consumer.callback_queue
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
	if err = c.Consumer.channel.Publish(
		c.Consumer.info.Exchange,   // publish to an exchange
		c.Consumer.info.BindingKey, // routing to 0 or more queues
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			Headers:         amqp.Table{"reply_to": c.Consumer.callback_queue},
			ContentType:     "text/plain",
			ContentEncoding: "",
			Body:            body,
			DeliveryMode:    amqp.Transient, // 1=non-persistent, 2=persistent
			Priority:        0,              // 0-9
			// a bunch of application/implementation-specific fields
		},
	); err != nil {
		log.Warning("Exchange Publish: %s", err)
		return err
	}
	return nil
}

/**
消息请求 不需要回复
*/
func (c *AMQPClient) CallNR(callInfo mqrpc.CallInfo) error {
	var err error

	body, err := c.Marshal(&callInfo.RpcInfo)
	if err != nil {
		return err
	}
	if err = c.Consumer.channel.Publish(
		c.Consumer.info.Exchange,   // publish to an exchange
		c.Consumer.info.BindingKey, // routing to 0 or more queues
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			Headers:         amqp.Table{"reply_to": c.Consumer.callback_queue},
			ContentType:     "text/plain",
			ContentEncoding: "",
			Body:            body,
			DeliveryMode:    amqp.Transient, // 1=non-persistent, 2=persistent
			Priority:        0,              // 0-9
			// a bunch of application/implementation-specific fields
		},
	); err != nil {
		log.Warning("Exchange Publish: %s", err)
		return err
	}
	return nil
}

func (c *AMQPClient) on_timeout_handle(args interface{}) {
	if c.callinfos != nil {
		//处理超时的请求
		for key, clinetCallInfo := range c.callinfos.Items() {
			if clinetCallInfo != nil {
				var clinetCallInfo = clinetCallInfo.(ClinetCallInfo)
				if clinetCallInfo.timeout < (time.Now().UnixNano() / 1000000) {
					//已经超时了
					resultInfo := &rpcpb.ResultInfo{
						Result: nil,
						Error:  "timeout: This is Call",
						ResultType:argsutil.NULL,
					}
					//发送一个超时的消息
					clinetCallInfo.call <- *resultInfo
					//关闭管道
					close(clinetCallInfo.call)
					//从Map中删除
					c.callinfos.Delete(key)
				}

			}
		}
		timer.SetTimer(1, c.on_timeout_handle, nil)
	}
}




/**
接收应答信息
*/
func (c *AMQPClient) on_response_handle(deliveries <-chan amqp.Delivery, done chan error) {
	for {
		select {
		case d, ok := <-deliveries:
			if !ok {
				deliveries = nil
			} else {
				//log.Printf(
				//	"got %dB on_response_handle delivery: [%v] %q",
				//	len(d.Body),
				//	d.DeliveryTag,
				//	d.Body,
				//)
				d.Ack(false)
				resultInfo,err := c.UnmarshalResult(d.Body)
				if err != nil {
					log.Error("Unmarshal faild", err)
				} else {
					correlation_id := resultInfo.Cid
					clinetCallInfo := c.callinfos.Get(correlation_id)
					if clinetCallInfo != nil {
						clinetCallInfo.(ClinetCallInfo).call <- *resultInfo
					}
					//删除
					c.callinfos.Delete(correlation_id)
				}
			}
		case <-done:
			c.Consumer.Shutdown()
			break
		}
		if deliveries == nil {
			break
		}
	}
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
