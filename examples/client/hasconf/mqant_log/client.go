package main

import (
	"fmt"
	"github.com/liangdas/mqant/examples/proto/examples/greeter"
	beegolog "github.com/liangdas/mqant/log/beego"
	"github.com/liangdas/mqant/registry"
	"github.com/liangdas/mqant/registry/consul"
	"github.com/nats-io/nats.go"
	"time"

	"github.com/liangdas/mqant"
	"github.com/liangdas/mqant/conf"
	"github.com/liangdas/mqant/log"
	"github.com/liangdas/mqant/module"
	basemodule "github.com/liangdas/mqant/module/base"
)

func main() {
	// 服务实例
	app := mqant.CreateApp(
		module.Debug(true),
		//module.WithLogger(func() log.Logger {
		//	bee := &Beego{}
		//	return bee
		//}),
		module.ProcessID("developments"),
	)
	// 配置加载
	consulURL := "127.0.0.1:8500"
	// connect to consul
	rs := consul.NewRegistry(func(options *registry.Options) {
		options.Addrs = []string{consulURL}
	})

	natsURL := "nats://127.0.0.1:4222"
	natsUser := ""
	natsPassword := ""
	nc, err := nats.Connect(natsURL, nats.ErrorHandler(nil), nats.UserInfo(natsUser, natsPassword), nats.MaxReconnects(10000))
	if err != nil {
		//panic("no connect nats")
		//return
	}
	_ = app.OnConfigurationLoaded(func(app module.App) {
		_ = app.UpdateOptions(
			module.Nats(nc),     //指定nats rpc
			module.Registry(rs), //指定服务发现
		)
	})
	go func() {
		// wait run
		time.Sleep(3 * time.Second)
		rsp, err := greeter.NewGreeterClient(app, "greeter").Hello(&greeter.Request{})
		if err != nil {
			log.Info("xxxx %s", err)
			return
		}
		log.Info("xxxx %s", rsp)
		return
	}()
	s := &Server{}
	app.Run(s)
}

type Beego struct {
	*beegolog.BeeLogger
}

// Info 输出普通日志
func (b *Beego) Info(format string, keyvals ...interface{}) {
	fmt.Printf(format, keyvals...)
	fmt.Printf("\n")
}

// Debug 输出调试日志
func (b *Beego) Debug(format string, keyvals ...interface{}) {
	fmt.Printf(format, keyvals...)
	fmt.Printf("\n")
}

// Warning 输出警告日志
func (b *Beego) Warning(format string, keyvals ...interface{}) {
	fmt.Printf(format, keyvals...)
	fmt.Printf("\n")
}

// Error 输出错误日志
func (b *Beego) Error(format string, keyvals ...interface{}) {
	fmt.Printf(format, keyvals...)
	fmt.Printf("\n")
}

type Server struct {
	basemodule.BaseModule
	version string
	// 模块名字
	Name string
}

// GetApp module.App
func (m *Server) GetApp() module.App {
	return m.App
}

// OnInit() 初始化配置
func (s *Server) OnInit(app module.App, settings *conf.ModuleSettings) {
	s.BaseModule.OnInit(s, app, settings)
}

// Run() 运行服务
func (s *Server) Run(closeSig chan bool) {
	//创建MongoDB连接实例
}

// 销毁服务
func (s *Server) OnDestroy() {
	//一定别忘了继承
	s.BaseModule.OnDestroy()
	s.GetServer().OnDestroy()
}

// Version() 获取当前服务的代码版本
func (s *Server) Version() string {
	//可以在监控时了解代码版本
	return s.version
}

func (s *Server) GetType() string {
	return "client"
}
