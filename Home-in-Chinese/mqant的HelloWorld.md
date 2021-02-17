# 客户端快速测试
如果你需要测试其他语言的mqtt客户端，可以使用mqant提供的测试接口来测试
### tcp mqtt :
	host: h5link.com
	port: 3563
	protocol=mqtt.MQTTv31
	tcp:  tls/TLSv1
	
	如果客户端需要ca证书可以使用下面这个网站提供的
	https://curl.haxx.se/docs/caextract.html

### websocket mqtt :
	host: wss://www.h5link.com:3653/mqant
	protocol=mqtt.MQTTv31
	
### 测试协议

1. 登陆接口

		向服务器publish一条登陆消息
	
		topic:		Login/HD_Login/{msgid}
		
		message:	{"userName": "liangdas", "passWord": "Hello,anyone!"}
	
	如果topic添加了msgid,则服务器会返回一条回复消息

2. 加入聊天室

		向服务器publish一条登陆消息
	
		topic:		Chat/HD_JoinChat/{msgid}
		
		message:	{"roomName": "mqant"}
	
		服务器会广播消息给所有聊天室成员
		
		topic:		Chat/OnJoin
			
		message:	{"users": [“liangdas”]}

3. 发送一条聊天

		向服务器publish一条登陆消息
	
		topic:		Chat/HD_Say/{msgid}
		
		message:	{"roomName": "mqant","from":"liangdas","target":"*","content": "大家好!!"}
	
		服务器会广播消息给所有聊天室成员
		
		topic:		Chat/OnChat
			
		message:	{"roomName": "mqant","from":"liangdas","target":"*","msg":"大家好!!"}
	

# Demo说明

我们采用以后用mqant实现的全平台聊天服务器来讲解,该Demo的web客户端Html代码参考了pomelo

获取 mqantserver：

	git clone https://github.com/liangdas/mqantserver

设置 mqantserver 目录到 GOPATH 后获取相关依赖：

	go get github.com/astaxie/beego
	go get github.com/gorilla/websocket
	go get github.com/streadway/amqp
	go get github.com/liangdas/mqant

编译 mqantserver：

go install server
如果一切顺利，运行 bin/server 你可以获得以下输出：

	[release] mqant 1.0.0 starting up
	[debug  ] RPCClient create success type(Gate) id(127.0.0.1:Gate)
	[debug  ] RPCClient create success type(Login) id(127.0.0.1:Login)
	[debug  ] RPCClient create success type(Chat) id(127.0.0.1:Chat)
	[release] This service belongs to [development]
	[release] RPCServer init success id(127.0.0.1:Gate)
	[release] RPCServer init success id(127.0.0.1:Login)
	[release] RPCServer init success id(127.0.0.1:Chat)
	[release] WS Listen :%!(EXTRA string=0.0.0.0:3653)
	[release] TCP Listen :%!(EXTRA string=0.0.0.0:3563)

敲击 Ctrl + C 关闭游戏服务器，服务器正常关闭输出：

	[debug  ] RPCServer close success id(127.0.0.1:Chat)
	[debug  ] RPCServer close success id(127.0.0.1:Login)
	[debug  ] RPCServer close success id(127.0.0.1:Gate)
	[debug  ] RPCClient close success type(Gate) id(127.0.0.1:Gate)
	[debug  ] RPCClient close success type(Login) id(127.0.0.1:Login)
	[debug  ] RPCClient close success type(Chat) id(127.0.0.1:Chat)
	[release] mqant closing down (signal: interrupt)

# 启动网页版本客户端
编译 mqantserver：

go install client

如果一切顺利，运行 bin/client

访问地址为：http://127.0.0.1:8080/mqant/index.html


# 启动python版本客户端

执行src/client/mqtt_chat_client.py即可 需要安装paho.mqtt库,请自行百度

# Demo演示说明
	1. 启动服务器
	2. 启动网页客户端	(默认房间名,用户名)
	4. 登陆成功后就可以聊天了


# 项目目录结构
[https://github.com/liangdas/mqantserver](https://github.com/liangdas/mqantserver) 仓库中包含了mqant框架,所用到的第三方库,聊天Demo服务端,聊天代码客户端代码

	bin		
		|-conf/server.conf			服务端配置文件
		|-public						web客户端静态文件
	src
		|-client
			|-mqtt_chat_client.py 	聊天客户端 Python版本
			|-webclient.go			聊天客户端网页版本
		|-github.com                需要执行 go get 命令拉取
			|-astaxie.beego框架 		webclient.go用到了
			|-gorilla.websocket		websocket框架
			|-liangdas.mqant			mqant框架代码
			|-streadway.amqp			rabbitmq通信框架
		|-server						聊天服务器Demo
			|-gate						网关模块
			|-chat						聊天模块
			|-conf						系统配置文件
			|-login						登陆模块
			|-main.go					服务器启动入口