## mqant如何分布式部署

mqant中的分布式即是将系统中的部分模块分布到不同的进程中,分布到不同进程的模块还能像在同一个进程中一样可以相互调用。

### 实现技术点

1. 在配置文件中为需要分离的模块指定ProcessID
2. 在配置文件中为需要分离的模块指定RPC通信通道(Redis/Rabbitmq)
3. 启动这些进程

经过以上三部就实现了分布式部署

### 配置文件中如何配置？

> 需要配置ProcessID,RPC通信通道

示例:

	"Chat":[
                   {
                        "Id":"Chat001",
                        "ProcessID":"Chat001",
                         "Redis":{
                            "Uri"          :"redis://:user@ip:port/db",
                            "Queue"        :"Chat001"
                        }
                   },
                  {
                    "Id":"Chat002",
                    "ProcessID":"Chat002",
                    "Redis":{
                      "Uri"          :"redis://:user@ip:port/db",
                      "Queue"        :"Chat002"
                    }
                  }
    ],
    
示例说明:

Chat是模块名，代表一个聊天功能模块

###### 配置ProcessID

将Chat分布为两个进程中运行, ProcessID分别为 Chat001, Chat002

###### 配置RPC通信通道

	"Redis":{
    	"Uri"          :"redis://:payplug@115.28.62.110:6379/10",
		"Queue"        :"Chat002"
   	}

	Uri和Queue 共同决定一条RPC通道
	

### 启动进程
>以上已经在配置层面上将Chat分布成两个进程了,但还需最后一个步骤，启动这些进程,Chat才能提供服务

###### 启动进程Chat001
/opt/go/server -pid=Chat001 -conf=/opt/go/bin/conf/server.json -wd=/opt/go/ 

###### 启动进程Chat002
/opt/go/server -pid=Chat002 -conf=/opt/go/bin/conf/server.json -wd=/opt/go/ 

	pid: ProcessID 与配置文件中ProcessID对应
	conf: 配置文件路径,建议使用绝对路径
	wd: 工作目录

###### 正式环境部署可以使用nohup
nohup /opt/go/server -pid=Chat001 -conf=/opt/go/bin/conf/server.json >/opt/go/bin/logs/nohup.log 2>&1 &

### 自动化部署
mqant自动化部署模块目前还未实现