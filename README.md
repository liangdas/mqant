# mqant
mqant是一款基于Golang语言的简洁,高效,高性能的分布式游戏服务器框架,研发的初衷是要实现一款能支持高并发,高性能,高实时性,的游戏服务器框架,也希望mqant未来能够做即时通讯和物联网方面的应用

# 为什么要用golang

[Node、PHP、Java 和 Go 服务端 I/O 性能PK](http://blog.csdn.net/listen2you/article/details/72935679)


#	特性
1. 高性能分布式
2. 支持分布式跟踪系统接口[传送门](http://bigbully.github.io/Dapper-translation/)
3. 基于golang协程,开发过程全程做到无callback回调,代码可读性更高
4. RPC支持本地和远程自动切换
5. 远程RPC使用redis,rabbitmq,udp作为通道,未来可以添加更多种类的通信协议
6. 网关采用MQTT协议,无需再开发客户端底层库,直接套用已有的MQTT客户端代码库,可以支持IOS,Android,websocket,PC等多平台通信
7. 默认支持mqtt协议,同时网关也支持开发者自定义的粘包协议

# 社区
QQ交流群 :463735103

技术交流社区:[www.mqant.com](http://www.mqant.com)

# 模块

> 将不断加入更多的模块

[mqant组件库](https://github.com/liangdas/mqant-modules)

        短信验证码
        房间模块

[压力测试工具:armyant](https://github.com/liangdas/armyant)

# 社区贡献的库
 [mqant-docker](https://github.com/bjfumac/mqant-docker)
 [MQTT-Laya](https://github.com/bjfumac/MQTT-Laya)

# 依赖项目

	go get github.com/gorilla/mux
	go get github.com/gorilla/websocket
	go get github.com/streadway/amqp
	go get github.com/golang/protobuf
	go get github.com/golang/net/context
	go get github.com/gomodule/redigo
	go get github.com/Jeffail/tunny

#	文档

 快速上手:
 
 [mqant wiki](https://github.com/liangdas/mqant/wiki)
 
# 概述
 
1. [mqant的设计动机](https://github.com/liangdas/mqant/wiki/mqant%E7%9A%84%E8%AE%BE%E8%AE%A1%E5%8A%A8%E6%9C%BA)
2. [mqant框架介绍](https://github.com/liangdas/mqant/wiki/%E%AC%A2%E8%BF%8E%E4%BD%BF%E7%94%A8mqant)
3. [框架架构概述](https://github.com/liangdas/mqant/wiki/mqant%E6%A1%86%E6%9E%B6%E6%A6%82%E8%BF%B0)
4. [通信协议与客户端支持](https://github.com/liangdas/mqant/wiki/%E9%80%9A%E4%BF%A1%E5%8D%8F%E8%AE%AE%E4%B8%8E%E5%AE%A2%E6%88%B7%E7%AB%AF%E6%94%AF%E6%8C%81%E4%BB%8B%E7%BB%8D)
5. 
...

# 演示示例
	mqant 项目只包含mqant的代码文件
	mqantserver 项目包括了完整的测试demo代码和mqant所依赖的库
	如果你是新手可以优先下载mqantserver项目进行试验
	
 
 [在线Demo演示](http://www.mqant.com/mqant/chat/) 【[源码下载](https://github.com/liangdas/mqantserver)】
 
 [多人对战吃小球游戏（绿色小球是在线玩家,点击屏幕任意位置移动小球,可以同时开两个浏览器测试,支持移动端）](http://www.mqant.com/mqant/hitball/)【[源码下载](https://github.com/liangdas/mqantserver)】

 

 

#	框架架构
[mqant的设计动机](https://github.com/liangdas/mqant/wiki/mqant%E7%9A%84%E8%AE%BE%E8%AE%A1%E5%8A%A8%E6%9C%BA)

[框架架构](https://github.com/liangdas/mqant/wiki/mqant%E6%A1%86%E6%9E%B6%E6%A6%82%E8%BF%B0)



## 下一步计划
1. 分布式架构管理模块(Master)
	1. 模块发现
	2. 模块管理
		1. 模块动态添加删除
		2. 模块状态监控
2.  新增英文版文档
    希望有兴趣的英语好的同学能参与帮忙编写英文版本的文档
3. 【已完成】异常日志监控和汇报
	1. 异常日志分类汇总
	2. 定时将异常日志发送到Email
	3. 定时将异常日志通过webhook发送到团队协作工具中(钉钉,worktile等)
4. 【已完成】rpc添加track分布式跟踪系统的接口[Appdash，用Go实现的分布式系统跟踪神器](http://tonybai.com/2015/06/17/appdash-distributed-systems-tracing-in-go/)

## 贡献者

欢迎提供dev分支的pull request

bug请直接通过issue提交

凡提交代码和建议, bug的童鞋, 均会在下列贡献者名单者出现

1. [xlionet](https://github.com/xlionet)
2. [lulucas](https://github.com/lulucas/mqant-UnityExample)
3. [c2matrix](https://github.com/c2matrix)
4. [bjfumac【mqant-docker】[MQTT-Laya]](https://github.com/bjfumac)
5. [jarekzha 【jarekzha-master】](https://github.com/jarekzha)


## 打赏作者

![alt mqant作者打赏码](https://github.com/liangdas/mqant/wiki/images/donation.png)

## 版本日志

### [v1.7.0新特性](https://github.com/liangdas/mqant/wiki/v1.7.0)

### [v1.6.6新特性](https://github.com/liangdas/mqant/wiki/v1.6.6)

### [v1.6.5新特性](https://github.com/liangdas/mqant/wiki/v1.6.5)

### [v1.6.4新特性](https://github.com/liangdas/mqant/wiki/v1.6.4)

### [v1.6.3新特性](https://github.com/liangdas/mqant/wiki/v1.6.3)

### [v1.6.2新特性](https://github.com/liangdas/mqant/wiki/v1.6.2)

### [v1.6.1新特性](https://github.com/liangdas/mqant/wiki/v1.6.1)

### [v1.6.0新特性](https://github.com/liangdas/mqant/wiki/v1.6.0)

### [v1.5.0新特性](https://github.com/liangdas/mqant/wiki/v1.5.0)

### [v1.4.0新特性](https://github.com/liangdas/mqant/wiki/v1.4.0)

### [v1.3.0新特性](https://github.com/liangdas/mqant/wiki/v1.3.0)

### [v1.2.0新特性](https://github.com/liangdas/mqant/wiki/v1.2.0)

### [v1.1.0新特性](https://github.com/liangdas/mqant/wiki/v1.1.0)

		
### v1.0.0

	mqant第一个版本
	
