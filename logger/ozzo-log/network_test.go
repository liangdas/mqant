// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package log_test

import (
	"net"
	"strings"
	"testing"

	"mqant/logger/ozzo-log"
)

func TestNewNetworkTarget(t *testing.T) {
	target := log.NewNetworkTarget()
	if target.MaxLevel != log.LevelDebug {
		t.Errorf("NetworkTarget.MaxLevel = %v, expected %v", target.MaxLevel, log.LevelDebug)
	}
	if !target.Persistent {
		t.Errorf("NetworkTarget.Persistent should be true, got false")
	}
}

type LogServer struct {
	t      *testing.T
	done   chan bool
	buffer []byte
}

func (s *LogServer) Start(network, address string) error {
	s.done = make(chan bool, 0)
	s.buffer = make([]byte, 1024)
	listener, err := net.Listen(network, address)
	if err != nil {
		return err
	}
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			s.t.Errorf("Server.Accept(): %v", err)
			s.done <- true
			return
		}
		if _, err := conn.Read(s.buffer); err != nil {
			s.t.Errorf("Server read error: %v", err)
		}
		s.done <- true
	}()

	return nil
}

func TestNetworkTarget(t *testing.T) {
	network := "tcp"
	address := ":10234"
	server := &LogServer{t: t}
	if err := server.Start(network, address); err != nil {
		t.Errorf("server.Open(): %v", err)
		return
	}

	logger := log.NewLogger()
	target := log.NewNetworkTarget()
	target.Network = network
	target.Address = address
	target.Categories = []string{"system.*"}
	logger.Targets = append(logger.Targets, target)
	logger.Open()

	logger.Info("t1: %v", 2)
	logger.GetLogger("system.db").Info("t2: %v", 3)

	logger.Close()

	<-server.done

	result := string(server.buffer)
	if strings.Contains(result, "t1: 2") {
		t.Errorf("Found unexpected %q", "t1: 2")
	}
	if !strings.Contains(result, "t2: 3") {
		t.Errorf("Expected %q not found", "t2: 3")
	}
}
