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
	"encoding/json"
	"flag"
	"fmt"
	"github.com/liangdas/mqant/conf"
	"github.com/liangdas/mqant/log"
	"github.com/liangdas/mqant/module"
	"github.com/liangdas/mqant/module/base"
	"github.com/liangdas/mqant/module/modules"
	"github.com/liangdas/mqant/rpc"
	"github.com/liangdas/mqant/rpc/base"
	"hash/crc32"
	"math"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

type resultInfo struct {
	Trace  string
	Error  string      //错误结果 如果为nil表示请求正确
	Result interface{} //结果
}

type protocolMarshalImp struct {
	data []byte
}

// GetData 获取app数据
func (this *protocolMarshalImp) GetData() []byte {
	return this.data
}

// NewApp 创建app
func NewApp(version string, debug bool, options ...interface{}) module.App {
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
	app.rpcserializes = map[string]module.RPCSerialize{}
	app.version = version
	// --------------------------- 执行app初始化配置 ---------------------------
	// 从options中解析初始化选项
	var (
		optConf *conf.Config
	)
	for _, option := range options {
		switch o := option.(type) {
		case conf.Config:
			optConf = &o
		default:
			panic(option)
		}
	}
	// 从flag中准备参数
	wdPath := flag.String("wd", "", "Server work directory")
	confPath := flag.String("conf", "", "Server configuration file path")
	ProcessID := flag.String("pid", "development", "Server ProcessID?")
	Logdir := flag.String("log", "", "Log file directory?")
	BIdir := flag.String("bi", "", "bi file directory?")
	flag.Parse() //解析输入的参数
	app.processId = *ProcessID
	ApplicationDir := ""
	// 根据策略选择working Path工作目录
	if *wdPath != "" {
		_, err := os.Open(*wdPath)
		if err != nil {
			panic(err)
		}
		os.Chdir(*wdPath)
		ApplicationDir, err = os.Getwd()
	} else {
		var err error
		ApplicationDir, err = os.Getwd()
		if err != nil {
			file, _ := exec.LookPath(os.Args[0])
			ApplicationPath, _ := filepath.Abs(file)
			ApplicationDir, _ = filepath.Split(ApplicationPath)
		}

	}

	defaultConfPath := fmt.Sprintf("%s/bin/conf/server.json", ApplicationDir)
	defaultLogPath := fmt.Sprintf("%s/bin/logs", ApplicationDir)
	defaultBIPath := fmt.Sprintf("%s/bin/bi", ApplicationDir)

	// 根据策略选择配置路径
	if *confPath == "" {
		// 如果未给出conf配置，则使用传入的conf（终端conf参数优先级高于传入conf，最低优先级才是默认confPath
		if optConf == nil {
			// 也未传入conf，则尝试使用默认的confPath来loadConf
			conf.LoadConfig(defaultConfPath)
			optConf = &conf.Conf
		}
	} else {
		// flag参数有conf，则最优先使用此conf
		conf.LoadConfig(*confPath)
		optConf = &conf.Conf
	}

	// 根据策略选择日志目录
	if *Logdir == "" {
		*Logdir = defaultLogPath
	}

	// 根据策略选择二进制文件目录
	if *BIdir == "" {
		*BIdir = defaultBIPath
	}

	// 在conf.LoadConfig时，如果读取错误，会直接panic，所以无需再此多检查一次是否可以打开
	// 验证conf文件是否可以打开
	// f, err := os.Open(*confPath)
	// if err != nil {
	// 	panic(err)
	// }

	// 检查日志目录是否存在
	_, err := os.Open(*Logdir)
	if err != nil {
		//文件不存在
		err := os.Mkdir(*Logdir, os.ModePerm) //
		if err != nil {
			fmt.Println(err)
		}
	}

	// 检查二进制目录是否存在
	_, err = os.Open(*BIdir)
	if err != nil {
		//文件不存在
		err := os.Mkdir(*BIdir, os.ModePerm) //
		if err != nil {
			fmt.Println(err)
		}
	}

	// 加载配置
	fmt.Println("Server configuration file path :", *confPath)
	app.Configure(*optConf) //配置信息
	// 配置文件加载完毕
	log.InitLog(debug, *ProcessID, *Logdir, optConf.Log)
	log.InitBI(debug, *ProcessID, *BIdir, optConf.BI)
	// app初始化完成
	return app
}

// DefaultApp 默认app
type DefaultApp struct {
	//module.App
	version       string
	serverList    map[string]module.ServerSession
	settings      conf.Config
	processId     string
	routes        map[string]func(app module.App, Type string, hash string) module.ServerSession
	defaultRoutes func(app module.App, Type string, hash string) module.ServerSession
	//将一个RPC调用路由到新的路由上
	mapRoute            func(app module.App, route string) string
	rpcserializes       map[string]module.RPCSerialize
	configurationLoaded func(app module.App)
	startup             func(app module.App)
	moduleInited        func(app module.App, module module.Module)
	protocolMarshal     func(Trace string, Result interface{}, Error string) (module.ProtocolMarshal, string)
}

// Run 运行app
func (app *DefaultApp) Run(mods ...module.Module) error {

	log.Info("mqant %v starting up", app.version)

	if app.configurationLoaded != nil {
		app.configurationLoaded(app)
	}

	manager := basemodule.NewModuleManager()
	manager.RegisterRunMod(modules.TimerModule()) //注册时间轮模块 每一个进程都默认运行
	// module
	for i := 0; i < len(mods); i++ {
		mods[i].OnAppConfigurationLoaded(app)
		manager.Register(mods[i])
	}
	app.OnInit(app.settings)
	manager.Init(app, app.processId)
	if app.startup != nil {
		app.startup(app)
	}
	// close
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM)
	sig := <-c

	//如果一分钟都关不了则强制关闭
	timeout := time.NewTimer(time.Minute)
	wait := make(chan struct{})
	go func() {
		manager.Destroy()
		app.OnDestroy()
		wait <- struct{}{}
	}()
	select {
	case <-timeout.C:
		panic(fmt.Sprintf("mqant close timeout (signal: %v)", sig))
	case <-wait:
		log.Info("mqant closing down (signal: %v)", sig)
	}
	return nil
}

// Route 按模块设置路由方法
func (app *DefaultApp) Route(moduleType string, fn func(app module.App, Type string, hash string) module.ServerSession) error {
	app.routes[moduleType] = fn
	return nil
}

// SetMapRoute 设置Map路由
func (app *DefaultApp) SetMapRoute(fn func(app module.App, route string) string) error {
	app.mapRoute = fn
	return nil
}

// getRoute 按模块获取路由方法
func (app *DefaultApp) getRoute(moduleType string) func(app module.App, Type string, hash string) module.ServerSession {
	fn := app.routes[moduleType]
	if fn == nil {
		//如果没有设置的路由,则使用默认的
		return app.defaultRoutes
	}
	return fn
}

// AddRPCSerialize 添加rpc序列化器
func (app *DefaultApp) AddRPCSerialize(name string, Interface module.RPCSerialize) error {
	if _, ok := app.rpcserializes[name]; ok {
		return fmt.Errorf("The name(%s) has been occupied", name)
	}
	app.rpcserializes[name] = Interface
	return nil
}

// GetRPCSerialize 获取rpc序列化器
func (app *DefaultApp) GetRPCSerialize() map[string]module.RPCSerialize {
	return app.rpcserializes
}

// Configure 设置配置
func (app *DefaultApp) Configure(settings conf.Config) error {
	app.settings = settings
	return nil
}

// OnInit app初始化方法
func (app *DefaultApp) OnInit(settings conf.Config) error {
	app.serverList = make(map[string]module.ServerSession)
	for Type, ModuleInfos := range settings.Module {
		for _, moduel := range ModuleInfos {
			m := app.serverList[moduel.Id]
			if m != nil {
				//如果Id已经存在,说明有两个相同Id的模块,这种情况不能被允许,这里就直接抛异常 强制崩溃以免以后调试找不到问题
				panic(fmt.Sprintf("ServerId (%s) Type (%s) of the modules already exist Can not be reused ServerId (%s) Type (%s)", m.GetId(), m.GetType(), moduel.Id, Type))
			}
			client, err := defaultrpc.NewRPCClient(app, moduel.Id)
			if err != nil {
				continue
			}
			if moduel.Rabbitmq != nil {
				//如果远程的rpc存在则创建一个对应的客户端
				client.NewRabbitmqClient(moduel.Rabbitmq)
			}
			if moduel.Redis != nil {
				//如果远程的rpc存在则创建一个对应的客户端
				client.NewRedisClient(moduel.Redis)
			}
			if moduel.UDP != nil {
				//如果远程的rpc存在则创建一个对应的客户端
				client.NewUdpClient(moduel.UDP)
			}
			session := basemodule.NewServerSession(moduel.Id, Type, client)
			app.serverList[moduel.Id] = session
			log.Info("RPCClient create success type(%s) id(%s)", Type, moduel.Id)
		}
	}
	return nil
}

// OnDestroy app析构函数
func (app *DefaultApp) OnDestroy() error {
	for id, session := range app.serverList {
		log.Info("RPCClient closeing type(%s) id(%s)", session.GetType(), id)
		err := session.GetRpc().Done()
		if err != nil {
			log.Warning("RPCClient close fail type(%s) id(%s)", session.GetType(), id)
		} else {
			log.Info("RPCClient close success type(%s) id(%s)", session.GetType(), id)
		}
	}
	return nil
}

// RegisterLocalClient 注册本地rpc client
func (app *DefaultApp) RegisterLocalClient(serverId string, server mqrpc.RPCServer) error {
	if session, ok := app.serverList[serverId]; ok {
		return session.GetRpc().NewLocalClient(server)
	} else {
		return fmt.Errorf("Server(%s) Not Found", serverId)
	}
	return nil
}

// GetServerById 根据id获取服务
func (app *DefaultApp) GetServerById(serverId string) (module.ServerSession, error) {
	if session, ok := app.serverList[serverId]; ok {
		return session, nil
	} else {
		return nil, fmt.Errorf("Server(%s) Not Found", serverId)
	}
}

// GetServersByType 根据类型获取服务列表
func (app *DefaultApp) GetServersByType(Type string) []module.ServerSession {
	sessions := make([]module.ServerSession, 0)
	for _, session := range app.serverList {
		if session.GetType() == Type {
			sessions = append(sessions, session)
		}
	}
	return sessions
}

// GetRouteServer
func (app *DefaultApp) GetRouteServer(filter string, hash string) (s module.ServerSession, err error) {
	if app.mapRoute != nil {
		//进行一次路由转换
		filter = app.mapRoute(app, filter)
	}
	sl := strings.Split(filter, "@")
	if len(sl) == 2 {
		moduleID := sl[1]
		if moduleID != "" {
			return app.GetServerById(moduleID)
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

// GetSettings 获取设置
func (app *DefaultApp) GetSettings() conf.Config {
	return app.settings
}

// GetProcessID 获取进程ID
func (app *DefaultApp) GetProcessID() string {
	return app.processId
}

// RpcInvoke rpc调用
func (app *DefaultApp) RpcInvoke(module module.RPCModule, moduleType string, _func string, params ...interface{}) (result interface{}, err string) {
	server, e := app.GetRouteServer(moduleType, module.GetServerId())
	if e != nil {
		err = e.Error()
		return
	}
	return server.Call(_func, params...)
}

// RpcInvokeNR rpcNR调用
func (app *DefaultApp) RpcInvokeNR(module module.RPCModule, moduleType string, _func string, params ...interface{}) (err error) {
	server, err := app.GetRouteServer(moduleType, module.GetServerId())
	if err != nil {
		return
	}
	return server.CallNR(_func, params...)
}

// RpcInvokeArgs rpc args调用
func (app *DefaultApp) RpcInvokeArgs(module module.RPCModule, moduleType string, _func string, ArgsType []string, args [][]byte) (result interface{}, err string) {
	server, e := app.GetRouteServer(moduleType, module.GetServerId())
	if e != nil {
		err = e.Error()
		return
	}
	return server.CallArgs(_func, ArgsType, args)
}

// RpcInvokeNRArgs rpc NR args调用
func (app *DefaultApp) RpcInvokeNRArgs(module module.RPCModule, moduleType string, _func string, ArgsType []string, args [][]byte) (err error) {
	server, err := app.GetRouteServer(moduleType, module.GetServerId())
	if err != nil {
		return
	}
	return server.CallNRArgs(_func, ArgsType, args)
}

// GetModuleInited 获取inited模块
func (app *DefaultApp) GetModuleInited() func(app module.App, module module.Module) {
	return app.moduleInited
}

// OnConfigurationLoaded
func (app *DefaultApp) OnConfigurationLoaded(_func func(app module.App)) error {
	app.configurationLoaded = _func
	return nil
}

// OnModuleInited
func (app *DefaultApp) OnModuleInited(_func func(app module.App, module module.Module)) error {
	app.moduleInited = _func
	return nil
}

// OnStartup
func (app *DefaultApp) OnStartup(_func func(app module.App)) error {
	app.startup = _func
	return nil
}

// SetProtocolMarshal
func (app *DefaultApp) SetProtocolMarshal(protocolMarshal func(Trace string, Result interface{}, Error string) (module.ProtocolMarshal, string)) error {
	app.protocolMarshal = protocolMarshal
	return nil
}

// ProtocolMarshal
func (app *DefaultApp) ProtocolMarshal(Trace string, Result interface{}, Error string) (module.ProtocolMarshal, string) {
	if app.protocolMarshal != nil {
		return app.protocolMarshal(Trace, Result, Error)
	}
	r := &resultInfo{
		Trace:  Trace,
		Error:  Error,
		Result: Result,
	}
	b, err := json.Marshal(r)
	if err == nil {
		return app.NewProtocolMarshal(b), ""
	} else {
		return nil, err.Error()
	}
}

// ProtocolMarshal
func (app *DefaultApp) NewProtocolMarshal(data []byte) module.ProtocolMarshal {
	return &protocolMarshalImp{
		data: data,
	}
}
