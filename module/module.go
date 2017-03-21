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
package module

import (
	"fmt"
	"github.com/liangdas/mqant/conf"
	"github.com/liangdas/mqant/log"
	"github.com/liangdas/mqant/module/modules/timer"
	"github.com/liangdas/mqant/rpc"
	"runtime"
	"sync"
)

type App interface {
	Run(debug bool, mods ...Module) error
	/**
	当同一个类型的Module存在多个服务时,需要根据情况选择最终路由到哪一个服务去
	fn: func(moduleType string,serverId string,[]*ServerSession)(*ServerSession)
	*/
	Route(moduleType string, fn func(app App, Type string, hash string) *ServerSession) error
	Configure(settings conf.Config) error
	OnInit(settings conf.Config) error
	OnDestroy() error
	RegisterLocalClient(serverId string, server *mqrpc.RPCServer) error
	GetServersById(id string) (*ServerSession, error)
	/**
	filter		 调用者服务类型    moduleType|moduleType@moduleID
	Type	   	想要调用的服务类型
	*/
	GetRouteServers(filter string, hash string) (*ServerSession, error) //获取经过筛选过的服务
	GetServersByType(Type string) []*ServerSession
	GetSettings() conf.Config //获取配置信息
	RpcInvoke(module RPCModule, moduleType string, _func string, params ...interface{}) (interface{}, string)
	RpcInvokeNR(module RPCModule, moduleType string, _func string, params ...interface{}) error
}

type ServerSession struct {
	Id    string
	Stype string
	Rpc   *mqrpc.RPCClient
}

/**
消息请求 需要回复
*/
func (c *ServerSession) Call(_func string, params ...interface{}) (interface{}, string) {
	return c.Rpc.Call(_func, params...)
}

/**
消息请求 需要回复
*/
func (c *ServerSession) CallNR(_func string, params ...interface{}) (err error) {
	return c.Rpc.CallNR(_func, params...)
}

type Module interface {
	Version() string //模块版本
	GetType() string //模块类型
	OnInit(app App, settings *conf.ModuleSettings)
	OnDestroy()
	Run(closeSig chan bool)
}
type RPCModule interface {
	Module
	GetServerId() string //模块类型
	RpcInvoke(moduleType string, _func string, params ...interface{}) (interface{}, string)
	RpcInvokeNR(moduleType string, _func string, params ...interface{}) error
	GetModuleSettings() (settings *conf.ModuleSettings)
	/**
	filter		 调用者服务类型    moduleType|moduleType@moduleID
	Type	   	想要调用的服务类型
	*/
	GetRouteServers(filter string, hash string) (*ServerSession, error)
	GetStatistical() (statistical string, err error)
	GetExecuting() int64
}

type module struct {
	mi       Module
	settings *conf.ModuleSettings
	closeSig chan bool
	wg       sync.WaitGroup
}

func NewModuleManager() (m *ModuleManager) {
	m = new(ModuleManager)
	return
}

type ModuleManager struct {
	app     App
	mods    []*module
	runMods []*module
}

func (mer *ModuleManager) Register(mi Module) {
	md := new(module)
	md.mi = mi
	md.closeSig = make(chan bool, 1)

	mer.mods = append(mer.mods, md)
}
func (mer *ModuleManager) RegisterRunMod(mi Module) {
	md := new(module)
	md.mi = mi
	md.closeSig = make(chan bool, 1)

	mer.runMods = append(mer.runMods, md)
}

func (mer *ModuleManager) Init(app App, ProcessID string) {
	log.Info("This service ProcessID is [%s]", ProcessID)
	mer.app = app
	mer.CheckModuleSettings() //配置文件规则检查
	for i := 0; i < len(mer.mods); i++ {
		for Type, modSettings := range conf.Conf.Module {
			if mer.mods[i].mi.GetType() == Type {
				//匹配
				for _, setting := range modSettings {
					//这里可能有BUG 公网IP和局域网IP处理方式可能不一样,先不管
					if ProcessID == setting.ProcessID {
						mer.runMods = append(mer.runMods, mer.mods[i]) //这里加入能够运行的组件
						mer.mods[i].settings = setting
					}
				}
				break //跳出内部循环
			}
		}
	}

	for i := 0; i < len(mer.runMods); i++ {
		m := mer.runMods[i]
		m.mi.OnInit(app, m.settings)
		m.wg.Add(1)
		go run(m)
	}
	timer.SetTimer(3, mer.ReportStatistics, nil) //统计汇报定时任务
}

/**
module配置文件规则检查
1. ID全局必须唯一
2. 每一个类型的Module列表中ProcessID不能重复
*/
func (mer *ModuleManager) CheckModuleSettings() {
	gid := map[string]string{} //用来保存全局ID-ModuleType
	for Type, modSettings := range conf.Conf.Module {
		pid := map[string]string{} //用来保存模块中的 ProcessID-ID
		for _, setting := range modSettings {
			if Stype, ok := gid[setting.Id]; ok {
				//如果Id已经存在,说明有两个相同Id的模块,这种情况不能被允许,这里就直接抛异常 强制崩溃以免以后调试找不到问题
				panic(fmt.Sprintf("ID (%s) been used in modules of type [%s] and cannot be reused", setting.Id, Stype))
			} else {
				gid[setting.Id] = Type
			}

			if Id, ok := pid[setting.ProcessID]; ok {
				//如果Id已经存在,说明有两个相同Id的模块,这种情况不能被允许,这里就直接抛异常 强制崩溃以免以后调试找不到问题
				panic(fmt.Sprintf("In the list of modules of type [%s], ProcessID (%s) has been used for ID module for (%s)", Type, setting.ProcessID, Id))
			} else {
				pid[setting.ProcessID] = setting.Id
			}
		}
	}
}
func (mer *ModuleManager) Destroy() {
	for i := len(mer.runMods) - 1; i >= 0; i-- {
		m := mer.runMods[i]
		m.closeSig <- true
		m.wg.Wait()
		destroy(m)
	}
}

func run(m *module) {
	defer func() {
		if r := recover(); r != nil {
			if conf.LenStackBuf > 0 {
				buf := make([]byte, conf.LenStackBuf)
				l := runtime.Stack(buf, false)
				log.Error("%v: %s", r, buf[:l])
			} else {
				log.Error("%v", r)
			}
		}
	}()
	m.mi.Run(m.closeSig)
	m.wg.Done()
}

func destroy(m *module) {
	defer func() {
		if r := recover(); r != nil {
			if conf.LenStackBuf > 0 {
				buf := make([]byte, conf.LenStackBuf)
				l := runtime.Stack(buf, false)
				log.Error("%v: %s", r, buf[:l])
			} else {
				log.Error("%v", r)
			}
		}
	}()
	m.mi.OnDestroy()
}

func (mer *ModuleManager) ReportStatistics(args interface{}) {
	if mer.app.GetSettings().Master.Enable {
		for _, m := range mer.runMods {
			mi := m.mi
			switch value := mi.(type) {
			case RPCModule:
				//汇报统计
				servers := mer.app.GetServersByType("Master")
				if len(servers) == 1 {
					b, _ := value.GetStatistical()
					_, err := servers[0].Call("ReportForm", value.GetType(), m.settings.ProcessID, m.settings.Id, value.Version(), b, value.GetExecuting())
					if err != "" {
						log.Warning("Report To Master error :", err)
					}
				}
			default:
			}
		}
		timer.SetTimer(3, mer.ReportStatistics, nil)
	}
}
