# 启动服务发现组件

1. gnatsd
2. consul agent --dev

mqant服务器的启动非常简单

	app := mqant.CreateApp(
    		module.Debug(true), //是否开启debug模式
    		module.Configure("/work/go/joy-go/bin/conf/server.json"), // 配置
    		module.ProcessID("development"), //模块组ID
    )
	app.Run(
			gate.Module(),	//这是默认网关模块,是必须的支持 TCP,websocket,MQTT协议
			login.Module(), //这是用户登录验证模块
			chat.Module())  //这是聊天模块

app.Run(...)加入所有编写好的模块即可


# mqant服务器启动命令
mqant服务器被编译以后是一个可执行的二进制文件,这里假设名为server

linux运行命令为

server -pid [ProcessID 进程ID] -conf [配置文件路径] -log [日志目录]

pid 	【可选】默认为development

conf  【可选】默认为运行二进制文件目录下的 conf/server.conf

log		【可选】日志文件存储目录

	会分三个日志文件
	pid.nohup.log    标准输入输出
	pid.access.log    用mqant日志模块打印的正常日志
	pid.error.log     用mqant日志模块打印的错误日志

# mqant服务器启动流程
mqant服务器首先会从配置文件中导出配置的module列表的配置参数

然后根据module中ProcessID参数来判断该进程是否初始化这个模块

因此配置文件中module的ProcessID参数最终决定了这个模块会运行在哪一个进程中,这里为以后分布式部署做好铺垫

