package log

import (
	"fmt"
	"github.com/liangdas/mqant/logger/ozzo-log"
	"time"
)

type MqantLog struct {
	*log.Logger
}

func NewMqantLog(debug bool, ProcessID string, Logdir string) *log.Logger {
	var Mqlog = &MqantLog{}
	Mqlog.GetLogger(debug, ProcessID, Logdir)
	return Mqlog.Logger
}
func NewDefaultLogger()(*log.Logger)  {
	logger := log.NewLogger()
	logger.CallStackDepth = 3
	t1 := log.NewConsoleTarget()
	t1.MinLevel = log.LevelNotice
	t1.MaxLevel = log.LevelDebug
	logger.Targets = append(logger.Targets, t1)
	return logger
}
func (m *MqantLog) GetLogger(debug bool, ProcessID string, Logdir string) {
	// 创建根记录器(root logger)
	logger := log.NewLogger()
	logger.CallStackDepth = 3
	//pid.nohup.log
	//pid.access.log
	//pid.error.log
	if debug {
		// 添加一个控制台标的（Console Target）和一个文件标的（File Target）
		t1 := log.NewConsoleTarget()
		logger.Targets = append(logger.Targets, t1)
	} else {
		t2 := log.NewFileTarget()
		t2.FileName = fmt.Sprintf("%s/%s.error.log", Logdir, ProcessID)
		t2.MaxLevel = log.LevelWarning
		t2.MinLevel = log.LevelEmergency
		t3 := log.NewFileTarget()
		t3.FileName = fmt.Sprintf("%s/%s.access.log", Logdir, ProcessID)
		t3.MinLevel = log.LevelNotice
		t3.MaxLevel = log.LevelDebug
		logger.Targets = append(logger.Targets, t2, t3)
	}

	logger.Open()
	//defer logger.Close()
	logger = logger.GetLogger("app", func(l *log.Logger, e *log.Entry) string {
		if e.Level <= log.LevelWarning {
			return fmt.Sprintf("%v [%v] %v %v", e.Time.Format(time.RFC3339), e.Level, e.Message, e.CallStack)
		} else {
			return fmt.Sprintf("%v [%v][%v] %v", e.Time.Format(time.RFC3339), e.Level, e.ShortFile, e.Message)
		}
	})
	m.Logger = logger
}
