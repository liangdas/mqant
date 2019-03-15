// Copyright 2014 mqant Author. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package mqtt

import (
	"bufio"
	"fmt"
	"github.com/liangdas/mqant/conf"
	"github.com/liangdas/mqant/log"
	"github.com/liangdas/mqant/network"
	"runtime"
	"time"
	"sync"
)

// Tcp write queue
type PackQueue struct {
	conf conf.Mqtt
	// The last error in the tcp connection
	writeError error
	// Notice read the error
	fch		chan struct{}
	writelock   	sync.Mutex
	recover 	func (pAndErr *packAndErr) (err error)
	// Pack connection
	r *bufio.Reader
	w *bufio.Writer

	conn network.Conn

	alive int

	status	int
}

type packAndErr struct {
	pack *Pack
	err  error
}

// 1 is delay, 0 is no delay, 2 is just flush.
const (
	NO_DELAY = iota
	DELAY
	FLUSH

	DISCONNECTED = iota
	CONNECTED
	CLOSED
	RECONNECTING
	CONNECTING
)

type packAndType struct {
	pack *Pack
	typ  byte
}

// Init a pack queue
func NewPackQueue(conf conf.Mqtt, r *bufio.Reader, w *bufio.Writer, conn network.Conn, recover 	func (pAndErr *packAndErr) (err error), alive int) *PackQueue {
	if alive < 1 {
		alive = conf.ReadTimeout
	}
	alive = int(float32(alive)*1.5 + 1)
	return &PackQueue{
		conf:      conf,
		alive:     alive,
		r:         r,
		w:         w,
		conn:      conn,
		recover:  recover,
		fch:		make(chan struct{},1024),
		status:		CONNECTED,
	}
}

func (queue *PackQueue) isConnected() bool{
	return queue.status == CONNECTED
}
// Get a read pack queue
// Only call once
func (queue *PackQueue) Flusher () {
	for queue.isConnected(){
		if _, ok := <-queue.fch; !ok {
			break
		}
		queue.writelock.Lock()
		if !queue.isConnected(){
			queue.writelock.Unlock()
			break
		}
		if queue.w.Buffered() > 0 {
			if err := queue.w.Flush(); err != nil {
				queue.writelock.Unlock()
				break
			}
		}
		queue.writelock.Unlock()
	}
	log.Info("flusher_loop Groutine will esc.")
}

// Write a pack , and get the last error
func (queue *PackQueue) WritePack(pack *Pack) (err error) {
	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, 1024)
			l := runtime.Stack(buf, false)
			errstr := string(buf[:l])
			err=fmt.Errorf("WritePack error %v",errstr)
			queue.Close(err)
		}
	}()
	if queue.writeError != nil {
		return queue.writeError
	}
	queue.writelock.Lock()
	err=DelayWritePack(pack,queue.w)
	queue.fch<- struct{}{}
	queue.writelock.Unlock()
	if err != nil {
		// Tell listener the error
		// Notice the read
		queue.Close(err)
	}
	return err
}

func (queue *PackQueue) SetAlive(alive int) error {
	if alive < 1 {
		alive = queue.conf.ReadTimeout
	}
	alive = int(float32(alive)*1.5 + 1)
	queue.alive = alive
	return nil
}


// Get a read pack queue
// Only call once
func (queue *PackQueue) ReadPackInLoop() {
	// defer recover()
	p := new(packAndErr)
	loop:
		for queue.isConnected(){
			if queue.alive > 0 {
				queue.conn.SetReadDeadline(time.Now().Add(time.Second * time.Duration(int(float64(queue.alive)*1.5))))
			} else {
				queue.conn.SetReadDeadline(time.Now().Add(time.Second * 90))
			}
			p.pack, p.err = ReadPack(queue.r)
			if p.err != nil {
				queue.Close(p.err)
				break loop
			}
			err:=queue.recover(p)
			if err != nil {
				queue.Close(err)
				break loop
			}
			p = new(packAndErr)
		}

	log.Info("read_loop Groutine will esc.")
}

// Close the all of queue's channels
func (queue *PackQueue) Close(err error) error {
	queue.writeError = err
	close(queue.fch)
	queue.status=CLOSED
	return nil
}

// Buffer
type buffer struct {
	index int
	data  []byte
}

func newBuffer(data []byte) *buffer {
	return &buffer{
		data:  data,
		index: 0,
	}
}
func (b *buffer) readString(length int) (s string, err error) {
	if (length + b.index) > len(b.data) {
		err = fmt.Errorf("Out of range error:%v", length)
		return
	}
	s = string(b.data[b.index:(length + b.index)])
	b.index += length
	return
}
func (b *buffer) readByte() (c byte, err error) {
	if (1 + b.index) > len(b.data) {
		err = fmt.Errorf("Out of range error")
		return
	}
	c = b.data[b.index]
	b.index++
	return
}

