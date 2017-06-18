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
	"encoding/json"
	"github.com/liangdas/mqant/conf"
	"github.com/liangdas/mqant/rpc"
	"sync"
	"time"
	"github.com/liangdas/mqant/rpc/pb"
	"github.com/liangdas/mqant/module"
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
	App         module.App
	subclass    module.RPCModule
	settings    *conf.ModuleSettings
	server      *rpcserver
	listener    mqrpc.RPCListener
	statistical map[string]*StatisticalMethod //统计
	rwmutex     sync.RWMutex
}

func (m *BaseModule) GetServerId() string {
	//很关键,需要与配置文件中的Module配置对应
	return m.settings.Id
}

func (m *BaseModule) GetApp() module.App {
	return m.App
}

func (m *BaseModule) GetServer() *rpcserver {
	if m.server == nil {
		m.server = new(rpcserver)
	}
	return m.server
}
func (m *BaseModule)OnConfChanged(settings *conf.ModuleSettings)  {

}
func (m *BaseModule) OnInit(subclass module.RPCModule, app module.App, settings *conf.ModuleSettings) {
	//初始化模块
	m.App = app
	m.subclass = subclass
	m.settings = settings
	m.statistical = map[string]*StatisticalMethod{}
	//创建一个远程调用的RPC
	m.GetServer().OnInit(subclass,app, settings)
	m.GetServer().GetRPCServer().SetListener(m)
}

func (m *BaseModule) OnDestroy() {
	//注销模块
	//一定别忘了关闭RPC
	m.GetServer().OnDestroy()
}
func (m *BaseModule) SetListener(listener mqrpc.RPCListener) {
	m.listener = listener
}
func (m *BaseModule) GetModuleSettings() *conf.ModuleSettings {
	return m.settings
}
func (m *BaseModule) GetRouteServers(moduleType string, hash string) (s module.ServerSession, err error) {
	return m.App.GetRouteServers(moduleType, hash)
}

func (m *BaseModule) RpcInvoke(moduleType string, _func string, params ...interface{}) (result interface{}, err string) {
	server, e := m.App.GetRouteServers(moduleType, m.subclass.GetServerId())
	if e != nil {
		err = e.Error()
		return
	}
	return server.Call(_func, params...)
}

func (m *BaseModule) RpcInvokeNR(moduleType string, _func string, params ...interface{}) (err error) {
	server, err := m.App.GetRouteServers(moduleType, m.subclass.GetServerId())
	if err != nil {
		return
	}
	return server.CallNR(_func, params...)
}

func (m *BaseModule) RpcInvokeArgs(moduleType string, _func string, ArgsType []string,args [][]byte) (result interface{}, err string) {
	server, e := m.App.GetRouteServers(moduleType, m.subclass.GetServerId())
	if e != nil {
		err = e.Error()
		return
	}
	return server.CallArgs(_func, ArgsType,args)
}

func (m *BaseModule) RpcInvokeNRArgs(moduleType string, _func string, ArgsType []string,args [][]byte) (err error) {
	server, err := m.App.GetRouteServers(moduleType, m.subclass.GetServerId())
	if err != nil {
		return
	}
	return server.CallNRArgs(_func, ArgsType,args)
}

func (m *BaseModule) OnTimeOut(fn string, Expired int64) {
	m.rwmutex.RLock()
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
	m.rwmutex.RUnlock()
	if m.listener != nil {
		m.listener.OnTimeOut(fn, Expired)
	}
}
func (m *BaseModule) OnError(fn string, callInfo *mqrpc.CallInfo, err error) {
	m.rwmutex.RLock()
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
	m.rwmutex.RUnlock()
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
	m.rwmutex.RLock()
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
	m.rwmutex.RUnlock()
	if m.listener != nil {
		m.listener.OnComplete(fn, callInfo, result, exec_time)
	}
}
func (m *BaseModule) GetExecuting() int64 {
	return m.GetServer().GetRPCServer().GetExecuting()
}
func (m *BaseModule) GetStatistical() (statistical string, err error) {
	m.rwmutex.RLock()
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
	m.rwmutex.RUnlock()
	return
}
