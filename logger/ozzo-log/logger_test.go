// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package log_test

import (
	"io"
	"testing"

	"github.com/go-ozzo/ozzo-config"
	"mqant/logger/ozzo-log"
)

func TestNewLogger(t *testing.T) {
	logger := log.NewLogger()
	if logger.MaxLevel != log.LevelDebug {
		t.Errorf("NewLogger().MaxLevel = %v, expected %v", logger.MaxLevel, log.LevelDebug)
	}
	if logger.Category != "app" {
		t.Errorf("NewLogger().Category = %v, expected %v", logger.Category, "app")
	}
	if logger.CallStackDepth != 0 {
		t.Errorf("NewLogger().CallStackDepth = %v, expected %v", logger.CallStackDepth, 0)
	}
}

func TestGetLogger(t *testing.T) {
	formatter := func(*log.Logger, *log.Entry) string {
		return "test"
	}
	logger := log.NewLogger()
	logger1 := logger.GetLogger("testing")
	if logger1.Category != "testing" {
		t.Errorf("logger1.Category = %v, expected %v", logger1.Category, "testing")
	}
	logger2 := logger.GetLogger("routing", formatter)
	if logger2.Category != "routing" {
		t.Errorf("logger2.Category = %v, expected %v", logger2.Category, "routing")
	}
	if logger2.Formatter(logger2, nil) != "test" {
		t.Errorf("logger2.Formatter has an unexpected value")
	}
}

type MemoryTarget struct {
	entries []*log.Entry
	open    bool
	ready   chan bool
	Option1 string
	Option2 bool
}

func (m *MemoryTarget) Open(io.Writer) error {
	m.open = true
	m.entries = make([]*log.Entry, 0)
	return nil
}

func (m *MemoryTarget) Process(e *log.Entry) {
	if e == nil {
		m.ready <- true
	} else {
		m.entries = append(m.entries, e)
	}
}

func (t *MemoryTarget) Close() {
	<-t.ready
}

func TestLoggerLog(t *testing.T) {
	logger := log.NewLogger()
	target := &MemoryTarget{
		ready: make(chan bool, 0),
	}
	logger.Targets = append(logger.Targets, target)

	if target.open {
		t.Errorf("target.open = %v, expected %v", target.open, false)
	}
	logger.Open()
	if !target.open {
		t.Errorf("target.open = %v, expected %v", target.open, true)
	}

	logger.Log(log.LevelInfo, "t0: %v", 1)
	logger.Debug("t1: %v", 2)
	logger.Info("t2")
	logger.Warning("t3")
	logger.Notice("t4")
	logger.Error("t5")
	logger.Critical("t6")
	logger.Alert("t7")
	logger.Emergency("t8")

	logger.Close()

	if len(target.entries) != 9 {
		t.Errorf("len(target.entries) = %v, expected %v", len(target.entries), 9)
	}
	levels := ""
	messages := ""
	for i := 0; i < 9; i++ {
		levels += target.entries[i].Level.String() + ","
		messages += target.entries[i].Message + ","
	}
	expectedLevels := "Info,Debug,Info,Warning,Notice,Error,Critical,Alert,Emergency,"
	expectedMessages := "t0: 1,t1: 2,t2,t3,t4,t5,t6,t7,t8,"
	if levels != expectedLevels {
		t.Errorf("levels = %v, expected %v", levels, expectedLevels)
	}
	if messages != expectedMessages {
		t.Errorf("messages = %v, expected %v", messages, expectedMessages)
	}
}

func TestLoggerConfig(t *testing.T) {
	c := config.New()
	err := c.LoadJSON([]byte(`{
		"Logger": {
			"MaxLevel": 2,
			"Category": "app2",
			"Targets": [
				{
					"type": "memory1",
					"Option1": "abc",
					"Option2": true
				},
				{
					"type": "memory2",
					"Option1": "xyz"
				}
			]
		}
	}`))
	if err != nil {
		t.Errorf("config.LoadJSON(): %v", err)
	}
	c.Register("memory1", func() *MemoryTarget {
		return &MemoryTarget{}
	})
	c.Register("memory2", func() *MemoryTarget {
		return &MemoryTarget{Option2: true}
	})
	logger := log.NewLogger()

	if err := c.Configure(logger, "Logger"); err != nil {
		t.Errorf("config.Configure(logger): %v", err)
	}
	if logger.MaxLevel != log.LevelCritical {
		t.Errorf("logger.MaxLevel = %v, expected %v", logger.MaxLevel, log.LevelCritical)
	}
	if logger.Category != "app2" {
		t.Errorf("logger.Category = %v, expected %v", logger.Category, "app2")
	}

	if len(logger.Targets) != 2 {
		t.Errorf("len(logger.Targets) = %v, expected %v", len(logger.Targets), 2)
		return
	}
	m1 := logger.Targets[0].(*MemoryTarget)
	m2 := logger.Targets[1].(*MemoryTarget)
	if m1.Option1 != "abc" || m1.Option2 != true {
		t.Errorf("m1.Option1 = %v, Option2 = %v, expected %v and %v", m1.Option1, m1.Option2, "abc", true)
	}
	if m2.Option1 != "xyz" || m2.Option2 != true {
		t.Errorf("m2.Option1 = %v, Option2 = %v, expected %v and %v", m2.Option1, m2.Option2, "xyz", true)
	}
}
