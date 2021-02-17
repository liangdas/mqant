# 概述

因为mqant2x跟1x在接口上保持了几乎完全的兼容,因此已开发的程序几乎<font color=red>不用改动</font>即可

## 配置差异

> 在1x版本中模块的所有进程都需要明确配置到配置文件中,相当于静态注册,而在2x已支持服务发现注册功能的情况下显然不再需要了

1. 1x版本模块配置

        "Module":{
                "Gate":[
                        {
                            //Id在整个Module中必须唯一,不能重复
                            "Id":"模块ID001",
                            //这个模块所属进程,非常重要,进程会根据该参数来判断是否需要运行该模块 [development]为默认值代表开发环境
                            "ProcessID":"development",
                        },
                        {
                            //Id在整个Module中必须唯一,不能重复
                            "Id":"模块ID002",
                            //这个模块所属进程,非常重要,进程会根据该参数来判断是否需要运行该模块 [development]为默认值代表开发环境
                            "ProcessID":"development",
                        }
                    ],


1. 2x版本模块配置

        "Module":{
                "Gate":[
                         //配置一个模块即可,模块Id会动态注册
                        {
                            "Id":"模块ID001",
                            //这个模块所属进程,非常重要,进程会根据该参数来判断是否需要运行该模块 [development]为默认值代表开发环境
                            "ProcessID":"development",
                        }
                    ],


# 初始化函数调用时差异

> mqant的服务发现相关配置不在配置文件中指定,而是在mqant.CreateApp函数传入,这一点参考了go-micro的设计理念,更加灵活


    rs:=registry.DefaultRegistry //etcdv3.NewRegistry()
	nc, err := nats.Connect(nats.DefaultURL,nats.MaxReconnects(10000))
	if err != nil {

	}
	app := mqant.CreateApp(
		module.Nats(nc),        //指定nats rpc
		module.Registry(rs),    //指定服务发现
	)

### 默认值
如若不传,mqant默认使用nats和consul的本机实例,你只需要在本机执行以下命令后即可实现mqant服务发现功能

gnatsd

consul agent --dev



