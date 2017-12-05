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

import ()
import (
	"fmt"
	"github.com/liangdas/mqant/log"
	"github.com/liangdas/mqant/rpc"
	"github.com/liangdas/mqant/rpc/pb"
	"sync"
)

type LocalServer struct {
	call_chan  chan mqrpc.CallInfo
	local_chan chan mqrpc.CallInfo
	done       chan error
	isclose    bool
	lock       *sync.Mutex
}

func NewLocalServer(call_chan chan mqrpc.CallInfo) (*LocalServer, error) {
	server := new(LocalServer)
	server.call_chan = call_chan
	server.local_chan = make(chan mqrpc.CallInfo, 1)
	server.isclose = false
	server.lock = new(sync.Mutex)
	go server.on_request_handle(server.local_chan)
	return server, nil
}

/**
停止接收请求
*/
func (s *LocalServer) IsClose() bool {
	return s.isclose
}

/**
停止接收请求
*/
func (s *LocalServer) Write(callInfo mqrpc.CallInfo) error {
	if s.isclose {
		return fmt.Errorf("LocalServer is closed")
	}
	s.local_chan <- callInfo
	return nil
}

/**
停止接收请求
*/
func (s *LocalServer) StopConsume() error {
	s.lock.Lock()
	s.isclose = true
	s.lock.Unlock()
	return nil
}

/**
注销消息队列
*/
func (s *LocalServer) Shutdown() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf(r.(string))
		}
	}()
	return nil
}

/**

 */
func (s *LocalServer) Callback(callinfo mqrpc.CallInfo) error {
	reply_to := callinfo.Props["reply_to"].(chan rpcpb.ResultInfo)
	reply_to <- callinfo.Result
	//select {
	//case reply_to<-callinfo.Result:
	//case <- time.After(time.Second *1):
	//	return fmt.Errorf("timeout : [%s]",callinfo.RpcInfo.Cid)
	//}
	//if s.SafeCallback(reply_to,callinfo.Result){
	//	log.Warning("rpc callback fail : [%s]",callinfo.RpcInfo.Cid)
	//}
	return nil
}

/**
接收请求信息
*/
func (s *LocalServer) on_request_handle(local_chan <-chan mqrpc.CallInfo) {
	for {
		select {
		case callInfo, ok := <-local_chan:
			if !ok {
				local_chan = nil
			} else {
				callInfo.Agent = s //设置代理为LocalServer
				if s.SafeSend(s.call_chan, callInfo) {
					log.Warning("rpc request fail : [%s]", callInfo.RpcInfo.Cid)
				}
			}
		}
		if local_chan == nil {
			break
		}
	}
}
func (s *LocalServer) SafeCallback(local_chan chan rpcpb.ResultInfo, callInfo rpcpb.ResultInfo) (closed bool) {
	defer func() {
		if recover() != nil {
			closed = true
		}
	}()

	// assume ch != nil here.
	local_chan <- callInfo
	return false
}
func (s *LocalServer) SafeSend(local_chan chan mqrpc.CallInfo, callInfo mqrpc.CallInfo) (closed bool) {
	defer func() {
		if recover() != nil {
			closed = true
		}
	}()

	// assume ch != nil here.
	local_chan <- callInfo
	return false
}
