# mqant配置文件
mqant配置文件目前管理以下几个功能的配置信息

1. Master 管理模块配置
2. Module 模块配置
3. Mqtt	网关的mqtt协议配置 【一般不常动】

### 配置文件模板

	{
    "Master":{
        //是否启动Master功能,如果为true,进程会每个三秒向Master模块汇报其所在模块运行信息
        "Enable":false,
        //管理模块的类型,与Module列表中ModuleType对应,开发者可以改为一个随机名称,以防止被非法访问
        "MasterType":"Master",
        //web静态文件路径
        "WebRoot":"/work/go/mqantserver/bin/console",
        //web控制台监听端口
        "WebHost":"0.0.0.0:8686",
        //用于远程服务器SSH的配置,本机IP如(127.0.0.1 localhost)无需配置
        "SSH":[
            {
            "Host":"xxx.xxx.xxx.xxx",
            "Port":22,
            "User":"xxx",
            "Password":"xxx"
            }
        ],
        "Process":[
            {
                "ProcessID":"development",
                "Host":"xxx.xxx.xxx.xxx",
                //执行文件根目录
                "Execfile":".../server",
                //日志文件目录 会生成如下三个日志文件 pid.nohup.log pid.access.log pid.error.log
                "LogDir":".../logs",
                //自定义的参数
                "Args":{
                }
            }
        ]
    },
	"Module":{
        "Gate":[
                {
                    //Id在整个Module中必须唯一,不能重复
                    "Id":"Gate001",
                    //这个模块所属进程,非常重要,进程会根据该参数来判断是否需要运行该模块 [development]为默认值代表开发环境
                    "ProcessID":"development",
                    "Settings":{
                        "WSAddr":      	 ":3653",
                        "TCPAddr":     	 ":3563",
                        "MaxMsgLen":     4096,
                        "HTTPTimeout":   10,
                        "MaxConnNum" :   20000,
                        "Tls"        :   false,
                        "CertFile"       :   "xxx.pem",
                        "KeyFile"        :   "xxx.key",
                        //Session持久化心跳包 单位/秒
                        "MinHBStorage"  :   60
                    }
                }
            ],
        "Master":[
                        {
                            "Id":"Master001",
                            "ProcessID":"development"
                        }
                ],
        "Login":[
                {
                    "Id":"Login001",
                    "ProcessID":"development",
                    "Rabbitmq":{
				    		"Uri"           :"amqp://user:pw@host:port/",
				    		"Exchange"      :"mqant",
				    		"ExchangeType"  :"direct",
				    		"Queue"        	:"Login001",
				    		"BindingKey"    :"Login001",
				    		"ConsumerTag"   :"Login001"
				    		}
                }
        ],
        "Chat":[
                   {
                        "Id":"Chat001",
                        "ProcessID":"development"，
                        "Rabbitmq":{
				    		"Uri"           :"amqp://user:pw@host:port/",
				    		"Exchange"      :"mqant",
				    		"ExchangeType"  :"direct",
				    		"Queue"        	:"Chat001",
				    		"BindingKey"    :"Chat001",
				    		"ConsumerTag"   :"Chat001"
				    		}
                   },
                   {
                        "Id":"Chat002",
                        "ProcessID":"development"，
                        "Rabbitmq":{
				    		"Uri"           :"amqp://user:pw@host:port/",
				    		"Exchange"      :"mqant",
				    		"ExchangeType"  :"direct",
				    		"Queue"        	:"Chat002",
				    		"BindingKey"    :"Chat002",
				    		"ConsumerTag"   :"Chat002"
				    		}
                   }
               ]
	},
	"Mqtt":{
	    // 最大写入包队列缓存
        "WirteLoopChanNum": 10,
        // 最大读取包队列缓存
        "ReadPackLoop": 10,
        // 读取超时
        "ReadTimeout": 60,
        // 写入超时
        "WriteTimeout": 30
	}
	}
### 配置说明
#### Master 配置字段说明

Master 字段用于mqant进程管理模块的配置,主要分为三大块

1.	Master模块提供的Web环境所需要的参数
2. 要管控的远程服务器登陆信息 SSH
3. 要管控的进程信息 Process

##### SSH字段

可以将需要远程访问的服务器登陆信息都列在此次,但本机IP如(127.0.0.1 localhost)无需配置

	{
       "Host":"xxx.xxx.xxx.xxx",
       "Port":22,
       "User":"xxx",
       "Password":"xxx"
    }
##### Process字段

该字段指明要所有要启动的进程,以及启动参数

	{
    	"ProcessID":"development",
    	"Host":"xxx.xxx.xxx.xxx",
    	//执行文件根目录
    	"Execfile":".../server",
    	//日志文件目录 
    	"LogDir":".../logs",
    	//自定义的参数
    	"Args":{
                }
   }

1. ProcessID
	
	进程ID,需要全局唯一

2. Host

	进程所属服务器,与SSH配置中信息相对应
	
3. LogDir

	日志文件目录,进程启动时会生成如下三个日志文件 
	
	pid.nohup.log  进程标准输出打印信息
	
	pid.access.log 使用mqant日志模块打印的正常日志
	
	pid.error.log  使用mqant日志模块打印的异常日志
	
	注意:
	
	日志打印也会根据app.Run()启动时传入的Debug参数配合
	
	1. 当Debug为true时只在控制台打印,不会将日志输出到日志文件中
	2. 当Debug为false时不会再控制台打印,但会将日志输出到日志文件中
	
4. Args 【备用】

	自定义参数,Master控制台启动进程时会带上自定义参数作为启动条件

#### Module 配置字段说明

Module 字段是最重要的一个配置项,其中的子节点字段格式为

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
    		}
   	}
1. Id 

	必须全局唯一,不同类型的模块也不能重复
2. ProcessID 

	决定模块运行在什么类型的进程,与Master.Process配置中的信息对应
	
3. Rabbitmq 

	rpc远程通信的配置,如果不需要远程通信则可以不填
	
	当要部署多个进程的时候该参数就一定要配置了,否则模块之间无法通信
4. Settings 

	模块的自定义参数,例如Gate网关模块就有TCP,websocket通信的相关配置等等

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
	
> 注意事项：
> mqant对每一个Uri创建了一个连接池，每一个进程中多个模块如果共用一个uri的话就会共用这些连接池，默认情况下连接池大小为30，
> 
> 1. mqant每一个模块(Server)会占用一个连接池，因为它需要持续监控redis队列中的信息
> 2. mqant进程初始化时对每一个模块(Server)会创建一个RedisClient，它也会独占一个连接。
> 
> 因此随着模块数量的增加你需要注意连接池的占用情况（必要时需要提升连接池数量），防止导致连接不够用的情况
	
## 模块部署说明
mqant中目前所有功能都是按Module来划分的,mqant中Moudule有以下这些特点

1. 一个类型的Module代码相同
2. 一个类型的Module可以在多个进程中运行
3. 一个进程中,一个类型的模块最多只有一个实例

因此在Module配置中一个类型的模块可以配置多个,但每一个模块只能被分配到不同的进程中（ProcessID不能重复）

![alt mqant模块部署示意图](https://github.com/liangdas/mqant/wiki/images/mqant_module_diagram.png "mqant模块部署示意图")



   	