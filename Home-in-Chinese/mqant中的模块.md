mqant服务器的功能是以模块为单位的

# 模块定义
	type Module interface {
		Version()(string)	//模块代码的版本
		GetType()(string)	//模块类型
		OnInit(app App,settings *conf.ModuleSettings)
		OnDestroy()
		Run(closeSig chan bool)
	}
	
只要结构体实现了以上接口就可以被框架认为是一个模块

## 模块说明
mqant中目前所有功能都是按Module来划分的,mqant中Moudule有以下这些特点

1. 一个类型的Module代码相同
2. 一个类型的Module可以在多个进程中运行
3. 一个进程中,一个类型的模块最多只有一个实例

因此在Module配置中一个类型的模块可以配置多个,但每一个模块只能被分配到不同的进程中（ProcessID不能重复）

![alt mqant模块部署示意图](https://github.com/liangdas/mqant/wiki/images/mqant_module_diagram.png "mqant模块部署示意图")

# 创建自己的模块
mqant已经为我们实现了基础的模块(module.BaseModule),这个模块已经提供了RPC以及对其他模块的RPC调用接口

在实际使用中,我们应该组合mqant提供的基础模块(module.BaseModule)类似面向对象语言中的继承吧。

> 注意 : gate网关模块继承gate.Gate而不是module.BaseModule

例如 mqantserver中的用户登录模块部分代码:

	var Module = func() (module.Module){
		gate := new(Login)
		return gate
	}
	
	type Login struct {
		module.BaseModule
	}
	func (m *Login) GetType()(string){
		//很关键,需要与配置文件中的Module配置对应
		return "Login"
	}
	
	func (m *Login) OnInit(app module.App,settings *conf.ModuleSettings) {
		m.BaseModule.OnInit(m,app,settings)
	
		m.GetServer().RegisterGO("HD_Login",m.login) //我们约定所有对客户端的请求都以Handler_开头
		m.GetServer().RegisterGO("getRand",m.getRand) //演示后台模块间的rpc调用
	}
	。。。
	
m.BaseModule.OnInit(m,app,settings) 代表对父模块的初始化
m.GetServer().RegisterGO(。。。) 	该模块注册提供远程调用的Handler

# RPC使用
mqant RPC本身是一个相对独立的功能,关于RPC的配置这里不做说明,只说在模块中RPC的使用

### 服务提供者

	//注册服务函数
	RegisterGO(_func string, fn interface{})
	//注册服务函数
	Register(_func string, fn interface{})

> RegisterGO与Register的区别是前者为每一条消息创建一个单独的协程来处理,后者注册的函数共用一个协程来处理所有消息,具体使用哪一种方式可以根据实际情况来定,但Register方式的函数请一定注意不要执行耗时功能,以免引起消息阻塞

### 服务调用者
在模块中调用其他模块

r,err:=module.RpcInvoke(moduleType,handler,参数列表...)

//不需要回复的调用
err:=module.RpcInvokeNR(moduleType,handler,参数列表...)

>因为golang支持协程的RPC调用摆脱了callback的束缚,虽然RPC是异步调用,但是我们可以按同步调用来编写代码。

# 注册模块到服务器中

这里不得不提mqant服务器的启动了,mqant的启动非常简单

	conf.LoadConfig(f.Name()) //加载配置文件
	app:=mqant.CreateApp()
	app.Configure(conf.Conf)  //配置信息
	//app.Route("Chat",ChatRoute)
	app.Run(gate.Module(),	//这是默认网关模块,是必须的支持 TCP,websocket,MQTT协议
			login.Module(), //这是用户登录验证模块
			chat.Module())  //这是聊天模块
			
将所有实现的模块假如Run(...)参数列表中即可