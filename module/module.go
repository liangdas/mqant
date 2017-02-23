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
	"runtime"
	"sync"
	"github.com/liangdas/mqant/log"
	"github.com/liangdas/mqant/conf"
	"fmt"
	"net"
	"os"
)

type Module interface {
	GetType()(string)	//模块类型
	OnInit(app App,settings *conf.ModuleSettings)
	OnDestroy()
	Run(closeSig chan bool)
}

type module struct {
	mi       Module
	closeSig chan bool
	wg       sync.WaitGroup
}

func NewModuleManager()(m *ModuleManager){
	m=new(ModuleManager)
	return
}

type ModuleManager struct {
	mods 	[]*module
	runMods []*module
}

func (mer *ModuleManager)Register(mi Module) {
	md := new(module)
	md.mi = mi
	md.closeSig = make(chan bool, 1)

	mer.mods = append(mer.mods, md)
}

func (mer *ModuleManager)Init(app App) {
	myHost:=getMySelfHost()
	log.Release("MySelfHost %s",myHost)
	for i := 0; i < len(mer.mods); i++ {
		for Type,modSettings:=range conf.Conf.Module {
			if mer.mods[i].mi.GetType()==Type{
				//匹配
				for _,setting:=range modSettings{
					//这里可能有BUG 公网IP和局域网IP处理方式可能不一样,先不管
					if myHost==setting.Host||"127.0.0.1"==setting.Host{
						mer.runMods = append(mer.runMods, mer.mods[i])	//这里加入能够运行的组件
						mer.mods[i].mi.OnInit(app,setting)
					}

				}
				break	//跳出内部循环
			}
		}

	}

	for i := 0; i < len(mer.runMods); i++ {
		m := mer.runMods[i]
		m.wg.Add(1)
		go run(m)
	}
}

func (mer *ModuleManager)Destroy() {
	for i := len(mer.runMods) - 1; i >= 0; i-- {
		m := mer.runMods[i]
		m.closeSig <- true
		m.wg.Wait()
		destroy(m)
	}
}

func run(m *module) {
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
func getMySelfHost()(string){
	addrs, err := net.InterfaceAddrs()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for _, address := range addrs {

		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}

		}
	}
	return	""
}
