package logger

import (
	"mqant/logger/ozzo-log"
	"github.com/go-ozzo/ozzo-config"
)

type MqantLog struct {
	*log.Logger
}

var Mqlog = &MqantLog{}

func (m *MqantLog)GetDefaultLogger(){
	c := config.New()
	c.Load("app.json")

	// 注册供 Logger.Targets 使用的日志标的类型
	c.Register("ConsoleTarget", log.NewConsoleTarget)
	c.Register("FileTarget", log.NewFileTarget)

	logger := log.NewLogger()
	if err := c.Configure(logger, "Logger"); err != nil {
		panic(err)
	}

	logger.Open()

	m.Logger = logger
}

