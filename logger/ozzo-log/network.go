// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package log

import (
	"errors"
	"fmt"
	"io"
	"net"
)

// NetworkTarget sends log messages over a network connection.
type NetworkTarget struct {
	*Filter
	// the network to connect to. Valid networks include
	// tcp", "tcp4" (IPv4-only), "tcp6" (IPv6-only),
	// "udp", "udp4" (IPv4-only), "udp6" (IPv6-only), "ip", "ip4"
	// (IPv4-only), "ip6" (IPv6-only), "unix", "unixgram" and
	// "unixpacket".
	Network string
	// the address on the network to connect to.
	// For TCP and UDP networks, addresses have the form host:port.
	// If host is a literal IPv6 address it must be enclosed
	// in square brackets as in "[::1]:80" or "[ipv6-host%zone]:80".
	Address string
	// whether to use a persistent network connection.
	// If this is false, for every message to be sent, a network
	// connection will be open and closed.
	Persistent bool
	// the size of the message channel.
	BufferSize int

	entries chan *Entry
	conn    net.Conn
	close   chan bool
}

// NewNetworkTarget creates a NetworkTarget.
// The new NetworkTarget takes these default options:
// MaxLevel: LevelDebug, Persistent: true, BufferSize: 1024.
// You must specify the Network and Address fields.
func NewNetworkTarget() *NetworkTarget {
	return &NetworkTarget{
		Filter:     &Filter{MaxLevel: LevelDebug},
		BufferSize: 1024,
		Persistent: true,
		close:      make(chan bool, 0),
	}
}

// Open prepares NetworkTarget for processing log messages.
func (t *NetworkTarget) Open(errWriter io.Writer) error {
	t.Filter.Init()

	if t.BufferSize < 0 {
		return errors.New("NetworkTarget.BufferSize must be no less than 0")
	}
	if t.Network == "" {
		return errors.New("NetworkTarget.Network must be specified")
	}
	if t.Address == "" {
		return errors.New("NetworkTarget.Address must be specified")
	}

	t.entries = make(chan *Entry, t.BufferSize)
	t.conn = nil

	if t.Persistent {
		if err := t.connect(); err != nil {
			return err
		}
	}

	go t.sendMessages(errWriter)

	return nil
}

// Process puts filtered log messages into a channel for sending over network.
func (t *NetworkTarget) Process(e *Entry) {
	if t.Allow(e) {
		select {
		case t.entries <- e:
		default:
		}
	}
}

// Close closes the network target.
func (t *NetworkTarget) Close() {
	<-t.close
}

func (t *NetworkTarget) connect() error {
	if t.conn != nil {
		t.conn.Close()
		t.conn = nil
	}

	conn, err := net.Dial(t.Network, t.Address)
	if err != nil {
		return err
	}

	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetKeepAlive(true)
	}

	t.conn = conn
	return nil
}

func (t *NetworkTarget) sendMessages(errWriter io.Writer) {
	for {
		entry := <-t.entries
		if entry == nil {
			if t.conn != nil {
				t.conn.Close()
			}
			t.close <- true
			break
		}
		if err := t.write(entry.String() + "\n"); err != nil {
			fmt.Fprintf(errWriter, "NetworkTarget write error: %v\n", err)
		}
	}
}

func (t *NetworkTarget) write(message string) error {
	if !t.Persistent {
		if err := t.connect(); err != nil {
			return err
		}
		defer t.conn.Close()
	}
	_, err := t.conn.Write([]byte(message))
	return err
}
