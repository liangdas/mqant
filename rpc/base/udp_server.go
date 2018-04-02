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
	"github.com/golang/protobuf/proto"
	"github.com/liangdas/mqant/conf"
	"github.com/liangdas/mqant/log"
	"github.com/liangdas/mqant/rpc"
	"github.com/liangdas/mqant/rpc/pb"
	"net"
	"runtime"
)

type UDPServer struct {
	call_chan chan mqrpc.CallInfo
	uri       string
	info      *conf.UDP
	conn      *net.UDPConn
	data      []byte
	done      chan error
	closed    bool
}

func NewUdpServer(info *conf.UDP, call_chan chan mqrpc.CallInfo) (*UDPServer, error) {
	server := new(UDPServer)
	server.call_chan = call_chan
	server.info = info
	server.done = make(chan error)
	server.closed = false
	server.data = make([]byte, info.UDPMaxPacketSize)
	//addr,err:=net.ResolveUDPAddr("udp4",server.uri)
	//if err != nil {
	//	return	err
	//}
	socket, err := net.ListenUDP("udp4", &net.UDPAddr{
		IP:   net.IPv4(0, 0, 0, 0),
		Port: info.Port,
	})
	if err != nil {
		return nil, err
	}
	server.conn = socket
	go server.on_request_handle(server.done)

	return server, nil
}

/**
停止接收请求
*/
func (this *UDPServer) StopConsume() error {
	this.closed = true
	if this.conn != nil {
		this.conn.Close()
		this.conn = nil
	}
	return nil
}

/**
注销消息队列
*/
func (this *UDPServer) Shutdown() error {
	this.closed = true
	if this.conn != nil {
		this.conn.Close()
		this.conn = nil
	}
	return nil
}

func (s *UDPServer) Callback(callinfo mqrpc.CallInfo) error {
	body, _ := s.MarshalResult(callinfo.Result)
	return s.response(callinfo.Props, body)
}

/**
消息应答
*/
func (this *UDPServer) response(props map[string]interface{}, body []byte) error {
	remoteAddr := props["reply_to"].(*net.UDPAddr)
	// 发送数据
	count, err := this.conn.WriteToUDP(body, remoteAddr)
	if err != nil {
		log.Warning("UDP RPC Server: %s", err.Error())
		return err
	}
	if count != len(body) {
		log.Warning("UDP RPC Server: Sending data length is wrong send(%d) body(%d)", count, len(body))
	}
	return nil
}

/**
接收请求信息
*/
func (this *UDPServer) on_request_handle(done chan error) {
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

	for !this.closed {
		// 读取数据
		read, remoteAddr, err := this.conn.ReadFromUDP(this.data)
		if err != nil {
			log.Warning("UDP RPC Server: ReadFromUDP ERROR %s", err.Error())
			continue
		}
		result := this.data[:read]
		if err == nil && result != nil {
			rpcInfo, err := this.Unmarshal(result)
			if err == nil {
				callInfo := &mqrpc.CallInfo{
					RpcInfo: *rpcInfo,
				}
				callInfo.Props = map[string]interface{}{
					"reply_to": remoteAddr,
				}

				callInfo.Agent = this //设置代理为AMQPServer

				this.call_chan <- *callInfo
			} else {
				log.Error("error ", err)
			}
		} else if err != nil {
			log.Warning("error %s", err.Error())
		}
	}
	log.Debug("finish on_request_handle")
}

func (s *UDPServer) Unmarshal(data []byte) (*rpcpb.RPCInfo, error) {
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
func (s *UDPServer) MarshalResult(resultInfo rpcpb.ResultInfo) ([]byte, error) {
	//log.Error("",map2)
	b, err := proto.Marshal(&resultInfo)
	return b, err
}
