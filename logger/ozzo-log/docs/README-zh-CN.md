# ozzo-log

[![GoDoc](https://godoc.org/github.com/go-ozzo/ozzo-log?status.png)](http://godoc.org/github.com/go-ozzo/ozzo-log)
[![Build Status](https://travis-ci.org/go-ozzo/ozzo-log.svg?branch=master)](https://travis-ci.org/go-ozzo/ozzo-log)
[![Coverage](http://gocover.io/_badge/github.com/go-ozzo/ozzo-log)](http://gocover.io/github.com/go-ozzo/ozzo-log)

## 其他语言

[English](../README.md) [Русский](/docs/README-ru.md)

## 说明

ozzo-log 是给 Go 程序提供加强型日志功能的 go 包。它支持以下功能：

* 通过异步记录实现的高性能
* 记录消息的严重级别(severity level)
* 记录消息的分类(Category)
* 记录错误信息的调用栈(call stack)
* 根据严重级别和消息分类进行过滤分流(Filtering)
* 信息格式可以自由定制(Formating)
* 通过日志标的(log target)对象，实现可配置、可接入的信息处理过程。
* 多种日志标的，包括 console、file、network 以及 email

## 需求

Go 1.2 或以上。

## 安装

执行以下指令安装：

```
go get github.com/go-ozzo/ozzo-log
```

## 准备开始

以下代码片段展示了本包的基本用法：

```go
package main

import (
	"github.com/go-ozzo/ozzo-log"
)

func main() {
	// 创建根记录器(root logger)
	logger := log.NewLogger()

	// 添加一个控制台标的（Console Target）和一个文件标的（File Target）
	t1 := log.NewConsoleTarget()
	t2 := log.NewFileTarget()
	t2.FileName = "app.log"
	t2.MaxLevel = log.LevelError
	logger.Targets = append(logger.Targets, t1, t2)

	logger.Open()
	defer logger.Close()

	// 调用不同的记录方法记录不同的日志信息。
	logger.Error("plain text error")
	logger.Error("error with format: %v", true)
	logger.Debug("some debug info")

	// 自定义日志类别
	l := logger.GetLogger("app.services")
	l.Info("some info")
	l.Warning("some warning")

	...
}
```

## 记录器与标的 (Logger and Targets)

记录器提供了一系列用于记录多种不同严重级别的日志方法。

标的对象会将这些日志消息依严重级别和分类的不同进行过滤分流（filtering），将数据以不同的方式对进行处理，比如写入文件或者以邮件形式发送等等。

记录器可以装配不同的标的对象，应用不同的过滤规则。

ozzo-log 原生自带以下过滤器：

* `ConsoleTarget`：于控制台窗口内显示过滤后的信息
* `FileTarget`：将过滤后的消息写入文件，支持文件分页轮替（file rotate）
* `NetworkTarget`：将过滤后的信息发送至网络内的某一地址
* `MailTarget`：把过滤后的信息以邮件形式发送

你可以创建一个记录器，配置好标的，然后就可以进行日志记录了：

```go
// 创建根记录器
logger := log.NewLogger()
logger.Targets = append(logger.Targets, target1, target2, ...)
logger.Open()
...于此处调用日志方法...
logger.Close()
```

## 严重级别 (Severity Levels)

记录消息时，可以通过调用 `Logger` 结构的下列方法记录特定的严重级别（级别的设置依照 RFC5424 标准）

* `Emergency()`：系统完全无法使用的情形
* `Alert()`：必须立刻采取措施的情形
* `Critical()`：危笃情形
* `Error()`：错误情形
* `Warning()`：警告情形
* `Notice()`：正常但值得注意的情形
* `Info()`：为了记录信息
* `Debug()`：为了调试

## 信息分类 (Message Categories)

每条日志消息都会关联有一个可用于信息分组的类别标识。比如，你可以给来自同一个 Go 
包的日志记录以相同的类别。之后就可以有选择地把不同类别的信息发送到不同的标的。分别进行处理。

调用 `log.NewLogger()` 方法会返回一个类型设定为 `app`（默认值）的根记录器。要记录不同的类别，则可以调用某根记录器或父记录器的 
`GetLogger()` 方法，从而获得一个不同类别的子记录器。

```go
logger := log.NewLogger()
// 消息归类于 "app"
logger.Error("...")

l1 := logger.GetLogger("system")
// 消息归类于 "system"
l1.Error("...")

l2 := l1.GetLogger("app.models")
// 消息归类于 "app.models"
l2.Error("...")
```

## 信息格式 (Message Formatting)

默认情况下，发送给不同标的的消息均会使用以下缺省格式：

```
2015-10-22T08:39:28-04:00 [Error][app.models] something is wrong
...调用栈（如果启用）...
```

在调用 `Logger.GetLogger()` 时，可以通过指定你自己的格式化器来对信息格式进行自定义。比如：

```go
logger := log.NewLogger()
logger = logger.GetLogger("app", func (l *Logger, e *Entry) string {
	return fmt.Sprintf("%v [%v][%v] %v%v", e.Time.Format(time.RFC822Z), e.Level, e.Category, e.Message, e.CallStack)
})
```


## 记录调用栈

通过给 `Logger.CallStackDepth` 属性设置一个正整数，人们是可以记录下每次日志方法被调用时的调用栈信息。也可以再进一步，通过配置 
`Logger.CallStackFilter`，让其只记录包含特定子字符串的调用栈栈帧。以下是示例：

```go
logger := log.NewLogger()
// 记录 "myapp/src" 的调用栈，最深可达每消息 5 栈帧
logger.CallStackDepth = 5
logger.CallStackFilter = "myapp/src"
```

## 信息过滤 (Message Filtering)

根据初始的设置，**所有**严重级别的消息都会被记录。但可以通过修改 `Logger.MaxLevel` 
从而改变这一默认行为，像这样：

> 译注：这里的 Max 背后所隐含的顺序是指从**紧急**到**调试**依次递增。可以理解为越高等级的记录意味着记录更多且更琐碎的东西

```go
logger := log.NewLogger()
// 只记录级别在 Emergency（紧急）和 Warning（警告）之间的消息
logger.MaxLevel = log.LevelWarning
```

除了在记录器层级进行过滤之外，也可以在日志标的层级进行更加细粒度地过滤。对于每一个标的，可以单独指定它的 
`MaxLevel`，用法则与在记录器层级配置时类似；你也能指定各个标的应该处理哪些消息类别。举个栗子详细解释下：

```go
target := log.NewConsoleTarget()
// 此标的会处理级别在 Emergency 和 Info 之间的消息
target.MaxLevel = log.LevelInfo
// 处理所有消息分类以 "system.db." 或 "app." 开头的消息
target.Categories = []string{"system.db.*", "app.*"}
```

## 配置记录器

当应用被部署于生产环境时，会有一个非常常见的需求，也就是在无需重新编译源代码的前提下，对记录器进行配置。ozzo-log 
在设计时就包含了这个考量。

具体来说，你可以用一个如下的 JSON 文件来配置日志器：

```
{
	"Logger": {
		"Targets": [
			{
				"type": "ConsoleTarget",
			},
			{
				"type": "FileTarget",
				"FileName": "app.log",
				"MaxLevel": 4   // Warning or above
			}
		]
	}
}
```

打个比方，若你的 JSON 文件名叫 `app.json`，我们便可通过 `ozzo-config` 包实现对 JSON 文件的导入和对日志记录器的配置：

```go
package main

import (
	"github.com/go-ozzo/ozzo-config"
	"github.com/go-ozzo/ozzo-log"
)

func main() {
	c := config.New()
	c.Load("app.json")
	// 注册供 Logger.Targets 使用的日志标的类型
	c.Register("ConsoleTarget", log.NewConsoleTarget)
	c.Register("FileTarget", log.NewFileTarget)

	logger := log.NewLogger()
	if err := c.Configure(logger, "Logger"); err != nil {
		panic(err)
	}
}
```

要修改 logger 的配置，只需修改 JSON 文件就好了，无需重新编译 Go 的源文件。