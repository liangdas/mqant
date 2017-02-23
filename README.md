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

##下一步计划
1. 插件系统
2. 日志模块
3. 讨论社区
4. 分布式架构管理工具
5. 可伸缩服务器部署
。。。

##技术交流
目前框架还处于Demo阶段,希望能有更多的大牛来参与框架的完善,一起将框架做的更好用
mqant技术交流群:QQ 463735103

##贡献者

欢迎提供dev分支的pull request

bug请直接通过issue提交

凡提交代码和建议, bug的童鞋, 均会在下列贡献者名单者出现