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
	"google.golang.org/protobuf/proto"
	"github.com/liangdas/mqant/log"
	"github.com/liangdas/mqant/module"
	"github.com/liangdas/mqant/rpc"
	"github.com/liangdas/mqant/rpc/pb"
	"github.com/nats-io/nats.go"
	"runtime"
	"strings"
	"time"
)

type NatsServer struct {
	call_chan chan mqrpc.CallInfo
	addr      string
	app       module.App
	server    *RPCServer
	done      chan bool
	stopeds   chan bool
	subs      *nats.Subscription
	isClose   bool
}

func setAddrs(addrs []string) []string {
	var cAddrs []string
	for _, addr := range addrs {
		if len(addr) == 0 {
			continue
		}
		if !strings.HasPrefix(addr, "nats://") {
			addr = "nats://" + addr
		}
		cAddrs = append(cAddrs, addr)
	}
	if len(cAddrs) == 0 {
		cAddrs = []string{nats.DefaultURL}
	}
	return cAddrs
}

func NewNatsServer(app module.App, s *RPCServer) (*NatsServer, error) {
	server := new(NatsServer)
	server.server = s
	server.done = make(chan bool)
	server.stopeds = make(chan bool)
	server.isClose = false
	server.app = app
	server.addr = nats.NewInbox()
	go func() {
		server.on_request_handle()
		safeClose(server.stopeds)
	}()
	return server, nil
}
func (s *NatsServer) Addr() string {
	return s.addr
}

func safeClose(ch chan bool) {
	defer func() {
		if recover() != nil {
			// close(ch) panic occur
		}
	}()

	close(ch) // panic if ch is closed
}

/**
注销消息队列
*/
func (s *NatsServer) Shutdown() (err error) {
	safeClose(s.done)
	s.isClose = true
	select {
	case <-s.stopeds:
		//等待nats注销完成
	}
	return
}

func (s *NatsServer) Callback(callinfo *mqrpc.CallInfo) error {
	body, err := s.MarshalResult(callinfo.Result)
	if err != nil {
		return err
	}
	reply_to := callinfo.Props["reply_to"].(string)
	return s.app.Transport().Publish(reply_to, body)
}

/**
接收请求信息
*/
func (s *NatsServer) on_request_handle() (err error) {
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
			fmt.Println(errstr)
		}
	}()
	s.subs, err = s.app.Transport().SubscribeSync(s.addr)
	if err != nil {
		return err
	}

	go func() {
		select {
		case <-s.done:
			//服务关闭
		}
		s.subs.Unsubscribe()
	}()

	for !s.isClose {
		m, err := s.subs.NextMsg(time.Minute)
		if err != nil && err == nats.ErrTimeout {
			//fmt.Println(err.Error())
			//log.Warning("NatsServer error with '%v'",err)
			if !s.subs.IsValid() {
				//订阅已关闭，需要重新订阅
				s.subs, err = s.app.Transport().SubscribeSync(s.addr)
				if err != nil {
					log.Error("NatsServer SubscribeSync[1] error with '%v'", err)
					continue
				}
			}
			continue
		} else if err != nil {
			log.Warning("NatsServer error with '%v'", err)
			if !s.subs.IsValid() {
				//订阅已关闭，需要重新订阅
				s.subs, err = s.app.Transport().SubscribeSync(s.addr)
				if err != nil {
					log.Error("NatsServer SubscribeSync[2] error with '%v'", err)
					continue
				}
			}
			continue
		}

		rpcInfo, err := s.Unmarshal(m.Data)
		if err == nil {
			callInfo := &mqrpc.CallInfo{
				RPCInfo: rpcInfo,
			}
			callInfo.Props = map[string]interface{}{
				"reply_to": rpcInfo.ReplyTo,
			}

			callInfo.Agent = s //设置代理为NatsServer

			s.server.Call(callInfo)
		} else {
			fmt.Println("error ", err)
		}
	}
	return nil
}

func (s *NatsServer) Unmarshal(data []byte) (*rpcpb.RPCInfo, error) {
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
func (s *NatsServer) MarshalResult(resultInfo *rpcpb.ResultInfo) ([]byte, error) {
	//log.Error("",map2)
	b, err := proto.Marshal(resultInfo)
	return b, err
}
