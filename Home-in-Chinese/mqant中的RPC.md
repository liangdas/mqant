## 说明

mqant RPC本身是一个相对独立的功能,RPC有以下的几个特点:

1. 目前支持nats作为服务发现通道，理论上可以扩展其他通信方式
2. 支持服务注册发现,是一个相对完善的微服务框架

## 在模块中使用RPC

module.BaseModule中已经集成了RPC,使用方式如下

### 服务提供者

	//注册服务函数
	module.GetServer().RegisterGO(_func string, fn interface{})
	//注册服务函数
	module.GetServer().Register(_func string, fn interface{})

> RegisterGO与Register的区别是前者为每一条消息创建一个单独的协程来处理,后者注册的函数共用一个协程来处理所有消息,具体使用哪一种方式可以根据实际情况来定,但Register方式的函数请一定注意不要执行耗时功能,以免引起消息阻塞

### 服务调用者

在模块中调用其他模块

result,err:=module.RpcInvoke(moduleType,handler,参数列表...)

result 是调用成功的返回值  
err		是调用失败的失败原因

//不需要回复的调用

err:=module.RpcInvokeNR(moduleType,handler,参数列表...)

>golang支持协程,因此RPC调用摆脱了callback的束缚,虽然RPC是异步调用,但是我们可以按同步调用来编写代码。



## RPC路由规则
mqant 每一类模块可以部署到多台服务器中,因此需要一个nodeId对同一类模块进行区分,nodeId通过服务发现模块在服务启动时自动生成

module.RpcInvoke(moduleType string, _func string, params ...interface{})

### 模块RPC调用规则：

1. 通过RpcCall函数调度（推荐）

        /*
        通用RPC调度函数
        ctx 		context.Context 			上下文,可以设置这次请求的超时时间
        moduleType	string 						服务名称 serverId 或 serverId@nodeId
        _func		string						需要调度的服务方法
        param 		mqrpc.ParamOption			方法传参
        opts ...selector.SelectOption			服务发现模块过滤，可以用来选择调用哪个服务节点
         */
        RpcCall(ctx context.Context, moduleType, _func string, param mqrpc.ParamOption, opts ...selector.SelectOption) (interface{}, string)

    + 支持设置调用超时时间
    + 支持自定义的服务节点选择

2. 指定唯一的moduleId,可以找到唯一的模块服务

	app.GetServersById(moduleId string)

3. 指定模块类型，可以找到模块服务的列表

	app.GetServersByType(Type string) []*module.ServerSession
	
4. BaseModule封装了RPC调用

	module.RpcInvoke(moduleType string, _func string, params ...interface{})
	
	moduleType   [moduleType@moduleID]

	会使用moduleID来作为 hash值


## RPC返回结果断言

mqant参考redis封装了几个RPC返回类型断言,方便开发者使用


1. 自定义结构断言

        r:=new(rsp)
        err:=mqrpc.Marshal(r, func() (reply interface{}, errstr interface{}) {
            return self.RpcInvoke("webapp","sendMessage",&req{id:"hello 我是RpcInvoke"})
        })

        log.Info("RpcInvoke %v , err %v",r.id,err)

2. 字符串断言

        rstr,err:=mqrpc.String(self.RpcInvoke("xxx","sayhello","hello 我是RpcInvoke"}))

        log.Info("RpcInvoke %v , err %v",rstr,err)

3. 其他类型断言

        int

        bool

        map\[string\]string

        ....

## RPC的分类

在mqant中rpc调用可分为两类：

### 前端服务(frontend) :

由客户端发起请求,Gate网关模块负责路由和中转,这类请求参数类型相对固定

	func handler(session map[string]interface{},msg map[string]interface{})(result interface{},err string)
	
	result 调用成功的返回值(可选数值类型见上文)
	err		调用失败时的异常说明
1. session :

	Gate网关服务器中维护的用户信息[详情见Gate网关模块](https://github.com/liangdas/mqant/wiki/Gate%E7%BD%91%E5%85%B3%E6%A8%A1%E5%9D%97)

2. msg		:

	客户端发送的请求参数,一般是一个json数据



### 后端服务(backend) :
	
	后端业务模块间RPC调用,无规范,开发者可以自己定义参数类型


## 如何监听

	1. 实现以下接口
	type RPCListener interface {
			/**
			BeforeHandle会对请求做一些前置处理，如：检查当前玩家是否已登录，打印统计日志等。
			@session  可能为nil
			return error  当error不为nil时将直接返回改错误信息而不会再执行后续调用
			 */
			BeforeHandle(fn string,session gate.Session, callInfo *CallInfo)error
			OnTimeOut(fn string, Expired int64)
			OnError(fn string, callInfo *CallInfo, err error)
			/**
			fn 		方法名
			params		参数
			result		执行结果
			exec_time 	方法执行时间 单位为 Nano 纳秒  1000000纳秒等于1毫秒
			*/
			OnComplete(fn string, callInfo *CallInfo, result *rpcpb.ResultInfo, exec_time int64)
		}

	
	2. 在module中注册该接口
	
	module.SetListener(m.listener)
	

	
	