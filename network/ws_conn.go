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
package network

import (
	"github.com/gorilla/websocket"
	//"github.com/liangdas/mqant/log"
	"net"
	"sync"
	"io"
	"bytes"
	"time"
)

type WebsocketConnSet map[*websocket.Conn]struct{}

type WSConn struct {
	io.Reader	//Read(p []byte) (n int, err error)
	io.Writer	//Write(p []byte) (n int, err error)
	sync.Mutex
	buf_lock chan bool	//当有写入一次数据设置一次
	buffer bytes.Buffer
	conn      *websocket.Conn
	closeFlag bool
}

func newWSConn(conn *websocket.Conn) *WSConn {
	wsConn := new(WSConn)
	wsConn.conn = conn
	wsConn.buf_lock=make(chan bool)

	go func() {
		for  {
			_, b, err := wsConn.conn.ReadMessage()
			if err!=nil{
				//log.Error("读取数据失败 %s",err.Error())
				wsConn.buf_lock<-false
			}else{
				wsConn.buffer.Write(b)
				wsConn.buf_lock<-true
			}
		}

		conn.Close()
		wsConn.Lock()
		wsConn.closeFlag = true
		wsConn.Unlock()
	}()

	return wsConn
}

func (wsConn *WSConn) doDestroy() {
	wsConn.conn.UnderlyingConn().(*net.TCPConn).SetLinger(0)
	wsConn.conn.Close()

	if !wsConn.closeFlag {
		wsConn.closeFlag = true
	}
}

func (wsConn *WSConn) Destroy() {
	wsConn.Lock()
	defer wsConn.Unlock()

	wsConn.doDestroy()
}

func (wsConn *WSConn) Close() (error){
	wsConn.Lock()
	defer wsConn.Unlock()
	if wsConn.closeFlag {
		return	nil
	}
	wsConn.closeFlag = true
	return wsConn.conn.Close()
}

func (wsConn *WSConn) Write(p []byte) (n int, err error){
	err = wsConn.conn.WriteMessage(websocket.BinaryMessage, p)
	if err != nil {
		return 0,err
	}
	return len(p),nil
}


// goroutine not safe
func (wsConn *WSConn) Read(p []byte) (n int, err error) {
	<-wsConn.buf_lock //等待写入数据
	return wsConn.buffer.Read(p)
}

func (wsConn *WSConn) LocalAddr() net.Addr {
	return wsConn.conn.LocalAddr()
}

func (wsConn *WSConn) RemoteAddr() net.Addr {
	return wsConn.conn.RemoteAddr()
}

// A zero value for t means I/O operations will not time out.
func (wsConn *WSConn) SetDeadline(t time.Time) error{
	err:= wsConn.conn.SetWriteDeadline(t)
	if err!=nil{
		return err
	}
	return wsConn.conn.SetWriteDeadline(t)
}

// SetReadDeadline sets the deadline for future Read calls.
// A zero value for t means Read will not time out.
func (wsConn *WSConn) SetReadDeadline(t time.Time) error{
	return wsConn.conn.SetReadDeadline(t)
}

// SetWriteDeadline sets the deadline for future Write calls.
// Even if write times out, it may return n > 0, indicating that
// some of the data was successfully written.
// A zero value for t means Write will not time out.
func (wsConn *WSConn) SetWriteDeadline(t time.Time) error{
	return wsConn.conn.SetWriteDeadline(t)
}


