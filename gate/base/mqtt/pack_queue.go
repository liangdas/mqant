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
	"time"
)

// Tcp write queue
type PackQueue struct {
	conf conf.Mqtt
	// The last error in the tcp connection
	writeError error
	// Notice read the error
	errorChan chan error
	noticeFin chan byte
	writeChan chan *pakcAdnType
	readChan  chan<- *packAndErr
	// Pack connection
	r *bufio.Reader
	w *bufio.Writer

	conn network.Conn

	alive int
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
)

type pakcAdnType struct {
	pack *Pack
	typ  byte
}

// Init a pack queue
func NewPackQueue(conf conf.Mqtt, r *bufio.Reader, w *bufio.Writer, conn network.Conn, readChan chan<- *packAndErr, alive int) *PackQueue {
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
		noticeFin: make(chan byte, 2),
		writeChan: make(chan *pakcAdnType, conf.WirteLoopChanNum),
		readChan:  readChan,
		errorChan: make(chan error, 1),
	}
}

// Start a pack write queue
// It should run in a new grountine
func (queue *PackQueue) writeLoop() {
	// defer recover()
	var err error
loop:
	for {
		select {
		case pt, ok := <-queue.writeChan:
			if !ok {
				break loop
			}
			if pt == nil {
				break loop
			}
			if queue.conf.WriteTimeout > 0 {
				queue.conn.SetWriteDeadline(time.Now().Add(time.Second * time.Duration(queue.conf.WriteTimeout)))
			}
			switch pt.typ {
			case NO_DELAY:
				err = WritePack(pt.pack, queue.w)
			case DELAY:
				err = DelayWritePack(pt.pack, queue.w)
			case FLUSH:
				err = queue.w.Flush()
			}

			if err != nil {
				// Tell listener the error
				// Notice the read
				queue.writeError = err
				queue.errorChan <- err
				queue.noticeFin <- 0
				break loop
			}
		}
	}
}

// Write a pack , and get the last error
func (queue *PackQueue) WritePack(pack *Pack) error {
	if queue.writeError != nil {
		return queue.writeError
	}
	queue.writeChan <- &pakcAdnType{pack: pack}
	return nil
}

func (queue *PackQueue) WriteDelayPack(pack *Pack) error {
	if queue.writeError != nil {
		return queue.writeError
	}
	queue.writeChan <- &pakcAdnType{
		pack: pack,
		typ:  DELAY,
	}
	return nil
}

func (queue *PackQueue) SetAlive(alive int) error {
	if alive < 1 {
		alive = queue.conf.ReadTimeout
	}
	alive = int(float32(alive)*1.5 + 1)
	queue.alive = alive
	return nil
}

func (queue *PackQueue) Flush() error {
	if queue.writeError != nil {
		return queue.writeError
	}
	queue.writeChan <- &pakcAdnType{typ: FLUSH}
	return nil
}

// Read a pack and retuen the write queue error
//func (queue *PackQueue) ReadPack() (pack *mqtt.Pack, err error) {
//	ch := make(chan *packAndErr, 1)
//	go func() {
//		p := new(packAndErr)
//		if Conf.ReadTimeout > 0 {
//			queue.conn.SetReadDeadline(time.Now().Add(time.Second * time.Duration(Conf.ReadTimeout)))
//		}
//		p.pack, p.err = mqtt.ReadPack(queue.r)
//		ch <- p
//	}()
//	select {
//	case err = <-queue.errorChan:
//		// Hava an error
//		// pass
//	case pAndErr := <-ch:
//		pack = pAndErr.pack
//		err = pAndErr.err
//	}
//	return
//}

// Get a read pack queue
// Only call once
func (queue *PackQueue) ReadPackInLoop() {

	go func() {
		// defer recover()
		is_continue := true
		p := new(packAndErr)
	loop:
		for {
			if queue.alive > 0 {
				queue.conn.SetReadDeadline(time.Now().Add(time.Second * time.Duration(queue.alive)))
			}
			if is_continue {
				p.pack, p.err = ReadPack(queue.r)
				if p.err != nil {
					is_continue = false
				}
				select {
				case queue.readChan <- p:
					// Without anything to do
				case <-queue.noticeFin:
					//queue.Close()
					log.Info("Queue FIN")
					break loop
				}
			} else {
				<-queue.noticeFin
				//
				log.Info("Queue not continue")
				break loop
			}

			p = new(packAndErr)
		}
		queue.Close()
	}()
}

// Close the all of queue's channels
func (queue *PackQueue) Close() error {
	close(queue.writeChan)
	close(queue.readChan)
	close(queue.errorChan)
	close(queue.noticeFin)
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
