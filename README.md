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

技术交流社区:[www.mqant.com](http://www.mqant.com)

#	文档

 快速上手:
 
 [mqant wiki](https://github.com/liangdas/mqant/wiki)
 
#概述
 
1. [mqant的设计动机](https://github.com/liangdas/mqant/wiki/mqant%E7%9A%84%E8%AE%BE%E8%AE%A1%E5%8A%A8%E6%9C%BA)
2. [mqant框架介绍](https://github.com/liangdas/mqant/wiki/%E%AC%A2%E8%BF%8E%E4%BD%BF%E7%94%A8mqant)
3. [框架架构概述](https://github.com/liangdas/mqant/wiki/mqant%E6%A1%86%E6%9E%B6%E6%A6%82%E8%BF%B0)
4. [通信协议与客户端支持](https://github.com/liangdas/mqant/wiki/%E9%80%9A%E4%BF%A1%E5%8D%8F%E8%AE%AE%E4%B8%8E%E5%AE%A2%E6%88%B7%E7%AB%AF%E6%94%AF%E6%8C%81%E4%BB%8B%E7%BB%8D)
5. 
...

#演示示例

 [全平台聊天Demo](https://github.com/liangdas/mqantserver)
 
 [在线Demo演示（随时可能关闭）](https://www.h5link.com/mqant/index.html)
 
 [多人对战吃小球游戏（绿色小球是在线玩家,可以同时开两个浏览器测试,支持移动端）](https://www.h5link.com/hitball/index.html)
 
 
 

#	框架架构
[mqant的设计动机](https://github.com/liangdas/mqant/wiki/mqant%E7%9A%84%E8%AE%BE%E8%AE%A1%E5%8A%A8%E6%9C%BA)

[框架架构](https://github.com/liangdas/mqant/wiki/mqant%E6%A1%86%E6%9E%B6%E6%A6%82%E8%BF%B0)

##下一步计划
1. 增加grpc作为一种rpc通信管道
2. 分布式架构管理模块(Master)
3. 异常日志监控和汇报
	1. 异常日志分类汇总
	2. 定时将异常日志发送到Email
	3. 定时将异常日志通过webhook发送到团队协作工具中(钉钉,worktile等)
4. 做一下多进程内存共享方面的实验性研究
5. 。。。

##贡献者

欢迎提供dev分支的pull request

bug请直接通过issue提交

凡提交代码和建议, bug的童鞋, 均会在下列贡献者名单者出现

1. [xlionet](https://github.com/xlionet)
2. [lulucas](https://github.com/lulucas/mqant-UnityExample)



##版本日志
###[v1.3.0新特性](https://github.com/liangdas/mqant/wiki/v1.3.0)

###[v1.2.0新特性](https://github.com/liangdas/mqant/wiki/v1.2.0)

###[v1.1.0新特性](https://github.com/liangdas/mqant/wiki/v1.1.0)

		
###v1.0.0

	mqant第一个版本
	
