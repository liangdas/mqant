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
	"encoding/json"
	"fmt"
	"google.golang.org/protobuf/proto"
	"github.com/liangdas/mqant/log"
	"github.com/liangdas/mqant/module"
	"github.com/liangdas/mqant/rpc"
	"github.com/liangdas/mqant/rpc/pb"
	"github.com/liangdas/mqant/rpc/util"
	"reflect"
	"runtime"
	"sync"
	"time"
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
}

func NewRPCServer(app module.App, module module.Module) (mqrpc.RPCServer, error) {
	rpc_server := new(RPCServer)
	rpc_server.app = app
	rpc_server.module = module
	rpc_server.call_chan_done = make(chan error)
	rpc_server.functions = make(map[string]*mqrpc.FunctionInfo)
	rpc_server.mq_chan = make(chan mqrpc.CallInfo)

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
	finfo := &mqrpc.FunctionInfo{
		Function:  reflect.ValueOf(f),
		FuncType:  reflect.ValueOf(f).Type(),
		Goroutine: false,
	}

	finfo.InType = []reflect.Type{}
	for i := 0; i < finfo.FuncType.NumIn(); i++ {
		rv := finfo.FuncType.In(i)
		finfo.InType = append(finfo.InType, rv)
	}
	s.functions[id] = finfo

}

// you must call the function before calling Open and Go
func (s *RPCServer) RegisterGO(id string, f interface{}) {

	if _, ok := s.functions[id]; ok {
		panic(fmt.Sprintf("function id %v: already registered", id))
	}

	finfo := &mqrpc.FunctionInfo{
		Function:  reflect.ValueOf(f),
		FuncType:  reflect.ValueOf(f).Type(),
		Goroutine: true,
	}

	finfo.InType = []reflect.Type{}
	for i := 0; i < finfo.FuncType.NumIn(); i++ {
		rv := finfo.FuncType.In(i)
		finfo.InType = append(finfo.InType, rv)
	}
	s.functions[id] = finfo
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

func (s *RPCServer) Call(callInfo *mqrpc.CallInfo) error {
	s.runFunc(callInfo)
	//if callInfo.RPCInfo.Expired < (time.Now().UnixNano() / 1000000) {
	//	//请求超时了,无需再处理
	//	if s.listener != nil {
	//		s.listener.OnTimeOut(callInfo.RPCInfo.Fn, callInfo.RPCInfo.Expired)
	//	} else {
	//		log.Warning("timeout: This is Call", s.module.GetType(), callInfo.RPCInfo.Fn, callInfo.RPCInfo.Expired, time.Now().UnixNano()/1000000)
	//	}
	//} else {
	//	s.runFunc(callInfo)
	//	//go func() {
	//	//	resultInfo := rpcpb.NewResultInfo(callInfo.RPCInfo.Cid, "", argsutil.STRING, []byte("success"))
	//	//	callInfo.Result = *resultInfo
	//	//	s.doCallback(callInfo)
	//	//}()
	//
	//}
	return nil
}

func (s *RPCServer) doCallback(callInfo *mqrpc.CallInfo) {
	if callInfo.RPCInfo.Reply {
		//需要回复的才回复
		err := callInfo.Agent.(mqrpc.MQServer).Callback(callInfo)
		if err != nil {
			log.Warning("rpc callback erro :\n%s", err.Error())
		}

		//if callInfo.RPCInfo.Expired < (time.Now().UnixNano() / 1000000) {
		//	//请求超时了,无需再处理
		//	err := callInfo.Agent.(mqrpc.MQServer).Callback(callInfo)
		//	if err != nil {
		//		log.Warning("rpc callback erro :\n%s", err.Error())
		//	}
		//}else {
		//	log.Warning("timeout: This is Call %s %s", s.module.GetType(), callInfo.RPCInfo.Fn)
		//}
	} else {
		//对于不需要回复的消息,可以判断一下是否出现错误，打印一些警告
		if callInfo.Result.Error != "" {
			log.Warning("rpc callback erro :\n%s", callInfo.Result.Error)
		}
	}
	if s.app.Options().ServerRPCHandler != nil {
		s.app.Options().ServerRPCHandler(s.app, s.module, callInfo)
	}
}

func (s *RPCServer) _errorCallback(start time.Time, callInfo *mqrpc.CallInfo, Cid string, Error string) {
	//异常日志都应该打印
	//log.TError(span, "rpc Exec ModuleType = %v Func = %v Elapsed = %v ERROR:\n%v", s.module.GetType(), callInfo.RPCInfo.Fn, time.Since(start), Error)
	resultInfo := rpcpb.NewResultInfo(Cid, Error, argsutil.NULL, nil)
	callInfo.Result = resultInfo
	callInfo.ExecTime = time.Since(start).Nanoseconds()
	s.doCallback(callInfo)
	if s.listener != nil {
		s.listener.OnError(callInfo.RPCInfo.Fn, callInfo, fmt.Errorf(Error))
	}
}

func (s *RPCServer) _runFunc(start time.Time, functionInfo *mqrpc.FunctionInfo, callInfo *mqrpc.CallInfo) {
	f := functionInfo.Function
	fType := functionInfo.FuncType
	fInType := functionInfo.InType
	params := callInfo.RPCInfo.Args
	ArgsType := callInfo.RPCInfo.ArgsType
	if len(params) != fType.NumIn() {
		//因为在调研的 _func的时候还会额外传递一个回调函数 cb
		s._errorCallback(start, callInfo, callInfo.RPCInfo.Cid, fmt.Sprintf("The number of params %v is not adapted.%v", params, f.String()))
		return
	}

	s.wg.Add(1)
	s.executing++
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
			allError := fmt.Sprintf("%s rpc func(%s) error %s\n ----Stack----\n%s", s.module.GetType(), callInfo.RPCInfo.Fn, rn, errstr)
			log.Error(allError)
			s._errorCallback(start, callInfo, callInfo.RPCInfo.Cid, allError)
		}
	}()

	//t:=RandInt64(2,3)
	//time.Sleep(time.Second*time.Duration(t))
	// f 为函数地址
	var in []reflect.Value
	var input []interface{}
	if len(ArgsType) > 0 {
		in = make([]reflect.Value, len(params))
		input = make([]interface{}, len(params))
		for k, v := range ArgsType {
			rv := fInType[k]
			var elemp reflect.Value
			if rv.Kind() == reflect.Ptr {
				//如果是指针类型就得取到指针所代表的具体类型
				elemp = reflect.New(rv.Elem())
			} else {
				elemp = reflect.New(rv)
			}

			if pb, ok := elemp.Interface().(mqrpc.Marshaler); ok {
				err := pb.Unmarshal(params[k])
				if err != nil {
					s._errorCallback(start, callInfo, callInfo.RPCInfo.Cid, err.Error())
					return
				}
				if pb == nil { //多选语句switch
					in[k] = reflect.Zero(rv)
				}
				if rv.Kind() == reflect.Ptr {
					//接收指针变量的参数
					in[k] = reflect.ValueOf(elemp.Interface())
				} else {
					//接收值变量
					in[k] = elemp.Elem()
				}
				input[k] = pb
			} else if pb, ok := elemp.Interface().(proto.Message); ok {
				err := proto.Unmarshal(params[k], pb)
				if err != nil {
					s._errorCallback(start, callInfo, callInfo.RPCInfo.Cid, err.Error())
					return
				}
				if pb == nil { //多选语句switch
					in[k] = reflect.Zero(rv)
				}
				if rv.Kind() == reflect.Ptr {
					//接收指针变量的参数
					in[k] = reflect.ValueOf(elemp.Interface())
				} else {
					//接收值变量
					in[k] = elemp.Elem()
				}
				input[k] = pb
			} else {
				//不是Marshaler 才尝试用 argsutil 解析
				ty, err := argsutil.Bytes2Args(s.app, v, params[k])
				if err != nil {
					s._errorCallback(start, callInfo, callInfo.RPCInfo.Cid, err.Error())
					return
				}
				switch v2 := ty.(type) { //多选语句switch
				case nil:
					in[k] = reflect.Zero(rv)
				case []uint8:
					if reflect.TypeOf(ty).AssignableTo(rv) {
						//如果ty "继承" 于接受参数类型
						in[k] = reflect.ValueOf(ty)
					} else {
						elemp := reflect.New(rv)
						err := json.Unmarshal(v2, elemp.Interface())
						if err != nil {
							log.Error("%v []uint8--> %v error with='%v'", callInfo.RPCInfo.Fn, rv, err)
							in[k] = reflect.ValueOf(ty)
						} else {
							in[k] = elemp.Elem()
						}
					}
				default:
					in[k] = reflect.ValueOf(ty)
				}
				input[k] = in[k].Interface()
			}
		}
	}

	if s.listener != nil {
		errs := s.listener.BeforeHandle(callInfo.RPCInfo.Fn, callInfo)
		if errs != nil {
			s._errorCallback(start, callInfo, callInfo.RPCInfo.Cid, errs.Error())
			return
		}
	}

	out := f.Call(in)
	var rs []interface{}
	if len(out) != 2 {
		s._errorCallback(start, callInfo, callInfo.RPCInfo.Cid, fmt.Sprintf("%s rpc func(%s) return error %s\n", s.module.GetType(), callInfo.RPCInfo.Fn, "func(....)(result interface{}, err error)"))
		return
	}
	if len(out) > 0 { //prepare out paras
		rs = make([]interface{}, len(out), len(out))
		for i, v := range out {
			rs[i] = v.Interface()
		}
	}
	if s.app.Options().RpcCompleteHandler != nil {
		s.app.Options().RpcCompleteHandler(s.app, s.module, callInfo, input, rs, time.Since(start))
	}
	var rerr string
	switch e := rs[1].(type) {
	case string:
		rerr = e
		break
	case error:
		rerr = e.Error()
	case nil:
		rerr = ""
	default:
		s._errorCallback(start, callInfo, callInfo.RPCInfo.Cid, fmt.Sprintf("%s rpc func(%s) return error %s\n", s.module.GetType(), callInfo.RPCInfo.Fn, "func(....)(result interface{}, err error)"))
		return
	}
	argsType, args, err := argsutil.ArgsTypeAnd2Bytes(s.app, rs[0])
	if err != nil {
		s._errorCallback(start, callInfo, callInfo.RPCInfo.Cid, err.Error())
		return
	}
	resultInfo := rpcpb.NewResultInfo(
		callInfo.RPCInfo.Cid,
		rerr,
		argsType,
		args,
	)
	callInfo.Result = resultInfo
	callInfo.ExecTime = time.Since(start).Nanoseconds()
	s.doCallback(callInfo)
	if s.app.GetSettings().RPC.Log {
		log.TInfo(nil, "rpc Exec ModuleType = %v Func = %v Elapsed = %v", s.module.GetType(), callInfo.RPCInfo.Fn, time.Since(start))
	}
	if s.listener != nil {
		s.listener.OnComplete(callInfo.RPCInfo.Fn, callInfo, resultInfo, time.Since(start).Nanoseconds())
	}
}

//---------------------------------if _func is not a function or para num and type not match,it will cause panic
func (s *RPCServer) runFunc(callInfo *mqrpc.CallInfo) {
	start := time.Now()
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
			s._errorCallback(start, callInfo, callInfo.RPCInfo.Cid, rn)
		}
	}()

	if s.control != nil {
		//协程数量达到最大限制
		s.control.Wait()
	}
	functionInfo, ok := s.functions[callInfo.RPCInfo.Fn]
	if !ok {
		if s.listener != nil {
			fInfo, err := s.listener.NoFoundFunction(callInfo.RPCInfo.Fn)
			if err != nil {
				s._errorCallback(start, callInfo, callInfo.RPCInfo.Cid, err.Error())
				return
			}
			functionInfo = fInfo
		}
	}
	if functionInfo.Goroutine {
		go s._runFunc(start, functionInfo, callInfo)
	} else {
		s._runFunc(start, functionInfo, callInfo)
	}
}
