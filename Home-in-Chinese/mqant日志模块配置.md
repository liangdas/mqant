# mqant日志模块配置

mqant v1.6.3 版本开始使用beego/logs日志模块

### 特性

#### 输出引擎

	支持的引擎有 file、console、net、smtp、dingtalk(钉钉) 、es(ElasticSearch)、jianliao(简聊)、slack
	
#### 文件输出

	1. 按照每天输出文件
	2. 可限制每个文件最大写入行
	3. 可限制每个文件最大文件大小
	4. error，access类日志分文件输出

### 使用方法

#### 配置方式

mqant的日志配置选项基本与beego的日志配置字段保持一致,可参考

[beego日志配置文档](https://beego.me/docs/module/logs.md)

在mqant的配置文件server.json中的Log字段内配置。

eg.

	server.json
	{
		"Log":{
		        "dingtalk":{
		          "webhookurl":"https://oapi.dingtalk.com/robot/send?access_token=xxx",
		          // RFC5424 log message levels.
		          "level":3
		        },
		        "file":{
		          //是否按照每天 logrotate，默认是 true
		        	"daily":true, 	
		        	"level":7
		        }
		    }
	}
	
#### 配置与beego的一些区别

1. 每一种引擎都需要在Log中配置才能生效(file引擎除外)
2. file是默认引擎，Log不配置的话会使用默认配置
3. file引擎的filename字段无法设置，mqant会默认为access级别和error级别的日志分文件输出到约定的日志目录中


#### 引擎字段映射

	文件输出				file
	邮件输出				smtp
	简聊					jianliao
	slack					slack
	钉钉					dingtalk
	网络					conn
	ElasticSearch			es
	
#### 关闭控制台打印

>在正式环境中我们只需要在file中输出日志，不需要再控制台输出日志了，因此我们需要关闭控制台日志输出。

	func main() {
		app := mqant.CreateApp()
		
		app.Run(false, //只有是在调试模式下才会在控制台打印日志, 非调试模式下只在日志文件中输出日志
			modules.MasterModule(),
			gamecenter.Module(),
			）
	}