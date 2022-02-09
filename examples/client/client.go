package main

import (
	"fmt"

	"github.com/liangdas/mqant"
	"github.com/liangdas/mqant/conf"
	"github.com/liangdas/mqant/examples/proto/examples/greeter"
	"github.com/liangdas/mqant/log"
	"github.com/liangdas/mqant/module"
	basemodule "github.com/liangdas/mqant/module/base"
)

func main() {
	// 服务实例
	app := mqant.CreateApp(
		module.Debug(false),
		module.WithLogFile(func(logdir, prefix, processID, suffix string) string {
			return fmt.Sprintf("%s/%v%s%s%s", logdir, prefix, processID, "xxx", suffix)
		}),
	)
	// 配置加载
	app.OnConfigurationLoaded(func(app module.App) {
	})
	// 调用hello方法
	rsp, err := greeter.NewGreeterClient(app, "greeter").Hello(&greeter.Request{})
	if err != nil {
		log.Info("xxxx %s", err)
	}
	log.Info("xxxx %s", rsp.Msg)
	s := &Server{}
	app.Run(s)
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
