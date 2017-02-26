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
package module

import (
	"github.com/liangdas/mqant/conf"
)

type BaseModule struct {
	App		App
	subclass 	Module
	settings 	*conf.ModuleSettings
	server 		*Server
}
func (m *BaseModule) GetServerId()(string){
	//很关键,需要与配置文件中的Module配置对应
	return m.settings.Id
}
func (m *BaseModule) GetServer() (*Server){
	if m.server==nil{
		m.server = new(Server)
	}
	return m.server
}

func (m *BaseModule) OnInit(subclass Module,app App,settings *conf.ModuleSettings) {
	//初始化模块
	m.App=app
	m.subclass=subclass
	m.settings=settings
	//创建一个远程调用的RPC
	m.GetServer().OnInit(app,settings)
}

func (m *BaseModule) OnDestroy() {
	//注销模块
	//一定别忘了关闭RPC
	m.GetServer().OnDestroy()
}
func (m *BaseModule)GetModuleSettings()(*conf.ModuleSettings){
	return m.settings
}
func (m *BaseModule)GetRouteServersByType(moduleType string)(s *ServerSession,err error)  {
	return	m.App.GetRouteServersByType(m.subclass,moduleType)
}

func (m *BaseModule)RpcInvoke(moduleType string,_func string,params ...interface{})(result interface{},err string)  {
	server,e:=m.App.GetRouteServersByType(m.subclass,moduleType)
	if e!=nil{
		err=e.Error()
		return
	}
	return server.Call(_func,params...)
}

func (m *BaseModule)RpcInvokeNR(moduleType string,_func string,params ...interface{})(err error)  {
	server,err:=m.App.GetRouteServersByType(m.subclass,moduleType)
	if err!=nil{
		return
	}
	return server.CallNR(_func,params...)
}
