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
	"github.com/liangdas/mqant/gate"
	"github.com/liangdas/mqant/log"
	"github.com/liangdas/mqant/module"
	"github.com/liangdas/mqant/rpc"
	"github.com/liangdas/mqant/rpc/pb"
	"github.com/liangdas/mqant/rpc/util"
	"reflect"
	"runtime"
	"sync"
	"time"
	"encoding/json"
)

type RPCServer struct {
	module         module.Module
	app            module.App
	functions      map[string]*mqrpc.FunctionInfo
	nats_server    *NatsServer
	mq_chan        chan mqrpc.CallInfo //接收到请求信息的队列
	wg             sync.WaitGroup      //任务阻塞
	call_chan_done chan error
	listener       mqrpc.RPCListener
	control        mqrpc.GoroutineControl //控制模块可同时开启的最大协程数
	executing      int64                  //正在执行的goroutine数量
	ch             chan int               //控制模块可同时开启的最大协程数
}

func NewRPCServer(app module.App, module module.Module) (mqrpc.RPCServer, error) {
	rpc_server := new(RPCServer)
	rpc_server.app = app
	rpc_server.module = module
	rpc_server.call_chan_done = make(chan error)
	rpc_server.functions = make(map[string]*mqrpc.FunctionInfo)
	rpc_server.mq_chan = make(chan mqrpc.CallInfo)
	rpc_server.ch = make(chan int, app.GetSettings().Rpc.MaxCoroutine)
	rpc_server.SetGoroutineControl(rpc_server)

	nats_server, err := NewNatsServer(app, rpc_server)
	if err != nil {
		log.Error("AMQPServer Dial: %s", err)
	}
	rpc_server.nats_server = nats_server

	//go rpc_server.on_call_handle(rpc_server.mq_chan, rpc_server.call_chan_done)

	return rpc_server, nil
}

func (this *RPCServer) Addr() string {
	return this.nats_server.Addr()
}

func (this *RPCServer) Wait() error {
	// 如果ch满了则会处于阻塞，从而达到限制最大协程的功能
	this.ch <- 1
	return nil
}
func (this *RPCServer) Finish() {
	// 完成则从ch推出数据
	<-this.ch
}

func (s *RPCServer) SetListener(listener mqrpc.RPCListener) {
	s.listener = listener
}
func (s *RPCServer) SetGoroutineControl(control mqrpc.GoroutineControl) {
	s.control = control
}

/**
获取当前正在执行的goroutine 数量
*/
func (s *RPCServer) GetExecuting() int64 {
	return s.executing
}

// you must call the function before calling Open and Go
func (s *RPCServer) Register(id string, f interface{}) {

	if _, ok := s.functions[id]; ok {
		panic(fmt.Sprintf("function id %v: already registered", id))
	}

	s.functions[id] = &mqrpc.FunctionInfo{
		Function:  reflect.ValueOf(f),
		Goroutine: false,
	}
}

// you must call the function before calling Open and Go
func (s *RPCServer) RegisterGO(id string, f interface{}) {

	if _, ok := s.functions[id]; ok {
		panic(fmt.Sprintf("function id %v: already registered", id))
	}

	s.functions[id] = &mqrpc.FunctionInfo{
		Function:  reflect.ValueOf(f),
		Goroutine: true,
	}
}

func (s *RPCServer) Done() (err error) {
	//等待正在执行的请求完成
	//close(s.mq_chan)   //关闭mq_chan通道
	//<-s.call_chan_done //mq_chan通道的信息都已处理完
	s.wg.Wait()
	//s.call_chan_done <- nil
	//关闭队列链接
	if s.nats_server != nil {
		err = s.nats_server.Shutdown()
	}
	return
}

func (s *RPCServer) Call(callInfo mqrpc.CallInfo) error {
	s.runFunc(callInfo)
	//if callInfo.RpcInfo.Expired < (time.Now().UnixNano() / 1000000) {
	//	//请求超时了,无需再处理
	//	if s.listener != nil {
	//		s.listener.OnTimeOut(callInfo.RpcInfo.Fn, callInfo.RpcInfo.Expired)
	//	} else {
	//		log.Warning("timeout: This is Call", s.module.GetType(), callInfo.RpcInfo.Fn, callInfo.RpcInfo.Expired, time.Now().UnixNano()/1000000)
	//	}
	//} else {
	//	s.runFunc(callInfo)
	//	//go func() {
	//	//	resultInfo := rpcpb.NewResultInfo(callInfo.RpcInfo.Cid, "", argsutil.STRING, []byte("success"))
	//	//	callInfo.Result = *resultInfo
	//	//	s.doCallback(callInfo)
	//	//}()
	//
	//}
	return nil
}

/**
接收请求信息
*/
func (s *RPCServer) on_call_handle(calls <-chan mqrpc.CallInfo, done chan error) {
	for {
		select {
		case callInfo, ok := <-calls:
			if !ok {
				goto ForEnd
			} else {
				if callInfo.RpcInfo.Expired < (time.Now().UnixNano() / 1000000) {
					//请求超时了,无需再处理
					if s.listener != nil {
						s.listener.OnTimeOut(callInfo.RpcInfo.Fn, callInfo.RpcInfo.Expired)
					} else {
						log.Warning("timeout: This is Call", s.module.GetType(), callInfo.RpcInfo.Fn, callInfo.RpcInfo.Expired, time.Now().UnixNano()/1000000)
					}
				} else {
					s.runFunc(callInfo)
				}
			}
		case <-done:
			goto ForEnd
		}
	}
ForEnd:
}

func (s *RPCServer) doCallback(callInfo mqrpc.CallInfo) {
	if callInfo.RpcInfo.Reply {
		//需要回复的才回复
		err := callInfo.Agent.(mqrpc.MQServer).Callback(callInfo)
		if err != nil {
			log.Warning("rpc callback erro :\n%s", err.Error())
		}

		//if callInfo.RpcInfo.Expired < (time.Now().UnixNano() / 1000000) {
		//	//请求超时了,无需再处理
		//	err := callInfo.Agent.(mqrpc.MQServer).Callback(callInfo)
		//	if err != nil {
		//		log.Warning("rpc callback erro :\n%s", err.Error())
		//	}
		//}else {
		//	log.Warning("timeout: This is Call %s %s", s.module.GetType(), callInfo.RpcInfo.Fn)
		//}
	} else {
		//对于不需要回复的消息,可以判断一下是否出现错误，打印一些警告
		if callInfo.Result.Error != "" {
			log.Warning("rpc callback erro :\n%s", callInfo.Result.Error)
		}
	}
}

//---------------------------------if _func is not a function or para num and type not match,it will cause panic
func (s *RPCServer) runFunc(callInfo mqrpc.CallInfo) {
	start := time.Now()
	_errorCallback := func(Cid string, Error string, span log.TraceSpan) {
		//异常日志都应该打印
		//log.TError(span, "RPC Exec ModuleType = %v Func = %v Elapsed = %v ERROR:\n%v", s.module.GetType(), callInfo.RpcInfo.Fn, time.Since(start), Error)
		resultInfo := rpcpb.NewResultInfo(Cid, Error, argsutil.NULL, nil)
		callInfo.Result = *resultInfo
		s.doCallback(callInfo)
		if s.listener != nil {
			s.listener.OnError(callInfo.RpcInfo.Fn, &callInfo, fmt.Errorf(Error))
		}
	}
	defer func() {
		if r := recover(); r != nil {
			var rn = ""
			switch r.(type) {

			case string:
				rn = r.(string)
			case error:
				rn = r.(error).Error()
			}
			log.Error("recover", rn)
			_errorCallback(callInfo.RpcInfo.Cid, rn, nil)
		}
	}()

	functionInfo, ok := s.functions[callInfo.RpcInfo.Fn]
	if !ok {
		if s.listener != nil {
			fInfo, err := s.listener.NoFoundFunction(callInfo.RpcInfo.Fn)
			if err != nil {
				_errorCallback(callInfo.RpcInfo.Cid, err.Error(), nil)
				return
			}
			functionInfo = fInfo
		}
	}
	f := functionInfo.Function
	params := callInfo.RpcInfo.Args
	ArgsType := callInfo.RpcInfo.ArgsType
	if len(params) != f.Type().NumIn() {
		//因为在调研的 _func的时候还会额外传递一个回调函数 cb
		_errorCallback(callInfo.RpcInfo.Cid, fmt.Sprintf("The number of params %v is not adapted.%v", params, f.String()), nil)
		return
	}
	//if len(params) != len(callInfo.RpcInfo.ArgsType) {
	//	//因为在调研的 _func的时候还会额外传递一个回调函数 cb
	//	_errorCallback(callInfo.RpcInfo.Cid,fmt.Sprintf("The number of params %s is not adapted ArgsType .%s", params, callInfo.RpcInfo.ArgsType))
	//	return
	//}

	//typ := reflect.TypeOf(_func)

	_runFunc := func() {
		s.wg.Add(1)
		s.executing++
		var span log.TraceSpan = nil
		defer func() {
			s.wg.Add(-1)
			s.executing--
			if s.control != nil {
				s.control.Finish()
			}
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
				allError := fmt.Sprintf("%s rpc func(%s) error %s\n ----Stack----\n%s", s.module.GetType(), callInfo.RpcInfo.Fn, rn, errstr)
				log.Error(allError)
				_errorCallback(callInfo.RpcInfo.Cid, allError, span)
			}
		}()

		//t:=RandInt64(2,3)
		//time.Sleep(time.Second*time.Duration(t))
		// f 为函数地址
		var session gate.Session = nil
		var in []reflect.Value

		if len(ArgsType) > 0 {
			in = make([]reflect.Value, len(params))
			for k, v := range ArgsType {
				ty, err := argsutil.Bytes2Args(s.app, v, params[k])
				if err != nil {
					_errorCallback(callInfo.RpcInfo.Cid, err.Error(), span)
					return
				}
				switch v2 := ty.(type) { //多选语句switch
				case gate.Session:
					//尝试加载Span
					if v2 != nil {
						session = v2.Clone()
						span = session
					}
					in[k] = reflect.ValueOf(ty)
				case log.TraceSpan:
					//尝试加载Span
					if v2 != nil {
						span = v2.ExtractSpan()
					}
					in[k] = reflect.ValueOf(ty)
				case []uint8:
					if reflect.TypeOf(ty).AssignableTo(f.Type().In(k)){
						in[k] = reflect.ValueOf(ty)
					}else{
						elemp := reflect.New(f.Type().In(k))
						err:=json.Unmarshal(v2,elemp.Interface())
						if err!=nil{
							log.Error("%v []uint8--> %v error with='%v'",callInfo.RpcInfo.Fn,f.Type().In(k),err)
							in[k] = reflect.ValueOf(ty)
						}else{
							in[k] = elemp.Elem()
						}
					}
				case nil:
					in[k] = reflect.Zero(f.Type().In(k))
				default:
					in[k] = reflect.ValueOf(ty)
				}

			}
		}

		if s.listener != nil {
			errs := s.listener.BeforeHandle(callInfo.RpcInfo.Fn, &callInfo)
			if errs != nil {
				_errorCallback(callInfo.RpcInfo.Cid, errs.Error(), span)
				return
			}
		}

		out := f.Call(in)
		var rs []interface{}
		if len(out) != 2 {
			_errorCallback(callInfo.RpcInfo.Cid, "The number of prepare is not adapted.", span)
			return
		}
		if len(out) > 0 { //prepare out paras
			rs = make([]interface{}, len(out), len(out))
			for i, v := range out {
				rs[i] = v.Interface()
			}
		}
		argsType, args, err := argsutil.ArgsTypeAnd2Bytes(s.app, rs[0])
		if err != nil {
			_errorCallback(callInfo.RpcInfo.Cid, err.Error(), span)
			return
		}
		resultInfo := rpcpb.NewResultInfo(
			callInfo.RpcInfo.Cid,
			rs[1].(string),
			argsType,
			args,
		)
		callInfo.Result = *resultInfo
		s.doCallback(callInfo)
		if s.app.GetSettings().Rpc.Log {
			log.TInfo(span, "RPC Exec ModuleType = %v Func = %v Elapsed = %v", s.module.GetType(), callInfo.RpcInfo.Fn, time.Since(start))
		}
		if s.listener != nil {
			s.listener.OnComplete(callInfo.RpcInfo.Fn, &callInfo, resultInfo, time.Since(start).Nanoseconds())
		}
	}
	if s.control != nil {
		//协程数量达到最大限制
		s.control.Wait()
	}
	if functionInfo.Goroutine {
		go _runFunc()
	} else {
		_runFunc()
	}
}
