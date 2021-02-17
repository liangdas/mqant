# 安装mqant
以下教程已默认您已具备了golang语言开发的基础知识

建议直接下载[mqantserver](https://github.com/liangdas/mqantserver)项目，mqantserver是一个完整的mqant测试工程,可以直接编译运行

## 创建项目

	设置【项目】目录到 GOPATH 后获取相关依赖：
	go get github.com/gorilla/mux
	go get github.com/gorilla/websocket
	go get github.com/streadway/amqp
	go get github.com/golang/protobuf
	go get github.com/golang/net/context
	go get github.com/opentracing/basictracer-go
	go get github.com/opentracing/opentracing-go
	
执行完以后项目目录下将生成以下目录


	src
  	|-github.com
  		|-liangdas/mqant
  		|-gorilla/websocket
  		|-streadway/amqp

创建您的开发目录[server]

	src
  	|-github.com
  		|-liangdas/mqant
  		|-gorilla/websocket
  		|-streadway/amqp
  	|-server
  	
编码以后go install server 

如果一切顺利，将生成 bin/server：

	bin
		|-server
	src
	  	|-github.com
	  		|-liangdas/mqant
	  		|-gorilla/websocket
	  		|-streadway/amqp
	  	|-server
	  	

在【bin/conf】下创建服务器配置文件,可参考配置文件模板

	bin
		|-conf
			|-server.conf
		|-server
	src
	  	|-github.com
	  		|-liangdas/mqant
	  		|-gorilla/websocket
	  		|-streadway/amqp
	  	|-server




以前完成以后您可以执行【bin/server】文件可获得输出

	[release] mqant 1.0.0 starting up
	[release] This service belongs to [development]
	[release] RPCServer init success id(127.0.0.1:Gate)
	[release] WS Listen :%!(EXTRA string=0.0.0.0:3653)
	[release] TCP Listen :%!(EXTRA string=0.0.0.0:3563)



	

	
