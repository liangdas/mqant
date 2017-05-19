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
	"errors"
	"fmt"
	"github.com/liangdas/mqant/conf"
	"github.com/liangdas/mqant/log"
	"github.com/liangdas/mqant/network"
	"math"
	"sync"
)

var notAlive = errors.New("Connection was dead")

type PackRecover interface {
	OnRecover(*Pack)
}

type Client struct {
	queue *PackQueue

	recover  PackRecover        //消息接收者,从上层接口传过来的 只接收正式消息(心跳包,回复包等都不要)
	readChan <-chan *packAndErr //读取底层接收到的所有数据包包

	closeChan   chan byte // Other gorountine Call notice exit
	isSendClose bool      // Wheather has a new login user.
	isLetClose  bool      // Wheather has relogin.

	isStop bool
	lock   *sync.Mutex

	// Online msg id
	curr_id int
}

func NewClient(conf conf.Mqtt, recover PackRecover, r *bufio.Reader, w *bufio.Writer, conn network.Conn, alive int) *Client {
	readChan := make(chan *packAndErr, conf.ReadPackLoop)
	return &Client{
		readChan:  readChan,
		queue:     NewPackQueue(conf, r, w, conn, readChan, alive),
		recover:   recover,
		closeChan: make(chan byte),
		lock:      new(sync.Mutex),
		curr_id:   0,
	}
}

// Push the msg and response the heart beat
func (c *Client) Listen_loop() (e error) {
	defer func() {
		if r := recover(); r != nil {
			if c.isSendClose {
				c.closeChan <- 0
			}
		}
	}()
	var (
		err error
		// wg        = new(sync.WaitGroup)
	)

	// Start the write queue
	go c.queue.writeLoop()

	c.queue.ReadPackInLoop()

	// Start push 读取数据包
loop:
	for {
		select {
		case pAndErr, ok := <-c.readChan:
			if !ok {
				log.Info("Get a connection error")
				break loop
			}
			if err = c.waitPack(pAndErr); err != nil {
				log.Info("Get a connection error , will break(%v)", err)
				break loop
			}
		case <-c.closeChan:
			c.waitQuit()
			break loop
		}
	}

	c.lock.Lock()
	c.isStop = true
	c.lock.Unlock()
	// Wrte the onlines msg to the db
	// Free resources
	// Close channels

	close(c.closeChan)
	log.Info("listen_loop Groutine will esc.")
	return
}

// Setting a mqtt pack's id.
func (c *Client) getOnlineMsgId() int {
	if c.curr_id == math.MaxUint16 {
		c.curr_id = 1
		return c.curr_id
	} else {
		c.curr_id = c.curr_id + 1
		return c.curr_id
	}
}

func (c *Client) waitPack(pAndErr *packAndErr) (err error) {
	// If connetion has a error, should break
	// if it return a timeout error, illustrate
	// hava not recive a heart beat pack at an
	// given time.
	if pAndErr.err != nil {
		err = pAndErr.err
		return
	}
	//log.Debug("Client msg(%v)\n", pAndErr.pack.GetType())

	// Choose the requst type
	switch pAndErr.pack.GetType() {
	case CONNECT:
		info, ok := (pAndErr.pack.GetVariable()).(*Connect)
		if !ok {
			err = errors.New("It's not a mqtt connection package.")
			return
		}
		//id := info.GetUserName()
		//psw := info.GetPassword()
		c.queue.SetAlive(info.GetKeepAlive())
		err = c.queue.WritePack(GetConnAckPack(0))
	case PUBLISH:
		pub := pAndErr.pack.GetVariable().(*Publish)
		//// Del the msg
		//c.delMsg(ack.GetMid())
		//这里向上层转发消息
		//log.Debug("Ack To Client Qos(%d) mid(%d) Topic(%v) msg(%s) \n",pAndErr.pack.GetQos(),pub.GetMid(), *pub.GetTopic(),pub.GetMsg())
		if pAndErr.pack.GetQos() == 1 {
			//回复已收到
			//log.Debug("Ack To Client By PUBACK \n")
			err = c.queue.WritePack(GetPubAckPack(pub.GetMid()))
			if err != nil {
				//log.Debug("PUBACK error(%s) \n",err.Error())
			}
		} else if pAndErr.pack.GetQos() == 2 {
			//log.Debug("Ack To Client By PUBREC \n")
			err = c.queue.WritePack(GetPubRECPack(pub.GetMid()))
		}
		//log.Debug("ss",string(pub.GetMsg()))
		//目前这个版本暂时先不保证消息的Qos 默认用Qos=1吧
		c.recover.OnRecover(pAndErr.pack)
	case PUBACK: //4
		//用于 Qos =1 的消息
		//ack := pAndErr.pack.GetVariable().(*mqtt.Puback)
		//log.Debug("Client Ack Qos(%d) Dup(%d) mid(%d) \n",pAndErr.pack.GetQos(),pAndErr.pack.GetDup(), ack.GetMid())
	case PUBREC: //5
		//log.Debug("Ack To Client By PUBREL \n")
		//用于 Qos =2 的消息 回复 PUBREL
		ack := pAndErr.pack.GetVariable().(*Puback)
		err = c.queue.WritePack(GetPubRELPack(ack.GetMid()))
	case PUBREL: //6
		//log.Debug("Ack To Client By PUBCOMP \n")
		//用于 Qos =2 的消息 回复 PUBCOMP
		ack := pAndErr.pack.GetVariable().(*Puback)
		err = c.queue.WritePack(GetPubCOMPPack(ack.GetMid()))
	case PUBCOMP: //7
		//消息发送端最终确认这条消息
		//log.Debug("消息最终确认")
	case SUBSCRIBE: //7
		//消息发送端最终确认这条消息
		sub := pAndErr.pack.GetVariable().(*Subscribe)
		for _, top := range sub.GetTopics() {
			//log.Debug("Subscribe %s",*top.GetName())
			if top.Qos == 2 {
				//log.Debug("Ack To Client By Suback \n")
				//用于 Qos =2 的消息 回复 PUBCOMP
				err = c.queue.WritePack(GetSubAckPack(sub.GetMid()))
			}
		}
		//目前这个版本暂时先不保证消息的Qos 默认用Qos=1吧
		c.recover.OnRecover(pAndErr.pack)
	case UNSUBSCRIBE: //7
		//消息发送端最终确认这条消息
		sub := pAndErr.pack.GetVariable().(*UNSubscribe)
		err = c.queue.WritePack(GetUNSubAckPack(sub.GetMid()))
		//目前这个版本暂时先不保证消息的Qos 默认用Qos=1吧
		c.recover.OnRecover(pAndErr.pack)
	case PINGREQ:
		// Reply the heart beat
		log.Debug("hb msg")
		err = c.queue.WritePack(GetPingResp(1, pAndErr.pack.GetDup()))
		c.recover.OnRecover(pAndErr.pack)
	default:
		// Not define pack type
		//log.Debug("其他类型的数据包")
		//err = fmt.Errorf("The type not define:%v\n", pAndErr.pack.GetType())
	}
	return
}

func (c *Client) waitQuit() {
	// Start close
	log.Info("Will break new relogin")
	c.isSendClose = true
}

func (c *Client) pushMsg(pack *Pack) error {
	// Write this pack
	err := c.queue.WritePack(pack)
	return err
}

func (c *Client) WriteMsg(topic string, body []byte) error {
	if c.isStop {
		return fmt.Errorf("connection is closed")
	}
	pack := GetPubPack(0, 0, c.getOnlineMsgId(), &topic, body)
	return c.pushMsg(pack)
	//return nil
}
