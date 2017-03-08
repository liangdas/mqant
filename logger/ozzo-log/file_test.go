// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package log_test

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"mqant/logger/ozzo-log"
)

func TestNewFileTarget(t *testing.T) {
	target := log.NewFileTarget()
	if target.MaxLevel != log.LevelDebug {
		t.Errorf("NewFileTarget.MaxLevel = %v, expected %v", target.MaxLevel, log.LevelDebug)
	}
	if target.Rotate != true {
		t.Errorf("NewFileTarget.Rotate = %v, expected %v", target.Rotate, true)
	}
	if target.BackupCount != 10 {
		t.Errorf("NewFileTarget.BackupCount = %v, expected %v", target.BackupCount, 10)
	}
	if target.MaxBytes != (1 << 20) {
		t.Errorf("NewFileTarget.MaxBytes = %v, expected %v", target.MaxBytes, 1<<20)
	}
}

func TestFileTarget(t *testing.T) {
	logFile := "app.log"
	os.Remove(logFile)

	logger := log.NewLogger()
	target := log.NewFileTarget()
	target.FileName = logFile
	target.Categories = []string{"system.*"}
	logger.Targets = append(logger.Targets, target)
	logger.Open()
	logger.Info("t1: %v", 2)
	logger.GetLogger("system.db").Info("t2: %v", 3)
	logger.Close()

	bytes, err := ioutil.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if strings.Contains(string(bytes), "t1: 2") {
		t.Errorf("Found unexpected %q", "t1: 2")
	}
	if !strings.Contains(string(bytes), "t2: 3") {
		t.Errorf("Expected %q not found", "t2: 3")
	}
}
