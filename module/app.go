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
	"github.com/liangdas/mqant/rpc"
	"github.com/liangdas/mqant/log"
	"github.com/liangdas/mqant/conf"
	"os"
	"os/signal"
)

func NewApp()(App){
	app:=new(DefaultApp)
	app.routes=map[string]func(app App,moduleType string,serverId string,Type string) (*ServerSession){}
	app.serverList=map[string]*ServerSession{}
	app.defaultRoutes= func(app App,moduleType string,serverId string,Type string) (*ServerSession){
		//默认使用第一个Server
		servers:=app.GetServersByType(Type)
		if len(servers)==0{
			return nil
		}
		return servers[0]
	}
	return app
}

type App interface {
	Run(mods ...Module)	error
	/**
	当同一个类型的Module存在多个服务时,需要根据情况选择最终路由到哪一个服务去
	fn: func(moduleType string,serverId string,[]*ServerSession)(*ServerSession)
	 */
	Route(moduleType string,fn func( app App,moduleType string,serverId string,Type string) (*ServerSession)) error
	Configure(settings conf.Config)error
	OnInit(settings conf.Config) error
	OnDestroy() error
	RegisterLocalClient(serverId string,server *mqrpc.RPCServer) error
	GetServersById(id string)(*ServerSession,error)
	/**
	moduleType 调用者服务类型
	serverId   调用者服务ID
	Type	   想要调用的服务类型
	 */
	GetRouteServersByType(module Module,moduleType string)(*ServerSession,error)	//获取经过筛选过的服务
	GetServersByType(Type string)([]*ServerSession)
	GetSettings()(conf.Config) //获取配置信息
	RpcInvoke(module Module,moduleType string,_func string,params ...interface{})(interface{},string)
	RpcInvokeNR(module Module,moduleType string,_func string,params ...interface{})(error)
}

type ServerSession struct {
	id 	string
	stype   string
	rpc	*mqrpc.RPCClient
}

/**
消息请求 需要回复
 */
func (c *ServerSession) Call(_func string,params ...interface{})(interface{},string)  {
	return c.rpc.Call(_func,params...)
}


/**
消息请求 需要回复
 */
func (c *ServerSession) CallNR(_func string,params ...interface{})(err error)  {
	return c.rpc.CallNR(_func,params...)
}


type DefaultApp struct {
	App
	serverList map[string]*ServerSession
	settings   conf.Config
	routes     map[string]func(app App,moduleType string,serverId string,Type string) (*ServerSession)
	defaultRoutes func(app App,moduleType string,serverId string,Type string) (*ServerSession)
}

func (app *DefaultApp)Run(mods ...Module)error{
	// logger
	if conf.LogLevel != "" {
		logger, err := log.New(conf.LogLevel, conf.LogPath, conf.LogFlag)
		if err != nil {
			panic(err)
		}
		log.Export(logger)
		defer logger.Close()
	}

	log.Release("mqant %v starting up", version)
	manager:=NewModuleManager()
	// module
	for i := 0; i < len(mods); i++ {
		manager.Register(mods[i])
	}
	app.OnInit(app.settings)
	manager.Init(app)
	// close
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	sig := <-c
	manager.Destroy()
	app.OnDestroy()
	log.Release("mqant closing down (signal: %v)", sig)
	return nil
}
func (app *DefaultApp) Route(moduleType string,fn func(app App,moduleType string,serverId string,Type string) (*ServerSession)) error{
	app.routes[moduleType]=fn
	return nil
}
func (app *DefaultApp) getRoute(moduleType string) (func(app App,moduleType string,serverId string,Type string) (*ServerSession)){
	fn:=app.routes[moduleType]
	if fn==nil{
		//如果没有设置的路由,则使用默认的
		return app.defaultRoutes
	}
	return fn
}

func (app *DefaultApp) Configure(settings conf.Config)error{
	app.settings=settings
	return nil
}

/**
 */
func (app *DefaultApp)OnInit(settings conf.Config) error{
	app.serverList=make(map[string]*ServerSession)
	for Type,ModuleInfos :=range settings.Module{
		for _,moduel:=range ModuleInfos{
			m:=app.serverList[moduel.Id]
			if m!=nil{
				//如果Id已经存在,说明有两个相同Id的模块,这种情况不能被允许,这里就直接抛异常 强制崩溃以免以后调试找不到问题
				panic(fmt.Sprintf("ServerId (%s) Type (%s) of the modules already exist Can not be reused ServerId (%s) Type (%s)", m.id,m.stype,moduel.Id,Type))
			}
			client,err:=mqrpc.NewRPCClient()
			if err != nil {
				continue
			}
			if moduel.Rabbitmq!=nil {
				//如果远程的rpc存在则创建一个对应的客户端
				client.NewRemoteClient(moduel.Rabbitmq)
			}
			session:=&ServerSession{
				id:moduel.Id,
				stype:Type,
				rpc:client,
			}
			app.serverList[moduel.Id]=session
			log.Debug("RPCClient create success type(%s) id(%s)",Type,moduel.Id)
		}
	}
	return nil
}

func (app *DefaultApp)OnDestroy()error{
	for id,session:=range app.serverList{
		err:=session.rpc.Done()
		if err!=nil{
			log.Error("RPCClient close fail type(%s) id(%s)",session.stype,id)
		}else{
			log.Debug("RPCClient close success type(%s) id(%s)",session.stype,id)
		}
	}
	return nil
}

func (app *DefaultApp)RegisterLocalClient(serverId string,server *mqrpc.RPCServer)(error)  {
	if session, ok := app.serverList[serverId]; ok {
		return session.rpc.NewLocalClient(server)
	} else {
		return fmt.Errorf("Server(%s) Not Found",serverId)
	}
	return nil
}

func (app *DefaultApp)GetServersById(serverId string)(*ServerSession,error)  {
	if session, ok := app.serverList[serverId]; ok {
		return session,nil
	} else {
		return nil,fmt.Errorf("Server(%s) Not Found",serverId)
	}
}

func (app *DefaultApp)GetServersByType(Type string)([]*ServerSession)  {
	sessions:=make([]*ServerSession,0)
	for _,session:=range app.serverList{
		if session.stype==Type{
			sessions=append(sessions,session)
		}
	}
	return sessions
}

func (app *DefaultApp)GetRouteServersByType(module Module,moduleType string)(s *ServerSession,err error)  {
	route:=app.getRoute(moduleType)
	s=route(app,module.GetType(),module.GetModuleSettings().Id,moduleType)
	if s==nil{
		err=fmt.Errorf("Server(type : %s) Not Found",moduleType)
	}
	return
}

func (app *DefaultApp)GetSettings()(conf.Config){
	return app.settings
}

func (app *DefaultApp)RpcInvoke(module Module,moduleType string,_func string,params ...interface{})(result interface{},err string)  {
	server,e:=app.GetRouteServersByType(module,moduleType)
	if e!=nil{
		err=e.Error()
		return
	}
	return server.Call(_func,params...)
}

func (app *DefaultApp)RpcInvokeNR(module Module,moduleType string,_func string,params ...interface{})(err error)  {
	server,err:=app.GetRouteServersByType(module,moduleType)
	if err!=nil{
		return
	}
	return server.CallNR(_func,params...)
}