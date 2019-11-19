// Copyright 2014 mqantserver Author. All Rights Reserved.
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
package basemodule

import (
	"context"
	"encoding/json"
	"github.com/liangdas/mqant/conf"
	"github.com/liangdas/mqant/module"
	"github.com/liangdas/mqant/rpc"
	"github.com/liangdas/mqant/rpc/pb"
	"github.com/liangdas/mqant/selector"
	"github.com/liangdas/mqant/server"
	"github.com/liangdas/mqant/service"
	"github.com/liangdas/mqant/utils"
	"github.com/pkg/errors"
	"sync"
	"time"
)

type StatisticalMethod struct {
	Name        string //方法名
	StartTime   int64  //开始时间
	EndTime     int64  //结束时间
	MinExecTime int64  //最短执行时间
	MaxExecTime int64  //最长执行时间
	ExecTotal   int    //执行总次数
	ExecTimeout int    //执行超时次数
	ExecSuccess int    //执行成功次数
	ExecFailure int    //执行错误次数
}

func LoadStatisticalMethod(j string) map[string]*StatisticalMethod {
	sm := map[string]*StatisticalMethod{}
	err := json.Unmarshal([]byte(j), &sm)
	if err == nil {
		return sm
	} else {
		return nil
	}
}

type BaseModule struct {
	context.Context
	exit 		context.CancelFunc
	App         module.App
	subclass    module.RPCModule
	settings    *conf.ModuleSettings
	service     service.Service
	listener    mqrpc.RPCListener
	statistical map[string]*StatisticalMethod //统计
	rwmutex     sync.RWMutex
}

func (m *BaseModule) GetServerId() string {
	//很关键,需要与配置文件中的Module配置对应
	return m.service.Server().Id()
}

func (m *BaseModule) GetApp() module.App {
	return m.App
}

func (m *BaseModule) GetSubclass() module.RPCModule {
	return m.subclass
}

func (m *BaseModule) GetServer() server.Server {
	return m.service.Server()
}
func (m *BaseModule) OnConfChanged(settings *conf.ModuleSettings) {

}
func (m *BaseModule) OnAppConfigurationLoaded(app module.App) {
	m.App = app
	//当App初始化时调用，这个接口不管这个模块是否在这个进程运行都会调用
}
func (m *BaseModule) OnInit(subclass module.RPCModule, app module.App, settings *conf.ModuleSettings, opt ...server.Option) {
	//初始化模块
	m.App = app
	m.subclass = subclass
	m.settings = settings
	m.statistical = map[string]*StatisticalMethod{}
	//创建一个远程调用的RPC
	opts := server.Options{
		Metadata: map[string]string{},
	}

	for _, o := range opt {
		o(&opts)
	}

	if opts.Registry == nil {
		opt = append(opt, server.Registry(app.Registry()))
	}

	if opts.RegisterInterval == 0 {
		opt = append(opt, server.RegisterInterval(app.Options().RegisterInterval))
	}

	if opts.RegisterTTL == 0 {
		opt = append(opt, server.RegisterTTL(app.Options().RegisterTTL))
	}

	if len(opts.Name) == 0 {
		opt = append(opt, server.Name(subclass.GetType()))
	}

	if len(opts.Id) == 0 {
		opt = append(opt, server.Id(utils.GenerateID().String()))
	}

	if len(opts.Version) == 0 {
		opt = append(opt, server.Version(subclass.Version()))
	}

	server := server.NewServer(opt...)
	server.OnInit(subclass, app, settings)
	ctx,cancel:=context.WithCancel(context.Background())
	m.exit=cancel
	m.service = service.NewService(
		service.Server(server),
		service.RegisterTTL(app.Options().RegisterTTL),
		service.RegisterInterval(app.Options().RegisterInterval),
		service.Context(ctx),
	)

	go m.service.Run()
	m.GetServer().SetListener(m)
}

func (m *BaseModule) OnDestroy() {
	//注销模块
	//一定别忘了关闭RPC
	m.exit()
	m.GetServer().OnDestroy()
}
func (m *BaseModule) SetListener(listener mqrpc.RPCListener) {
	m.listener = listener
}
func (m *BaseModule) GetModuleSettings() *conf.ModuleSettings {
	return m.settings
}
func (m *BaseModule) GetRouteServer(moduleType string, opts ...selector.SelectOption) (s module.ServerSession, err error) {
	return m.App.GetRouteServer(moduleType, opts...)
}

func (m *BaseModule) RpcInvoke(moduleType string, _func string, params ...interface{}) (result interface{}, err string) {
	return m.App.RpcInvoke(m.GetSubclass(), moduleType, _func, params...)
}

func (m *BaseModule) RpcInvokeNR(moduleType string, _func string, params ...interface{}) (err error) {
	return m.App.RpcInvokeNR(m.GetSubclass(), moduleType, _func, params...)
}

func (m *BaseModule) RpcInvokeArgs(moduleType string, _func string, ArgsType []string, args [][]byte) (result interface{}, err string) {
	server, e := m.App.GetRouteServer(moduleType)
	if e != nil {
		err = e.Error()
		return
	}
	return server.CallArgs(nil, _func, ArgsType, args)
}

func (m *BaseModule) RpcInvokeNRArgs(moduleType string, _func string, ArgsType []string, args [][]byte) (err error) {
	server, err := m.App.GetRouteServer(moduleType)
	if err != nil {
		return
	}
	return server.CallNRArgs(_func, ArgsType, args)
}

func (m *BaseModule) RpcCall(ctx context.Context, moduleType, _func string, param mqrpc.ParamOption, opts ...selector.SelectOption) (interface{}, string) {
	return m.App.RpcCall(ctx, moduleType, _func, param, opts...)
}

func (m *BaseModule) NoFoundFunction(fn string) (*mqrpc.FunctionInfo, error) {
	if m.listener != nil {
		return m.listener.NoFoundFunction(fn)
	}
	return nil, errors.Errorf("Remote function(%s) not found", fn)
}

func (m *BaseModule) BeforeHandle(fn string, callInfo *mqrpc.CallInfo) error {
	if m.listener != nil {
		return m.listener.BeforeHandle(fn, callInfo)
	}
	return nil
}

func (m *BaseModule) OnTimeOut(fn string, Expired int64) {
	m.rwmutex.Lock()
	if statisticalMethod, ok := m.statistical[fn]; ok {
		statisticalMethod.ExecTimeout++
		statisticalMethod.ExecTotal++
	} else {
		statisticalMethod := &StatisticalMethod{
			Name:        fn,
			StartTime:   time.Now().UnixNano(),
			ExecTimeout: 1,
			ExecTotal:   1,
		}
		m.statistical[fn] = statisticalMethod
	}
	m.rwmutex.Unlock()
	if m.listener != nil {
		m.listener.OnTimeOut(fn, Expired)
	}
}
func (m *BaseModule) OnError(fn string, callInfo *mqrpc.CallInfo, err error) {
	m.rwmutex.Lock()
	if statisticalMethod, ok := m.statistical[fn]; ok {
		statisticalMethod.ExecFailure++
		statisticalMethod.ExecTotal++
	} else {
		statisticalMethod := &StatisticalMethod{
			Name:        fn,
			StartTime:   time.Now().UnixNano(),
			ExecFailure: 1,
			ExecTotal:   1,
		}
		m.statistical[fn] = statisticalMethod
	}
	m.rwmutex.Unlock()
	if m.listener != nil {
		m.listener.OnError(fn, callInfo, err)
	}
}

/**
fn 		方法名
params		参数
result		执行结果
exec_time 	方法执行时间 单位为 Nano 纳秒  1000000纳秒等于1毫秒
*/
func (m *BaseModule) OnComplete(fn string, callInfo *mqrpc.CallInfo, result *rpcpb.ResultInfo, exec_time int64) {
	m.rwmutex.Lock()
	if statisticalMethod, ok := m.statistical[fn]; ok {
		statisticalMethod.ExecSuccess++
		statisticalMethod.ExecTotal++
		if statisticalMethod.MinExecTime > exec_time {
			statisticalMethod.MinExecTime = exec_time
		}
		if statisticalMethod.MaxExecTime < exec_time {
			statisticalMethod.MaxExecTime = exec_time
		}
	} else {
		statisticalMethod := &StatisticalMethod{
			Name:        fn,
			StartTime:   time.Now().UnixNano(),
			ExecSuccess: 1,
			ExecTotal:   1,
			MaxExecTime: exec_time,
			MinExecTime: exec_time,
		}
		m.statistical[fn] = statisticalMethod
	}
	m.rwmutex.Unlock()
	if m.listener != nil {
		m.listener.OnComplete(fn, callInfo, result, exec_time)
	}
}
func (m *BaseModule) GetExecuting() int64 {
	return 0
	//return m.GetServer().GetRPCServer().GetExecuting()
}
func (m *BaseModule) GetStatistical() (statistical string, err error) {
	m.rwmutex.Lock()
	//重置
	now := time.Now().UnixNano()
	for _, s := range m.statistical {
		s.EndTime = now
	}
	b, err := json.Marshal(m.statistical)
	if err == nil {
		statistical = string(b)
	}

	//重置
	//for _,s:=range m.statistical{
	//	s.StartTime=now
	//	s.ExecFailure=0
	//	s.ExecSuccess=0
	//	s.ExecTimeout=0
	//	s.ExecTotal=0
	//	s.MaxExecTime=0
	//	s.MinExecTime=math.MaxInt64
	//}
	m.rwmutex.Unlock()
	return
}
