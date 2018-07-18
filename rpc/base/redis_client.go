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
	"github.com/gomodule/redigo/redis"
	"github.com/liangdas/mqant/conf"
	"github.com/liangdas/mqant/log"
	"github.com/liangdas/mqant/rpc"
	"github.com/liangdas/mqant/rpc/pb"
	"github.com/liangdas/mqant/rpc/util"
	"github.com/liangdas/mqant/utils"
	"sync"
	"time"
)

type RedisClient struct {
	//callinfos map[string]*ClinetCallInfo
	callinfos         utils.ConcurrentMap
	cmutex            sync.Mutex //操作callinfos的锁
	info              *conf.Redis
	queueName         string
	callbackqueueName string
	done              chan error
	timeout_done      chan error
	pool              redis.Conn
	closed            bool
	ch                chan int //控制一个RPC 客户端可以同时等待的最大消息量，
	// 如果等待的请求过多代表rpc server请求压力大，
	// 就不要在往消息队列发消息了,客户端先阻塞

}

func createQueueName() string {
	//return "callbackqueueName"
	return fmt.Sprintf("callbackqueueName:%d", time.Now().Nanosecond())
}
func NewRedisClient(info *conf.Redis) (client *RedisClient, err error) {
	client = new(RedisClient)
	client.callinfos = utils.New()
	client.info = info
	client.callbackqueueName = createQueueName()
	client.queueName = info.Queue
	client.done = make(chan error)
	client.timeout_done = make(chan error)
	client.closed = false
	client.ch = make(chan int, 100)
	pool := utils.GetRedisFactory().GetPool(info.Uri).Get()
	defer pool.Close()
	//_, errs:=pool.Do("EXPIRE", client.callbackqueueName, 60)
	//if errs != nil {
	//	log.Warning(errs.Error())
	//}
	go client.on_response_handle(client.done)
	go client.on_timeout_handle(client.timeout_done) //处理超时请求的协程
	return client, nil
	//log.Printf("shutting down")
	//
	//if err := c.Shutdown(); err != nil {
	//	log.Fatalf("error during shutdown: %s", err)
	//}
}
func (this *RedisClient) Wait() error {
	// 如果ch满了则会处于阻塞，从而达到限制最大协程的功能
	//this.ch <- 1
	return nil
}
func (this *RedisClient) Finish() {
	// 完成则从ch推出数据
	//<-this.ch
}
func (c *RedisClient) Done() (err error) {
	pool := utils.GetRedisFactory().GetPool(c.info.Uri).Get()
	defer pool.Close()
	//删除临时通道
	pool.Do("DEL", c.callbackqueueName)
	c.closed = true
	c.timeout_done <- nil
	//err = c.psc.Close()
	//清理 callinfos 列表
	if c.pool != nil {
		c.pool.Close()
	}
	for tuple := range c.callinfos.Iter() {
		//关闭管道
		close(tuple.Val.(*ClinetCallInfo).call)
		//从Map中删除
		c.callinfos.Remove(tuple.Key)
	}
	return
}

/**
消息请求
*/
func (c *RedisClient) Call(callInfo mqrpc.CallInfo, callback chan rpcpb.ResultInfo) error {
	pool := utils.GetRedisFactory().GetPool(c.info.Uri).Get()
	defer pool.Close()
	var err error
	if c.closed {
		return fmt.Errorf("RedisClient is closed")
	}
	callInfo.RpcInfo.ReplyTo = c.callbackqueueName
	var correlation_id = callInfo.RpcInfo.Cid

	clinetCallInfo := &ClinetCallInfo{
		correlation_id: correlation_id,
		call:           callback,
		timeout:        callInfo.RpcInfo.Expired,
	}

	body, err := c.Marshal(&callInfo.RpcInfo)
	if err != nil {
		return err
	}
	c.callinfos.Set(correlation_id, clinetCallInfo)
	c.Wait() //阻塞等待可以发下一条消息
	_, err = pool.Do("lpush", c.queueName, body)
	if err != nil {
		log.Warning("Publish: %s", err)
		return err
	}
	return nil
}

/**
消息请求 不需要回复
*/
func (c *RedisClient) CallNR(callInfo mqrpc.CallInfo) error {
	c.Wait() //阻塞等待可以发下一条消息
	pool := utils.GetRedisFactory().GetPool(c.info.Uri).Get()
	defer pool.Close()
	var err error

	body, err := c.Marshal(&callInfo.RpcInfo)
	if err != nil {
		return err
	}
	_, err = pool.Do("lpush", c.queueName, body)
	if err != nil {
		log.Warning("Publish: %s", err)
		return err
	}
	return nil
}

func (c *RedisClient) on_timeout_handle(done chan error) {
	timeout := time.NewTimer(time.Second * 1)
	for {
		select {
		case <-timeout.C:
			timeout.Reset(time.Second * 1)
			for tuple := range c.callinfos.Iter() {
				var clinetCallInfo = tuple.Val.(*ClinetCallInfo)
				if clinetCallInfo.timeout < (time.Now().UnixNano() / 1000000) {
					c.Finish() //完成一个请求
					//从Map中删除
					c.callinfos.Remove(tuple.Key)
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
func (c *RedisClient) on_response_handle(done chan error) {
	for !c.closed {
		c.pool = utils.GetRedisFactory().GetPool(c.info.Uri).Get()
		result, err := c.pool.Do("brpop", c.callbackqueueName, 0)
		c.pool.Close()
		if err == nil && result != nil {
			c.Finish() //完成一个请求
			resultInfo, err := c.UnmarshalResult(result.([]interface{})[1].([]byte))
			if err != nil {
				log.Error("Unmarshal faild", err)
			} else {
				correlation_id := resultInfo.Cid
				clinetCallInfo, ok := c.callinfos.Get(correlation_id)
				if ok {
					c.callinfos.Remove(correlation_id)
					clinetCallInfo.(*ClinetCallInfo).call <- *resultInfo
					close(clinetCallInfo.(*ClinetCallInfo).call)
				} else {
					//可能客户端已超时了，但服务端处理完还给回调了
					log.Warning("rpc callback no found : [%s]", correlation_id)
				}
			}
		} else if err != nil {
			log.Warning("error %s", err.Error())
		}
	}
	log.Debug("finish on_response_handle")
}

func (c *RedisClient) UnmarshalResult(data []byte) (*rpcpb.ResultInfo, error) {
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

func (c *RedisClient) Unmarshal(data []byte) (*rpcpb.RPCInfo, error) {
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
func (c *RedisClient) Marshal(rpcInfo *rpcpb.RPCInfo) ([]byte, error) {
	//map2:= structs.Map(callInfo)
	b, err := proto.Marshal(rpcInfo)
	return b, err
}
