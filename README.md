# mqant
mqant是一款基于Golang语言的简洁,高效,高性能的分布式游戏服务器框架,研发的初衷是要实现一款能支持高并发,高性能,高实时性,的游戏服务器框架,也希望mqant未来能够做即时通讯和物联网方面的应用

#	特性
1. 分模块机制
2. 基于golang协程,开发过程全程做到无callback回调,代码可读性更高
3. RPC支持本地和远程自动切换
4. 远程RPC默认使用rabbitmq,未来可以添加更多种类的通信协议
5. 网关采用MQTT协议,无需再开发客户端底层库,直接套用已有的MQTT客户端代码库,可以支持IOS,Android,websocket,PC等多平台通信

# 社区
QQ交流群 :463735103

#	文档

 快速上手:
 
 [https://github.com/liangdas/mqantserver](https://github.com/liangdas/mqantserver)
 

#	框架架构
mqant采用的是模块化的架构思想,每一个服务会独立成单独的模块,模块之间采用标准的RPC通讯
	
##mqant rpc

mqant rpc底层使用了本地管道和远程rabbitmq消息队列进行消息传到,rpc会根据实际情况选择适当的通信通道以提升性能,也正是基于此mqant可以支持分布式部署

##网络通信以及MQTT

mqant网络通信目前支持TCP和websocket,通信协议采用的是MQTT,因此mqant可以采用MQTT现有的客户端代码来支持 Android,IOS,Html5,嵌入式设备等等

##网关模块
mqant内置一个网关模块(Gate),这个网关模块负责TCP,websocket连接,MQTT协议解析,消息路由,以及一个简单的Session


##技术交流
目前框架还处于Demo阶段,希望能有更多的大牛来参与框架的完善,一起将框架做的更好用
mqant技术交流群:QQ 463735103

##贡献者

欢迎提供dev分支的pull request

bug请直接通过issue提交

凡提交代码和建议, bug的童鞋, 均会在下列贡献者名单者出现

##版本日志
###2013-3-8  v1.1.0
1. 重写了日志模块

	启动进程时提供一个参数  -log=日志目录  如果不提供就默认选择执行文件目录下的logs目录

	1. 会分三个日志文件
	pid.nohup.log    标准输入输出
	pid.access.log    用mqant日志模块打印的正常日志
	pid.error.log     用mqant日志模块打印的错误日志
	
	2. 进程启动入口Run()新增了一个debug参数
	在debug环境下 只在控制台打印  非debug模式下日志不在控制台打印，只在文件内输出

2. 新增了Master模块(未全部完成)
	
	1. 已实现分布式远程进程的启动,停止,状态查询
	2. 已实现分布式进程中模块运行状态信息的收集(多进程部署时需要配置rabbitmq,否则RPC无法远程通信)
	3. 下一步工作:
		实现在web页面上查看进程状态,模块状态,可以操作进程启动,停止

3. 接口优化

	1.	gate.Session 新增了客户端IP属性
	
		后端模块可以通过这个属性来做IP白名单等功能
	2. app.Run	新增debug属性
	
		用来在启动时指定调试模式/非调试模式
	
	3. module新增Vesion()函数
	
		开发者在模块实现中指定当前模块的版本,在Master中可以实时看到,该功能主要用于在分布式部署环境下对代码版本的管控
		
	4. module配置属性更改
	
		Group参数改成了ProcessID
		当前的分布式部署方案是:
		
		SSH 代表一台服务器
			|- Host		主机IP
		
		Process 代表一个进程
			|- Host		 主机IP
			|- ProcessID  进程ID
			
		module
		   |- Id			  模块ID
		   |- ProcessID   所属进程
		   
		1. 一个进程中一个类型的module只能启动一个
		2. 模块的ProcessID约定了其属于哪一个进程
	5. 新增linux时间轮算法模块
		
		module.Timer可以添加定时器,最小单位是1/毫秒
		
###2013-2-28  v1.0.0

	mqant第一个版本
	
##下一步计划
1. 讨论社区
2. 分布式架构管理模块(Master)
。。。