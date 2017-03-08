// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package log_test

import (
	"strings"
	"testing"

	"mqant/logger/ozzo-log"
)

func TestNewConsoleTarget(t *testing.T) {
	target := log.NewConsoleTarget()
	if target.MaxLevel != log.LevelDebug {
		t.Errorf("ConsoleTarget.MaxLevel = %v, expected %v", target.MaxLevel, log.LevelDebug)
	}
	if target.ColorMode != true {
		t.Errorf("ConsoleTarget.ColorMode = %v, expected %v", target.ColorMode, true)
	}
}

type MemoryWriter struct {
	bytes []byte
}

func (m *MemoryWriter) Write(p []byte) (int, error) {
	if m.bytes == nil {
		m.bytes = make([]byte, 0)
	}
	m.bytes = append(m.bytes, p...)
	return len(p), nil
}

type ConsoleTargetMock struct {
	done chan bool
	*log.ConsoleTarget
}

func (t *ConsoleTargetMock) Process(e *log.Entry) {
	t.ConsoleTarget.Process(e)
	if e == nil {
		t.done <- true
	}
}

func TestConsoleTarget(t *testing.T) {
	logger := log.NewLogger()
	target := &ConsoleTargetMock{
		done:          make(chan bool, 0),
		ConsoleTarget: log.NewConsoleTarget(),
	}
	writer := &MemoryWriter{}
	target.Writer = writer
	target.Categories = []string{"system.*"}
	logger.Targets = append(logger.Targets, target)
	logger.Open()

	logger.Info("t1: %v", 2)
	logger.GetLogger("system.db").Info("t2: %v", 3)

	logger.Close()
	<-target.done

	if strings.Contains(string(writer.bytes), "t1: 2") {
		t.Errorf("Found unexpected %q", "t1: 2")
	}
	if !strings.Contains(string(writer.bytes), "t2: 3") {
		t.Errorf("Expected %q not found", "t2: 3")
	}
}
