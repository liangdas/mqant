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
package app

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/liangdas/mqant/conf"
	"github.com/liangdas/mqant/log"
	"github.com/liangdas/mqant/module"
	"github.com/liangdas/mqant/module/base"
	"github.com/liangdas/mqant/module/modules"
	"github.com/liangdas/mqant/registry"
	"github.com/nats-io/nats.go"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
	//"github.com/liangdas/mqant/registry/etcdv3"
	"context"
	"github.com/liangdas/mqant/rpc"
	"github.com/liangdas/mqant/selector"
	"github.com/liangdas/mqant/selector/cache"
	"github.com/pkg/errors"
)

type resultInfo struct {
	Trace  string
	Error  string      //错误结果 如果为nil表示请求正确
	Result interface{} //结果
}

type protocolMarshalImp struct {
	data []byte
}

func (this *protocolMarshalImp) GetData() []byte {
	return this.data
}

func newOptions(opts ...module.Option) module.Options {
	var wdPath,confPath,Logdir,BIdir string
	var ProcessID="development"
	opt := module.Options{
		Registry:         registry.DefaultRegistry,
		Selector:         cache.NewSelector(),
		RegisterInterval: time.Second * time.Duration(10),
		RegisterTTL:      time.Second * time.Duration(20),
		KillWaitTTL:      time.Second * time.Duration(60),
		Debug:            true,
		Parse:			  true,
	}

	for _, o := range opts {
		o(&opt)
	}

	if opt.Parse{
		wdPath = *flag.String("wd", "", "Server work directory")
		confPath = *flag.String("conf", "", "Server configuration file path")
		ProcessID = *flag.String("pid", "development", "Server ProcessID?")
		Logdir = *flag.String("log", "", "Log file directory?")
		BIdir = *flag.String("bi", "", "bi file directory?")
		flag.Parse() //解析输入的参数
	}

	if opt.Nats == nil {
		nc, err := nats.Connect(nats.DefaultURL)
		if err != nil {
			log.Error("nats agent: %s", err.Error())
			//panic(fmt.Sprintf("nats agent: %s", err.Error()))
		}
		opt.Nats = nc
	}

	if opt.WorkDir == "" {
		opt.WorkDir = wdPath
	}
	if opt.ProcessID == "" {
		opt.ProcessID = ProcessID
	}
	ApplicationDir := ""
	if opt.WorkDir != "" {
		_, err := os.Open(opt.WorkDir)
		if err != nil {
			panic(err)
		}
		os.Chdir(opt.WorkDir)
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
	opt.WorkDir = ApplicationDir
	defaultConfPath := fmt.Sprintf("%s/bin/conf/server.json", ApplicationDir)
	defaultLogPath := fmt.Sprintf("%s/bin/logs", ApplicationDir)
	defaultBIPath := fmt.Sprintf("%s/bin/bi", ApplicationDir)

	if opt.ConfPath == "" {
		if confPath == "" {
			opt.ConfPath = defaultConfPath
		} else {
			opt.ConfPath = confPath
		}
	}

	if opt.LogDir == "" {
		if Logdir == "" {
			opt.LogDir = defaultLogPath
		} else {
			opt.LogDir = Logdir
		}
	}

	if opt.BIDir == "" {
		if BIdir == "" {
			opt.BIDir = defaultBIPath
		} else {
			opt.BIDir = BIdir
		}
	}

	_, err := os.Open(opt.ConfPath)
	if err != nil {
		//文件不存在
		panic(fmt.Sprintf("config path error %v", err))
	}
	_, err = os.Open(opt.LogDir)
	if err != nil {
		//文件不存在
		err := os.Mkdir(opt.LogDir, os.ModePerm) //
		if err != nil {
			fmt.Println(err)
		}
	}

	_, err = os.Open(opt.BIDir)
	if err != nil {
		//文件不存在
		err := os.Mkdir(opt.BIDir, os.ModePerm) //
		if err != nil {
			fmt.Println(err)
		}
	}
	return opt
}

func NewApp(opts ...module.Option) module.App {
	options := newOptions(opts...)
	app := new(DefaultApp)
	app.opts = options
	options.Selector.Init(selector.SetWatcher(app.Watcher))
	app.rpcserializes = map[string]module.RPCSerialize{}
	return app
}

type DefaultApp struct {
	//module.App
	version       string
	settings      conf.Config
	serverList    sync.Map
	opts          module.Options
	defaultRoutes func(app module.App, Type string, hash string) module.ServerSession
	//将一个RPC调用路由到新的路由上
	mapRoute            func(app module.App, route string) string
	rpcserializes       map[string]module.RPCSerialize
	configurationLoaded func(app module.App)
	startup             func(app module.App)
	moduleInited        func(app module.App, module module.Module)
	protocolMarshal     func(Trace string, Result interface{}, Error string) (module.ProtocolMarshal, string)
}

func (app *DefaultApp) Run(mods ...module.Module) error {
	f, err := os.Open(app.opts.ConfPath)
	if err != nil {
		//文件不存在
		panic(fmt.Sprintf("config path error %v", err))
	}
	var cof conf.Config
	fmt.Println("Server configuration path :", app.opts.ConfPath)
	conf.LoadConfig(f.Name()) //加载配置文件
	cof = conf.Conf
	app.Configure(cof) //解析配置信息
	log.InitLog(app.opts.Debug, app.opts.ProcessID, app.opts.LogDir, cof.Log)
	log.InitBI(app.opts.Debug, app.opts.ProcessID, app.opts.BIDir, cof.BI)

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
	manager.Init(app, app.opts.ProcessID)
	if app.startup != nil {
		app.startup(app)
	}
	// close
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM)
	sig := <-c

	//如果一分钟都关不了则强制关闭
	timeout := time.NewTimer(app.opts.KillWaitTTL)
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

func (app *DefaultApp) SetMapRoute(fn func(app module.App, route string) string) error {
	app.mapRoute = fn
	return nil
}

func (app *DefaultApp) AddRPCSerialize(name string, Interface module.RPCSerialize) error {
	if _, ok := app.rpcserializes[name]; ok {
		return fmt.Errorf("The name(%s) has been occupied", name)
	}
	app.rpcserializes[name] = Interface
	return nil
}

func (app *DefaultApp) Options() module.Options {
	return app.opts
}

func (app *DefaultApp) Transport() *nats.Conn {
	return app.opts.Nats
}
func (app *DefaultApp) Registry() registry.Registry {
	return app.opts.Registry
}

func (app *DefaultApp) GetRPCSerialize() map[string]module.RPCSerialize {
	return app.rpcserializes
}

func (app *DefaultApp) Watcher(node *registry.Node) {
	//把注销的服务ServerSession删除掉
	session, ok := app.serverList.Load(node.Id)
	if ok && session != nil {
		session.(module.ServerSession).GetRpc().Done()
		app.serverList.Delete(node.Id)
	}
}

func (app *DefaultApp) Configure(settings conf.Config) error {
	app.settings = settings
	return nil
}

/**
 */
func (app *DefaultApp) OnInit(settings conf.Config) error {

	return nil
}

func (app *DefaultApp) OnDestroy() error {

	return nil
}

func (app *DefaultApp) GetServerById(serverId string) (module.ServerSession, error) {
	session, ok := app.serverList.Load(serverId)
	if !ok {
		serviceName := serverId
		s := strings.Split(serverId, "@")
		if len(s) == 2 {
			serviceName = s[0]
		} else {
			return nil, errors.Errorf("serverId is error %v", serverId)
		}
		sessions := app.GetServersByType(serviceName)
		for _, s := range sessions {
			if s.GetNode().Id == serverId {
				return s, nil
			}
		}
	} else {
		return session.(module.ServerSession), nil
	}
	return nil, errors.Errorf("nofound %v", serverId)
}

func (app *DefaultApp) GetServerBySelector(serviceName string, opts ...selector.SelectOption) (module.ServerSession, error) {
	next, err := app.opts.Selector.Select(serviceName, opts...)
	if err != nil {
		return nil, err
	}
	node, err := next()
	if err != nil {
		return nil, err
	}
	session, ok := app.serverList.Load(node.Id)
	if !ok {
		s, err := basemodule.NewServerSession(app, serviceName, node)
		if err != nil {
			return nil, err
		}
		app.serverList.Store(node.Id, s)
		return s, nil
	}
	session.(module.ServerSession).SetNode(node)
	return session.(module.ServerSession), nil

}

func (app *DefaultApp) GetServersByType(serviceName string) []module.ServerSession {
	sessions := make([]module.ServerSession, 0)
	services, err := app.opts.Selector.GetService(serviceName)
	if err != nil {
		log.Warning("GetServersByType %v", err)
		return sessions
	}
	for _, service := range services {
		//log.TInfo(nil,"GetServersByType3 %v %v",Type,service.Nodes)
		for _, node := range service.Nodes {
			session, ok := app.serverList.Load(node.Id)
			if !ok {
				s, err := basemodule.NewServerSession(app, serviceName, node)
				if err != nil {
					log.Warning("NewServerSession %v", err)
				} else {
					app.serverList.Store(node.Id, s)
					sessions = append(sessions, s)
				}
			} else {
				session.(module.ServerSession).SetNode(node)
				sessions = append(sessions, session.(module.ServerSession))
			}
		}
	}
	return sessions
}

func (app *DefaultApp) GetRouteServer(filter string, opts ...selector.SelectOption) (s module.ServerSession, err error) {
	if app.mapRoute != nil {
		//进行一次路由转换
		filter = app.mapRoute(app, filter)
	}
	sl := strings.Split(filter, "@")
	if len(sl) == 2 {
		moduleID := sl[1]
		if moduleID != "" {
			return app.GetServerById(filter)
		}
	}
	moduleType := sl[0]
	return app.GetServerBySelector(moduleType, opts...)
}

func (app *DefaultApp) GetSettings() conf.Config {
	return app.settings
}
func (app *DefaultApp) GetProcessID() string {
	return app.opts.ProcessID
}
func (app *DefaultApp) WorkDir() string {
	return app.opts.WorkDir
}
func (app *DefaultApp) RpcInvoke(module module.RPCModule, moduleType string, _func string, params ...interface{}) (result interface{}, err string) {
	server, e := app.GetRouteServer(moduleType)
	if e != nil {
		err = e.Error()
		return
	}
	return server.Call(nil, _func, params...)
}

func (app *DefaultApp) RpcInvokeNR(module module.RPCModule, moduleType string, _func string, params ...interface{}) (err error) {
	server, err := app.GetRouteServer(moduleType)
	if err != nil {
		return
	}
	return server.CallNR(_func, params...)
}

func (app *DefaultApp) RpcInvokeArgs(module module.RPCModule, moduleType string, _func string, ArgsType []string, args [][]byte) (result interface{}, err string) {
	server, e := app.GetRouteServer(moduleType)
	if e != nil {
		err = e.Error()
		return
	}
	return server.CallArgs(nil, _func, ArgsType, args)
}

func (app *DefaultApp) RpcInvokeNRArgs(module module.RPCModule, moduleType string, _func string, ArgsType []string, args [][]byte) (err error) {
	server, err := app.GetRouteServer(moduleType)
	if err != nil {
		return
	}
	return server.CallNRArgs(_func, ArgsType, args)
}

func (app *DefaultApp) RpcCall(ctx context.Context, moduleType, _func string, param mqrpc.ParamOption, opts ...selector.SelectOption) (result interface{}, errstr string) {
	server, err := app.GetRouteServer(moduleType, opts...)
	if err != nil {
		errstr = err.Error()
		return
	}
	return server.Call(ctx, _func, param()...)
}

func (app *DefaultApp) GetModuleInited() func(app module.App, module module.Module) {
	return app.moduleInited
}

func (app *DefaultApp) OnConfigurationLoaded(_func func(app module.App)) error {
	app.configurationLoaded = _func
	return nil
}

func (app *DefaultApp) OnModuleInited(_func func(app module.App, module module.Module)) error {
	app.moduleInited = _func
	return nil
}

func (app *DefaultApp) OnStartup(_func func(app module.App)) error {
	app.startup = _func
	return nil
}

func (app *DefaultApp) SetProtocolMarshal(protocolMarshal func(Trace string, Result interface{}, Error string) (module.ProtocolMarshal, string)) error {
	app.protocolMarshal = protocolMarshal
	return nil
}

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

func (app *DefaultApp) NewProtocolMarshal(data []byte) module.ProtocolMarshal {
	return &protocolMarshalImp{
		data: data,
	}
}
