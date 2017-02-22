# mqant
mqant是一款基于Golang语言的高性能分布式游戏服务器框架,研发的初衷是要实现一款能支持高并发,高性能,高实时性,的游戏服务器框架
	
#	框架架构
mqant采用的是模块化的架构思想,每一个服务会独立成单独的模块,模块之间采用标准的RPC通讯
	
##mqant rpc

mqant rpc底层使用了本地管道和远程rabbitmq消息队列进行消息传到,rpc会根据实际情况选择适当的通信通道以提升性能,也正是基于此mqant可以支持分布式部署

##网络通信以及MQTT

mqant网络通信目前支持TCP和websocket,通信协议采用的是MQTT,因此mqant可以采用MQTT现有的客户端代码来支持 Android,IOS,Html5,嵌入式设备等等

##网关模块
mqant内置一个网关模块(Gate),这个网关模块负责TCP,websocket连接,MQTT协议解析,消息路由,以及一个简单的Session

