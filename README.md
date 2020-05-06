# 因为golang 1.13以后go mod强制要求使用major版本规则，mqant目前还不需要，因此只能先将版本号回退为v1.x


# mqant
mqant是一款基于Golang语言的简洁,高效,高性能的分布式微服务游戏服务器框架,研发的初衷是要实现一款能支持高并发,高性能,高实时性,的游戏服务器框架,也希望mqant未来能够做即时通讯和物联网方面的应用

#	特性
1. 高性能分布式
2. 支持分布式服务注册发现,是一款功能完整的微服务框架
3. 基于golang协程,开发过程全程做到无callback回调,代码可读性更高
4. 远程RPC使用nats作为通道
5. 网关采用MQTT协议,无需再开发客户端底层库,直接套用已有的MQTT客户端代码库,可以支持IOS,Android,websocket,PC等多平台通信
6. 默认支持mqtt协议,同时网关也支持开发者自定义的粘包协议

# 文档

[在线文档](https://liangdas.github.io/mqant/)
[在线文档-访问不了用这个](http://docs.mqant.com/)

# 模块

> 将不断加入更多的模块

[mqant组件库](https://github.com/liangdas/mqant-modules)

        短信验证码
        房间模块

[压力测试工具:armyant](https://github.com/liangdas/armyant)

# 社区贡献的库
 [mqant-docker](https://github.com/bjfumac/mqant-docker)
 [MQTT-Laya](https://github.com/bjfumac/MQTT-Laya)


# 演示示例
	mqant 项目只包含mqant的代码文件
	mqantserver 项目包括了完整的测试demo代码和mqant所依赖的库
	如果你是新手可以优先下载mqantserver项目进行试验
	
 
 [在线Demo演示](http://www.mqant.com/mqant/chat/) 【[源码下载](https://github.com/liangdas/mqantserver)】
 
 [多人对战吃小球游戏（绿色小球是在线玩家,点击屏幕任意位置移动小球,可以同时开两个浏览器测试,支持移动端）](http://www.mqant.com/mqant/hitball/)【[源码下载](https://github.com/liangdas/mqantserver)】


# 贡献者

欢迎提供dev分支的pull request

bug请直接通过issue提交

凡提交代码和建议, bug的童鞋, 均会在下列贡献者名单者出现

1. [xlionet](https://github.com/xlionet)
2. [lulucas](https://github.com/lulucas/mqant-UnityExample)
3. [c2matrix](https://github.com/c2matrix)
4. [bjfumac【mqant-docker】[MQTT-Laya]](https://github.com/bjfumac)
5. [jarekzha 【jarekzha-master】](https://github.com/jarekzha)


## 打赏作者

![](http://docs.mqant.com/images/donation.png)

