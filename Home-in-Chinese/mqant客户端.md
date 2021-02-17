# mqant客户端支持

mqant gate网关支持mqtt协议,因此客户端可以直接用mqtt客户端库

https://github.com/eclipse/paho.mqtt.python

[各种语言的mqtt客户端开发库收集(非常全)](https://github.com/mqtt/mqtt.github.io/wiki/libraries)

# mqant客户端库

## JavaScript/websocket

[https://github.com/liangdas/mqantserver/blob/master/bin/chat/js/lib/mqttws31.js](https://github.com/liangdas/mqantserver/blob/master/bin/chat/js/lib/mqttws31.js)

[https://github.com/liangdas/mqantserver/blob/master/bin/chat/js/lib/mqantlib.js](https://github.com/liangdas/mqantserver/blob/master/bin/chat/js/lib/mqantlib.js)

## Unity

> 注意：由于测试服务器已换成mqant.com 因此客户端代码中也请修改一下请求地主，并且请求类型由wss改为ws

[https://github.com/lulucas/mqant-UnityExample](https://github.com/lulucas/mqant-UnityExample)

## python

[https://github.com/liangdas/mqantserver/blob/master/src/client/mqtt_chat_client.py](https://github.com/liangdas/mqantserver/blob/master/src/client/mqtt_chat_client.py)

# 客户端快速测试
如果你需要测试其他语言的mqtt客户端，可以使用mqant提供的测试接口来测试,同时欢迎将提交代码到mqant
### tcp mqtt :
	host: mqant.com
	port: 3563
	protocol=mqtt.MQTTv31
	tcp:  tls/TLSv1
	
	如果客户端需要ca证书可以使用下面这个网站提供的
	https://curl.haxx.se/docs/caextract.html

### websocket mqtt :
	host: ws://www.mqant.com:3653/mqant
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

