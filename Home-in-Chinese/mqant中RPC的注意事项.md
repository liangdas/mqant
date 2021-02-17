# mqant的RPC注意事项
mqant的rpc目前已支持四种通道

1. local进程内通信队列(基于chan)
2. rabbitmq
3. redis pub/sub
4. udp  不可靠传输

## 配置文件

local方式是默认添加的，无需配置

rabbitmq和redis都需要在server.conf的每一个模块中单独配置

### Module 配置字段示例


	{
    		"Id":"127.0.0.1:Login",
    		"Host":"127.0.0.1",
    		"ProcessID":"development",
    		"Settings":{
    			//模块自定义配置项
    		},
    		"Rabbitmq":{
    		"Uri"           :"amqp://user:pw@host:port/",
    		"Exchange"      :"xxx",
    		"ExchangeType"  :"direct",
    		"Queue"        	:"xxx",
    		"BindingKey"    :"xxx",
    		"ConsumerTag"   :"xxx"
    		},
    		"Redis":{
         		"Uri"          :"redis://:[password]@[ip]:[port]/[db]",
         		"Queue"        :"Tank001"
    		}
   	}


## Rabbitmq配置说明

Rabbitmq 可以按最简单的方式来配置,如

	"Rabbitmq":{
    		"Uri"           :"amqp://user:pw@host:port/",
    		"Exchange"      :"所有模块可以共用一个Exchange",
    		"ExchangeType"  :"direct",
    		"Queue"        	:"每一个模块,在每一个进程中queue分开",
    		"BindingKey"    :"与queue相同,rpcclient连接服务时使用",
    		"ConsumerTag"   :"随便"
    }

	详细可见上文配置文件模板


Rabbitmq 可以按最简单的方式来配置,如

	"Rabbitmq":{
    		"Uri"           :"amqp://user:pw@host:port/",
    		"Exchange"      :"所有模块可以共用一个Exchange",
    		"ExchangeType"  :"direct",
    		"Queue"        	:"每一个模块,在每一个进程中queue分开",
    		"BindingKey"    :"与queue相同,rpcclient连接服务时使用",
    		"ConsumerTag"   :"随便"
    }

	详细可见上文配置文件模板
	
## Redis配置说明
	"Redis":{
         "Uri"          :"redis://:[password]@[ip]:[port]/[db]",
         "Queue"        :"Tank001"
    }
	
	Queue需要保证全局唯一
	
## UDP配置说明

	"UDP":{
              "Uri"          :"127.0.0.1:7080",
              "Port"         :7080
            }
	
> 注意事项：
> mqant对每一个Uri创建了一个连接池，每一个进程中多个模块如果共用一个uri的话就会共用这些连接池，默认情况下连接池大小为30，
> 
> 1. mqant每一个模块(Server)会占用一个连接池，因为它需要持续监控redis队列中的信息
> 2. mqant进程初始化时对每一个模块(Server)会创建一个RedisClient，它也会独占一个连接。
> 
> 因此随着模块数量的增加你需要注意连接池的占用情况（必要时需要提升连接池数量），防止导致连接不够用的情况

## RPC通信是选用原则

 按优先级选择可用的通信通道
 
 1. local
 2. rabbitmq
 3. redis

 >**注意：** 如果选择的通道通信不成功,并不会选用下一个通道，而是直接返回错误

## 不可靠RPC通信(UDP)

mqant新增提供UDP的rpc通信机制,由于udp是不可靠消息传输有丢包的可能(内网丢包率应该很低),因此udp作为mqant的一种附加RPC通信机制。

> 注意udp通道的RPC默认允许传递数据长度大小为3k，如需更大的传输范围可以在配置文件中配置

#### 为什么提供一种不可靠的rpc机制？

在游戏开发中有一部分消息是不需要完全可达的,允许一定的丢包率出现，例如给所有网关的玩家广播玩家登陆消息等

因此mqant提供两种级别的通道，开发者可以根据*消息重要程度*合理选择RPC通道

   	