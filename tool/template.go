package main

import (
	"github.com/liangdas/mqant/conf"
	basemodule "github.com/liangdas/mqant/module/base"
)

type ToolServer struct {
	*basemodule.BaseModule
}

// //模块版本
func (t *ToolServer) Version() string {
	return "v1.0.0"
}

//  //模块类型
func (t *ToolServer) GetType() string {
	return "tool"
}

//当App初始化时调用，这个接口不管这个模块是否在这个进程运行都会调用
func (t *ToolServer) OnAppConfigurationLoaded(app *basemodule.BaseModule) {

}

//为以后动态服务发现做准备
func (t *ToolServer) OnConfChanged(settings *conf.ModuleSettings) {

}
func (t *ToolServer) OnInit(app *basemodule.BaseModule, settings *conf.ModuleSettings) {

}
func (t *ToolServer) OnDestroy() {

}
func (t *ToolServer) GetApp() *basemodule.BaseModule {
	return t.BaseModule
}
func (t *ToolServer) Run(closeSig chan bool) {

}
