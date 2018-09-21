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
	"github.com/pkg/errors"
	"net"
	"sync"
	"time"
)

type UDPClient struct {
	//callinfos map[string]*ClinetCallInfo
	callinfos    *utils.BeeMap
	cmutex       sync.Mutex //操作callinfos的锁
	uri          string
	info         *conf.UDP
	conn         *net.UDPConn
	data         []byte
	done         chan error
	isClose      bool
	timeout_done chan error
}

func NewUDPClient(info *conf.UDP) (*UDPClient, error) {
	client := new(UDPClient)
	//client.callinfos=make(map[string]*ClinetCallInfo)
	client.callinfos = utils.NewBeeMap()
	client.info = info
	client.uri = info.Uri
	client.done = make(chan error)
	client.timeout_done = make(chan error)
	client.isClose = false
	client.data = make([]byte, info.UDPMaxPacketSize)
	go client.on_response_handle(client.done)
	go client.on_timeout_handle(client.timeout_done) //处理超时请求的协程
	return client, nil
}
func (this *UDPClient) GetConn() (*net.UDPConn, error) {
	if this.conn == nil {
		// 创建连接
		addr, err := net.ResolveUDPAddr("udp4", this.uri)
		if err != nil {
			return nil, err
		}
		socket, err := net.DialUDP("udp4", nil, addr)
		if err != nil {
			return nil, err
		}
		this.conn = socket
	}

	return this.conn, nil
}
func (c *UDPClient) Done() error {
	//关闭消息回复通道
	c.isClose = true
	c.timeout_done <- nil
	//c.done <- nil
	//清理 callinfos 列表
	for key, clinetCallInfo := range c.callinfos.Items() {
		if clinetCallInfo != nil {
			//关闭管道
			close(clinetCallInfo.(ClinetCallInfo).call)
			//从Map中删除
			c.callinfos.Delete(key)
		}
	}
	return nil
}

/**
消息请求
*/
func (this *UDPClient) Call(callInfo mqrpc.CallInfo, callback chan rpcpb.ResultInfo) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf(r.(string))
		}
	}()

	if this.isClose {
		return fmt.Errorf("MQClient is closed")
	}
	var correlation_id = callInfo.RpcInfo.Cid

	clinetCallInfo := &ClinetCallInfo{
		correlation_id: correlation_id,
		call:           callback,
		timeout:        callInfo.RpcInfo.Expired,
	}
	this.callinfos.Set(correlation_id, *clinetCallInfo)
	callInfo.Props = map[string]interface{}{
	//"reply_to": this.result_chan,
	}

	body, err := this.Marshal(&callInfo.RpcInfo)
	if err != nil {
		return err
	}
	if this.info.UDPMaxPacketSize < len(body) {
		return errors.New(fmt.Sprintf("PacketSize(%d) is greater than UDP maximum packet(%d) now", len(body), this.info.UDPMaxPacketSize))
	}
	//发送消息
	conn, err := this.GetConn()
	if err != nil {
		return err
	}

	count, err := conn.Write(body)
	if count != len(body) {
		log.Warning("UDP RPC Server: Sending data length is wrong send(%d) body(%d)", count, len(body))
	}

	return nil
}

/**
消息请求 不需要回复
*/
func (this *UDPClient) CallNR(callInfo mqrpc.CallInfo) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf(r.(string))
		}
	}()

	body, err := this.Marshal(&callInfo.RpcInfo)
	if err != nil {
		return err
	}
	if this.info.UDPMaxPacketSize < len(body) {
		return errors.New(fmt.Sprintf("PacketSize(%d) is greater than UDP maximum packet(%d) now", len(body), this.info.UDPMaxPacketSize))
	}
	//发送消息
	conn, err := this.GetConn()
	if err != nil {
		return err
	}

	count, err := conn.Write(body)
	if count != len(body) {
		log.Warning("UDP RPC Server: Sending data length is wrong send(%d) body(%d)", count, len(body))
	}

	return nil
}

func (c *UDPClient) on_timeout_handle(done chan error) {
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
func (this *UDPClient) on_response_handle(done chan error) {
	for !this.isClose {
		conn, err := this.GetConn()
		if err != nil {
			continue
		}
		read, _, err := conn.ReadFromUDP(this.data)
		if err != nil {
			log.Warning("UDP RPC Server: ReadFromUDP ERROR %s", err.Error())
			continue
		}
		result := this.data[:read]

		resultInfo, err := this.UnmarshalResult(result)
		if err != nil {
			log.Error("Unmarshal faild", err)
		} else {
			correlation_id := resultInfo.Cid
			clinetCallInfo := this.callinfos.Get(correlation_id)
			//删除
			this.callinfos.Delete(correlation_id)
			if clinetCallInfo != nil {
				clinetCallInfo.(ClinetCallInfo).call <- *resultInfo
				close(clinetCallInfo.(ClinetCallInfo).call)
			} else {
				//可能客户端已超时了，但服务端处理完还给回调了
				log.Warning("rpc callback no found : [%s]", correlation_id)
			}
		}
	}
	log.Debug("finish on_response_handle")
}

func (c *UDPClient) UnmarshalResult(data []byte) (*rpcpb.ResultInfo, error) {
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

func (c *UDPClient) Unmarshal(data []byte) (*rpcpb.RPCInfo, error) {
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
func (c *UDPClient) Marshal(rpcInfo *rpcpb.RPCInfo) ([]byte, error) {
	//map2:= structs.Map(callInfo)
	b, err := proto.Marshal(rpcInfo)
	return b, err
}
