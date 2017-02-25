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
package mqrpc

import (
	"github.com/liangdas/mqant/log"
	"fmt"
	"reflect"
	"time"
	"sync"
	"github.com/liangdas/mqant/conf"
)
type RPCListener interface {
	OnTimeOut(fn string ,Expired int64)
	OnError(fn string,params []interface{},err error)
	/**
	fn 		方法名
	params		参数
	result		执行结果
	exec_time 	方法执行时间 单位为 Nano 纳秒  1000000纳秒等于1毫秒
	 */
	OnComplete(fn string,params []interface{},result *ResultInfo,exec_time int64)
}


type CallInfo struct {
	Cid string	//Correlation_id
	Fn      string
	Args    []interface{}
	Expired int64	//超时
	Reply 	bool	//客户端是否需要结果
	Result 	ResultInfo
	props	map[string]interface{}
	agent	interface{} //代理者  AMQPServer / LocalServer 都继承 Callback(callinfo CallInfo)(error) 方法
}
type ResultInfo struct {
	Cid string	//Correlation_id
	Error   string  //错误结果 如果为nil表示请求正确
	Result  interface{}	//结果
}

type FunctionInfo struct{
	function 	interface{}
	goroutine	bool
}

type MQServer interface{
	Callback(callinfo CallInfo)(error)
}

type RPCServer struct{
	functions 		map[string]FunctionInfo
	remote_server 		*AMQPServer
	local_server 		*LocalServer
	mq_chan	  		chan CallInfo	//接收到请求信息的队列
	callback_chan 		chan CallInfo //信息处理完成的队列
	wg sync.WaitGroup	//任务阻塞
	call_chan_done  	chan error
	listener		RPCListener
	executing		int64		//正在执行的goroutine数量
}

func NewRPCServer() (*RPCServer,error){
	rpc_server:=new(RPCServer)
	rpc_server.call_chan_done = make(chan error)
	rpc_server.functions = make(map[string]FunctionInfo)
	rpc_server.mq_chan = make(chan CallInfo)
	rpc_server.callback_chan = make(chan CallInfo)

	//先创建一个本地的RPC服务
	local_server,err:=NewLocalServer(rpc_server.mq_chan)
	if err != nil {
		log.Error("LocalServer Dial: %s", err)
	}
	rpc_server.local_server = local_server

	go rpc_server.on_call_handle(rpc_server.mq_chan,rpc_server.callback_chan, rpc_server.call_chan_done)

	go rpc_server.on_callback_handle(rpc_server.callback_chan)	//结果发送队列
	return rpc_server,nil
}

/**
创建一个支持远程RPC的服务
 */
func (s *RPCServer) NewRemoteRPCServer(info *conf.Rabbitmq) (err error){
	remote_server,err:=NewAMQPServer(info,s.mq_chan)
	if err != nil {
		log.Error("AMQPServer Dial: %s", err)
	}
	s.remote_server = remote_server
	return
}

func (s *RPCServer) SetListener(listener RPCListener){
	s.listener=listener
}
/**
获取当前正在执行的goroutine 数量
 */
func (s *RPCServer) GetExecuting()(int64){
	return s.executing
}

// you must call the function before calling Open and Go
func (s *RPCServer) Register(id string, f interface{}) {

	if _, ok := s.functions[id]; ok {
		panic(fmt.Sprintf("function id %v: already registered", id))
	}

	s.functions[id] = *&FunctionInfo{
		function:f,
		goroutine:false,
	}
}
// you must call the function before calling Open and Go
func (s *RPCServer) RegisterGO(id string, f interface{}) {

	if _, ok := s.functions[id]; ok {
		panic(fmt.Sprintf("function id %v: already registered", id))
	}

	s.functions[id] = *&FunctionInfo{
		function:f,
		goroutine:true,
	}
}

func (s *RPCServer) Done()(err error) {
	//设置队列停止接收请求
	if s.remote_server!=nil{
		err=s.remote_server.StopConsume()
	}
	if s.local_server!=nil{
		err=s.local_server.StopConsume()
	}
	//等待正在执行的请求完成
	close(s.mq_chan) //关闭mq_chan通道
	<-s.call_chan_done //mq_chan通道的信息都已处理完
	s.wg.Wait()
	close(s.callback_chan) //关闭结果发送队列
	//关闭队列链接
	if s.remote_server!=nil{
		err=s.remote_server.Shutdown()
	}
	if s.local_server!=nil{
		err=s.local_server.Shutdown()
	}
	return
}

/**
处理结果信息
 */
func (s *RPCServer)on_callback_handle(callbacks <-chan CallInfo) {
	for {
		select {
		case callInfo, ok := <-callbacks:
			if !ok {
				callbacks = nil
			} else {
				if callInfo.Reply{
					//需要回复的才回复
					callInfo.agent.(MQServer).Callback(callInfo)
				}
			}
		}
		if callbacks == nil {
			break
		}
	}
}

/**
接收请求信息
 */
func (s *RPCServer)on_call_handle(calls <-chan CallInfo,callbacks chan <-CallInfo, done chan error) {
	for {
		select {
		case callInfo, ok := <-calls:
			if !ok {
				calls = nil
			} else {
				if callInfo.Expired<(time.Now().UnixNano()/ 1000000){
					//请求超时了,无需再处理
					if s.listener!=nil{
						s.listener.OnTimeOut(callInfo.Fn,callInfo.Expired)
					}else{
						fmt.Println("timeout: This is Call",callInfo.Fn,callInfo.Expired,time.Now().UnixNano()/ 1000000)
					}
				}else{
					s.runFunc(callInfo,callbacks)
				}
			}
		}
		if calls == nil {
			done<-nil
			break
		}
	}
}

//---------------------------------if _func is not a function or para num and type not match,it will cause panic
func (s *RPCServer)runFunc(callInfo CallInfo,callbacks  chan<- CallInfo)  {
	defer func() {
		if r := recover(); r != nil {
			var lerr error;
			var rn=""
			switch r.(type){

			case string:
				rn=r.(string)
				lerr=fmt.Errorf(rn)
			case error:
				rn=r.(error).Error()
				lerr=r.(error)
			}
			resultInfo:=&ResultInfo{
				Cid:callInfo.Cid,
				Error:rn,
				Result:nil,
			}
			callInfo.Result=*resultInfo
			callbacks<-callInfo

			if s.listener!=nil{
				s.listener.OnError(callInfo.Fn,callInfo.Args,lerr)
			}
		}
	}()

	functionInfo,ok := s.functions[callInfo.Fn]
	if !ok {
		resultInfo:=&ResultInfo{
			Cid:callInfo.Cid,
			Error:fmt.Sprintf("Remote function(%s) not found",callInfo.Fn) ,
			Result:nil,
		}
		callInfo.Result=*resultInfo
		callbacks<-callInfo
		if s.listener!=nil{
			s.listener.OnError(callInfo.Fn,callInfo.Args,fmt.Errorf("function not found"))
		}
		return
	}
	_func:=functionInfo.function
	params:=callInfo.Args
	f := reflect.ValueOf(_func)
	if len(params) != f.Type().NumIn() {
		//因为在调研的 _func的时候还会额外传递一个回调函数 cb
		resultInfo:=&ResultInfo{
			Cid:callInfo.Cid,
			Error:fmt.Sprintf("The number of params %s is not adapted.%s",params, f.String()) ,
			Result:nil,
		}
		callInfo.Result=*resultInfo
		callbacks<-callInfo
		if s.listener!=nil{
			s.listener.OnError(callInfo.Fn,callInfo.Args,fmt.Errorf("The number of params is not adapted."))
		}
		return
	}

	typ := reflect.TypeOf(_func)
	var in []reflect.Value
	if len(params) > 0 { //prepare in paras
		in = make([]reflect.Value, len(params))
		for k, param := range params {
			field := typ.In(k)
			if(field!=reflect.TypeOf(param)){
				switch param.(type) {      //多选语句switch
				case int:
					p:=param.(int)
					switch field.Kind().String() {      //多选语句switch
					case "int":
						in[k]=reflect.ValueOf(int(p))
					case "int32":
						in[k]=reflect.ValueOf(int32(p))
					case "int64":
						in[k]=reflect.ValueOf(int64(p))
					case "float32":
						in[k]=reflect.ValueOf(float32(p))
					case "float64":
						in[k]=reflect.ValueOf(float64(p))
					}
				case int32:
					p:=param.(int32)
					switch field.Kind().String() {      //多选语句switch
					case "int":
						in[k]=reflect.ValueOf(int(p))
					case "int32":
						in[k]=reflect.ValueOf(int32(p))
					case "int64":
						in[k]=reflect.ValueOf(int64(p))
					case "float32":
						in[k]=reflect.ValueOf(float32(p))
					case "float64":
						in[k]=reflect.ValueOf(float64(p))
					}
				case int64:
					p:=param.(int64)
					switch field.Kind().String() {      //多选语句switch
					case "int":
						in[k]=reflect.ValueOf(int(p))
					case "int32":
						in[k]=reflect.ValueOf(int32(p))
					case "int64":
						in[k]=reflect.ValueOf(int64(p))
					case "float32":
						in[k]=reflect.ValueOf(float32(p))
					case "float64":
						in[k]=reflect.ValueOf(float64(p))
					}
				case float32:
					p:=param.(float32)
					switch field.Kind().String() {      //多选语句switch
					case "int":
						in[k]=reflect.ValueOf(int(p))
					case "int32":
						in[k]=reflect.ValueOf(int32(p))
					case "int64":
						in[k]=reflect.ValueOf(int64(p))
					case "float32":
						in[k]=reflect.ValueOf(float32(p))
					case "float64":
						in[k]=reflect.ValueOf(float64(p))
					}
				case float64:
					p:=param.(float64)
					switch field.Kind().String() {      //多选语句switch
					case "int":
						in[k]=reflect.ValueOf(int(p))
					case "int32":
						in[k]=reflect.ValueOf(int32(p))
					case "int64":
						in[k]=reflect.ValueOf(int64(p))
					case "float32":
						in[k]=reflect.ValueOf(float32(p))
					case "float64":
						in[k]=reflect.ValueOf(float64(p))
					}
				}

			}else{
				in[k] = reflect.ValueOf(param)
			}
		}
	}
	s.wg.Add(1)
	s.executing++
	_runFunc:=func(){
		defer func() {
			if r := recover(); r != nil {
				var lerr error;
				var rn=""
				switch r.(type){

				case string:
					rn=r.(string)
					lerr=fmt.Errorf(rn)
				case error:
					rn=r.(error).Error()
					lerr=r.(error)
				}
				resultInfo:=&ResultInfo{
					Cid:callInfo.Cid,
					Error:rn,
					Result:nil,
				}
				callInfo.Result=*resultInfo
				callbacks<-callInfo
				if s.listener!=nil{
					s.listener.OnError(callInfo.Fn,callInfo.Args,lerr)
				}
			}
			s.wg.Add(-1)
			s.executing--
		}()
		exec_time :=time.Now().UnixNano()
		//t:=RandInt64(2,3)
		//time.Sleep(time.Second*time.Duration(t))
		// f 为函数地址
		out := f.Call(in)
		var rs []interface{}
		if len(out)!=2{
			resultInfo:=&ResultInfo{
				Cid:callInfo.Cid,
				Error:"The number of prepare is not adapted.",
				Result:nil,
			}
			callInfo.Result=*resultInfo
			callbacks<-callInfo
			return
		}
		if len(out) > 0 { //prepare out paras
			rs = make([]interface{}, len(out), len(out))
			for i, v := range out {
				rs[i] = v.Interface()
			}
		}
		resultInfo:=&ResultInfo{
			Cid:callInfo.Cid,
			Error:rs[1].(string),
			Result:rs[0],
		}
		callInfo.Result=*resultInfo
		callbacks<-callInfo
		if s.listener!=nil{
			s.listener.OnComplete(callInfo.Fn,callInfo.Args,resultInfo,time.Now().UnixNano()-exec_time)
		}
	}

	if functionInfo.goroutine{
		go _runFunc()
	}else{
		_runFunc()
	}
}



