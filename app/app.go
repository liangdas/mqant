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
package defaultApp

import (
	"flag"
	"fmt"
	"github.com/liangdas/mqant/conf"
	"github.com/liangdas/mqant/log"
	"github.com/liangdas/mqant/module/base"
	"github.com/liangdas/mqant/module"
	"github.com/liangdas/mqant/rpc"
	"hash/crc32"
	"math"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"github.com/liangdas/mqant/rpc/base"
	"github.com/liangdas/mqant/module/modules"
	opentracing "github.com/opentracing/opentracing-go"
)


func NewApp(version string) module.App {
	app := new(DefaultApp)
	app.routes = map[string]func(app module.App, Type string, hash string) module.ServerSession{}
	app.serverList = map[string]module.ServerSession{}
	app.defaultRoutes = func(app module.App, Type string, hash string) module.ServerSession {
		//默认使用第一个Server
		servers := app.GetServersByType(Type)
		if len(servers) == 0 {
			return nil
		}
		index := int(math.Abs(float64(crc32.ChecksumIEEE([]byte(hash))))) % len(servers)
		return servers[index]
	}
	app.rpcserializes=map[string]module.RPCSerialize{}
	app.version = version
	return app
}

type DefaultApp struct {
	module.App
	version       string
	serverList    map[string]module.ServerSession
	settings      conf.Config
	routes        map[string]func(app module.App, Type string, hash string) module.ServerSession
	defaultRoutes func(app module.App, Type string, hash string) module.ServerSession
	rpcserializes	map[string]module.RPCSerialize
	getTracer 	func ()opentracing.Tracer
}

func (app *DefaultApp) Run(debug bool, mods ...module.Module) error {
	file, _ := exec.LookPath(os.Args[0])
	ApplicationPath, _ := filepath.Abs(file)
	ApplicationDir, _ := filepath.Split(ApplicationPath)
	defaultPath := fmt.Sprintf("%sconf/server.conf", ApplicationDir)
	confPath := flag.String("conf", defaultPath, "Server configuration file path")
	ProcessID := flag.String("pid", "development", "Server ProcessID?")
	Logdir := flag.String("log", fmt.Sprintf("%slogs", ApplicationDir), "Log file directory?")
	flag.Parse() //解析输入的参数

	f, err := os.Open(*confPath)
	if err != nil {
		panic(err)
	}

	_, err = os.Open(*Logdir)
	if err != nil {
		//文件不存在
		err := os.Mkdir(*Logdir, os.ModePerm) //
		if err != nil {
			fmt.Println(err)
		}
	}
	log.Init(debug, *ProcessID, *Logdir)
	log.Info("Server configuration file path [%s]", *confPath)
	conf.LoadConfig(f.Name()) //加载配置文件
	app.Configure(conf.Conf)  //配置信息

	log.Info("mqant %v starting up", app.version)
	manager := basemodule.NewModuleManager()
	manager.RegisterRunMod(modules.TimerModule()) //注册时间轮模块 每一个进程都默认运行
	// module
	for i := 0; i < len(mods); i++ {
		manager.Register(mods[i])
	}
	app.OnInit(app.settings)
	manager.Init(app, *ProcessID)
	// close
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	sig := <-c
	manager.Destroy()
	app.OnDestroy()
	log.Info("mqant closing down (signal: %v)", sig)
	return nil
}
func (app *DefaultApp) Route(moduleType string, fn func(app module.App, Type string, hash string) module.ServerSession) error {
	app.routes[moduleType] = fn
	return nil
}
func (app *DefaultApp) getRoute(moduleType string) func(app module.App, Type string, hash string) module.ServerSession {
	fn := app.routes[moduleType]
	if fn == nil {
		//如果没有设置的路由,则使用默认的
		return app.defaultRoutes
	}
	return fn
}


func (app *DefaultApp) AddRPCSerialize(name string, Interface module.RPCSerialize) error{
	if _,ok:=app.rpcserializes[name];ok{
		return fmt.Errorf("The name(%s) has been occupied",name)
	}
	app.rpcserializes[name]=Interface
	return nil
}

func (app *DefaultApp)GetRPCSerialize()(map[string]module.RPCSerialize){
	return app.rpcserializes
}

func (app *DefaultApp) Configure(settings conf.Config) error {
	app.settings = settings
	return nil
}

/**
 */
func (app *DefaultApp) OnInit(settings conf.Config) error {
	app.serverList = make(map[string]module.ServerSession)
	for Type, ModuleInfos := range settings.Module {
		for _, moduel := range ModuleInfos {
			m := app.serverList[moduel.Id]
			if m != nil {
				//如果Id已经存在,说明有两个相同Id的模块,这种情况不能被允许,这里就直接抛异常 强制崩溃以免以后调试找不到问题
				panic(fmt.Sprintf("ServerId (%s) Type (%s) of the modules already exist Can not be reused ServerId (%s) Type (%s)", m.GetId(), m.GetType(), moduel.Id, Type))
			}
			client, err := defaultrpc.NewRPCClient(app,moduel.Id)
			if err != nil {
				continue
			}
			if moduel.Rabbitmq != nil {
				//如果远程的rpc存在则创建一个对应的客户端
				client.NewRemoteClient(moduel.Rabbitmq)
			}
			session := basemodule.NewServerSession(moduel.Id,Type,client)
			app.serverList[moduel.Id] = session
			log.Info("RPCClient create success type(%s) id(%s)", Type, moduel.Id)
		}
	}
	return nil
}

func (app *DefaultApp) OnDestroy() error {
	for id, session := range app.serverList {
		err := session.GetRpc().Done()
		if err != nil {
			log.Warning("RPCClient close fail type(%s) id(%s)", session.GetType(), id)
		} else {
			log.Info("RPCClient close success type(%s) id(%s)", session.GetType(), id)
		}
	}
	return nil
}

func (app *DefaultApp) RegisterLocalClient(serverId string, server mqrpc.RPCServer) error {
	if session, ok := app.serverList[serverId]; ok {
		return session.GetRpc().NewLocalClient(server)
	} else {
		return fmt.Errorf("Server(%s) Not Found", serverId)
	}
	return nil
}

func (app *DefaultApp) GetServersById(serverId string) (module.ServerSession, error) {
	if session, ok := app.serverList[serverId]; ok {
		return session, nil
	} else {
		return nil, fmt.Errorf("Server(%s) Not Found", serverId)
	}
}

func (app *DefaultApp) GetServersByType(Type string) []module.ServerSession {
	sessions := make([]module.ServerSession, 0)
	for _, session := range app.serverList {
		if session.GetType() == Type {
			sessions = append(sessions, session)
		}
	}
	return sessions
}

func (app *DefaultApp) GetRouteServers(filter string, hash string) (s module.ServerSession, err error) {
	sl := strings.Split(filter, "@")
	if len(sl) == 2 {
		moduleID := sl[1]
		if moduleID != "" {
			return app.GetServersById(moduleID)
		}
	}
	moduleType := sl[0]
	route := app.getRoute(moduleType)
	s = route(app, moduleType, hash)
	if s == nil {
		err = fmt.Errorf("Server(type : %s) Not Found", moduleType)
	}
	return
}

func (app *DefaultApp) GetSettings() conf.Config {
	return app.settings
}

func (app *DefaultApp) RpcInvoke(module module.RPCModule, moduleType string, _func string, params ...interface{}) (result interface{}, err string) {
	server, e := app.GetRouteServers(moduleType, module.GetServerId())
	if e != nil {
		err = e.Error()
		return
	}
	return server.Call(_func, params...)
}

func (app *DefaultApp) RpcInvokeNR(module module.RPCModule, moduleType string, _func string, params ...interface{}) (err error) {
	server, err := app.GetRouteServers(moduleType, module.GetServerId())
	if err != nil {
		return
	}
	return server.CallNR(_func, params...)
}

func (app *DefaultApp) RpcInvokeArgs(module module.RPCModule, moduleType string, _func string, ArgsType []string,args [][]byte) (result interface{}, err string) {
	server, e := app.GetRouteServers(moduleType, module.GetServerId())
	if e != nil {
		err = e.Error()
		return
	}
	return server.CallArgs(_func, ArgsType,args)
}

func (app *DefaultApp) RpcInvokeNRArgs(module module.RPCModule, moduleType string, _func string, ArgsType []string,args [][]byte) (err error) {
	server, err := app.GetRouteServers(moduleType, module.GetServerId())
	if err != nil {
		return
	}
	return server.CallNRArgs(_func, ArgsType,args)
}
func (app *DefaultApp)DefaultTracer(_func func ()opentracing.Tracer) error{
	app.getTracer=_func
	return nil
}
func (app *DefaultApp)GetTracer()opentracing.Tracer{
	if app.getTracer!=nil{
		return app.getTracer()
	}
	return nil
}