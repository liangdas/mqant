# ozzo-log

[![GoDoc](https://godoc.org/github.com/go-ozzo/ozzo-log?status.png)](http://godoc.org/github.com/go-ozzo/ozzo-log)
[![Build Status](https://travis-ci.org/go-ozzo/ozzo-log.svg?branch=master)](https://travis-ci.org/go-ozzo/ozzo-log)
[![Coverage Status](https://coveralls.io/repos/github/go-ozzo/ozzo-log/badge.svg?branch=master)](https://coveralls.io/github/go-ozzo/ozzo-log?branch=master)
[![Go Report](https://goreportcard.com/badge/github.com/go-ozzo/ozzo-log)](https://goreportcard.com/report/github.com/go-ozzo/ozzo-log)

## Other languages

[简体中文](/docs/README-zh-CN.md) [Русский](/docs/README-ru.md)

## Description

ozzo-log is a Go package providing enhanced logging support for Go programs. It has the following features:

* High performance through asynchronous logging;
* Recording message severity levels;
* Recording message categories;
* Recording message call stacks;
* Filtering via severity levels and categories;
* Customizable message format;
* Configurable and pluggable message handling through log targets;
* Included console, file, network, and email log targets.

## Requirements

Go 1.2 or above.

## Installation

Run the following command to install the package:

```
go get github.com/go-ozzo/ozzo-log
```

## Getting Started

The following code snippet shows how you can use this package.

```go
package main

import (
	"github.com/go-ozzo/ozzo-log"
)

func main() {
    // creates the root logger
	logger := log.NewLogger()

	// adds a console target and a file target
	t1 := log.NewConsoleTarget()
	t2 := log.NewFileTarget()
	t2.FileName = "app.log"
	t2.MaxLevel = log.LevelError
	logger.Targets = append(logger.Targets, t1, t2)

	logger.Open()
	defer logger.Close()

	// calls log methods to log various log messages
	logger.Error("plain text error")
	logger.Error("error with format: %v", true)
	logger.Debug("some debug info")

	// customizes log category
	l := logger.GetLogger("app.services")
	l.Info("some info")
	l.Warning("some warning")

	...
}
```

## Loggers and Targets

A logger provides various log methods that can be called by application code
to record messages of various severity levels.

A target filters log messages by their severity levels and message categories
and processes the filtered messages in various ways, such as saving them in files,
sending them in emails, etc.

A logger can be equipped with multiple targets each with different filtering conditions.

The following targets are included in the ozzo-log package.

* `ConsoleTarget`: displays filtered messages to console window
* `FileTarget`: saves filtered messages in a file (supporting file rotating)
* `NetworkTarget`: sends filtered messages to an address on a network
* `MailTarget`: sends filtered messages in emails

You can create a logger, configure its targets, and start to use logger with the following code:

```go
// creates the root logger
logger := log.NewLogger()
logger.Targets = append(logger.Targets, target1, target2, ...)
logger.Open()
...calling log methods...
logger.Close()
```

## Severity Levels

You can log a message of a particular severity level (following the RFC5424 standard)
by calling one of the following methods of the `Logger` struct:

* `Emergency()`: the system is unusable.
* `Alert()`: action must be taken immediately.
* `Critical()`: critical conditions.
* `Error()`: error conditions.
* `Warning()`: warning conditions.
* `Notice()`: normal but significant conditions.
* `Info()`: informational purpose.
* `Debug()`: debugging purpose.

## Message Categories

Each log message is associated with a category which can be used to group messages.
For example, you may use the same category for messages logged by the same Go package.
This will allow you to selectively send messages to different targets.

When you call `log.NewLogger()`, a root logger is returned which logs messages using
the category named as `app`. To log messages with a different category, call the `GetLogger()`
method of the root logger or a parent logger to get a child logger and then call its
log methods:

```go
logger := log.NewLogger()
// the message is of category "app"
logger.Error("...")

l1 := logger.GetLogger("system")
// the message is of category "system"
l1.Error("...")

l2 := l1.GetLogger("app.models")
// the message is of category "app.models"
l2.Error("...")
```

## Message Formatting

By default, each log message takes this format when being sent to different targets:

```
2015-10-22T08:39:28-04:00 [Error][app.models] something is wrong
...call stack (if enabled)...
```

You may customize the message format by specifying your own message formatter when calling
`Logger.GetLogger()`. For example,

```go
logger := log.NewLogger()
logger = logger.GetLogger("app", func (l *Logger, e *Entry) string {
    return fmt.Sprintf("%v [%v][%v] %v%v", e.Time.Format(time.RFC822Z), e.Level, e.Category, e.Message, e.CallStack)
})
```


## Logging Call Stacks

By setting `Logger.CallStackDepth` as a positive number, it is possible to record call stack information for
each log method call. You may further configure `Logger.CallStackFilter` so that only call stack frames containing
the specified substring will be recorded. For example,

```go
logger := log.NewLogger()
// record call stacks containing "myapp/src" up to 5 frames per log message
logger.CallStackDepth = 5
logger.CallStackFilter = "myapp/src"
```


## Message Filtering

By default, messages of *all* severity levels will be recorded. You may customize
`Logger.MaxLevel` to change this behavior. For example,

```go
logger := log.NewLogger()
// only record messages between Emergency and Warning levels
logger.MaxLevel = log.LevelWarning
```

Besides filtering messages at the logger level, a finer grained message filtering can be done
at target level. For each target, you can specify its `MaxLevel` similar to that with the logger;
you can also specify which categories of the messages the target should handle. For example,

```go
target := log.NewConsoleTarget()
// handle messages between Emergency and Info levels
target.MaxLevel = log.LevelInfo
// handle messages of categories which start with "system.db." or "app."
target.Categories = []string{"system.db.*", "app.*"}
```

## Configuring Logger

When an application is deployed for production, a common need is to allow changing
the logging configuration of the application without recompiling its source code.
ozzo-log is designed with this in mind.

For example, you can use a JSON file to specify how the application and its
logger should be configured:

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

Assuming the JSON file is `app.json`, in your application code you can use the `ozzo-config` package
to load the JSON file and configure the logger used by the application:

```go
package main

import (
	"github.com/go-ozzo/ozzo-config"
    "github.com/go-ozzo/ozzo-log"
)

func main() {
    c := config.New()
    c.Load("app.json")
    // register the target types to allow configuring Logger.Targets.
    c.Register("ConsoleTarget", log.NewConsoleTarget)
    c.Register("FileTarget", log.NewFileTarget)

    logger := log.NewLogger()
    if err := c.Configure(logger, "Logger"); err != nil {
        panic(err)
    }
}
```

To change the logger configuration, simply modify the JSON file without
recompiling the Go source files.
