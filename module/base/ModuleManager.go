// Copyright 2014 loolgame Author. All Rights Reserved.
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

// Package basemodule  模块管理器
package basemodule

import (
	"fmt"
	"github.com/liangdas/mqant/conf"
	"github.com/liangdas/mqant/log"
	"github.com/liangdas/mqant/module"
)

// NewModuleManager 新建模块管理器
func NewModuleManager() (m *ModuleManager) {
	m = new(ModuleManager)
	return
}

// ModuleManager 模块管理器
type ModuleManager struct {
	app     module.App
	mods    []*DefaultModule
	runMods []*DefaultModule
}

// Register 注册模块
func (mer *ModuleManager) Register(mi module.Module) {
	md := new(DefaultModule)
	md.mi = mi
	md.closeSig = make(chan bool, 1)

	mer.mods = append(mer.mods, md)
}

// RegisterRunMod 注册需要运行的模块
func (mer *ModuleManager) RegisterRunMod(mi module.Module) {
	md := new(DefaultModule)
	md.mi = mi
	md.closeSig = make(chan bool, 1)

	mer.runMods = append(mer.runMods, md)
}

// Init 初始化
func (mer *ModuleManager) Init(app module.App, ProcessID string) {
	log.Info("This service ModuleGroup(ProcessID) is [%s]", ProcessID)
	mer.app = app
	mer.CheckModuleSettings() //配置文件规则检查
	for i := 0; i < len(mer.mods); i++ {
		for Type, modSettings := range app.GetSettings().Module {
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

		if app.GetModuleInited() != nil {
			app.GetModuleInited()(app, m.mi)
		}

		m.wg.Add(1)
		go run(m)
	}
	//timer.SetTimer(3, mer.ReportStatistics, nil) //统计汇报定时任务
}

// CheckModuleSettings module配置文件规则检查
// ID全局必须唯一
// 每一个类型的Module列表中ProcessID不能重复
func (mer *ModuleManager) CheckModuleSettings() {
	gid := map[string]string{} //用来保存全局ID-ModuleType
	for Type, modSettings := range conf.Conf.Module {
		pid := map[string]string{} //用来保存模块中的 ProcessID-ID
		for _, setting := range modSettings {
			if Stype, ok := gid[setting.ID]; ok {
				//如果Id已经存在,说明有两个相同Id的模块,这种情况不能被允许,这里就直接抛异常 强制崩溃以免以后调试找不到问题
				panic(fmt.Sprintf("ID (%s) been used in modules of type [%s] and cannot be reused", setting.ID, Stype))
			} else {
				gid[setting.ID] = Type
			}

			if id, ok := pid[setting.ProcessID]; ok {
				//如果Id已经存在,说明有两个相同Id的模块,这种情况不能被允许,这里就直接抛异常 强制崩溃以免以后调试找不到问题
				panic(fmt.Sprintf("In the list of modules of type [%s], ProcessID (%s) has been used for ID module for (%s)", Type, setting.ProcessID, id))
			} else {
				pid[setting.ProcessID] = setting.ID
			}
		}
	}
}

// Destroy 停止模块
func (mer *ModuleManager) Destroy() {
	for i := len(mer.runMods) - 1; i >= 0; i-- {
		m := mer.runMods[i]
		m.closeSig <- true
		m.wg.Wait()
		destroy(m)
	}
}
