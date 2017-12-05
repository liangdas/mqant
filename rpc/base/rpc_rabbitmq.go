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
package defaultrpc

import (
	"fmt"
	"github.com/liangdas/mqant/conf"
	"github.com/liangdas/mqant/log"
	"github.com/streadway/amqp"
	"time"
)

var (
	TypeServer = 0
	TypeClient = 1
)

type RabbitAgent struct {
	info           *conf.Rabbitmq
	wconn          *amqp.Connection
	rconn          *amqp.Connection
	wchannel       *amqp.Channel
	rchannel       *amqp.Channel
	rCloseError    chan *amqp.Error
	wCloseError    chan *amqp.Error
	readMsg        chan amqp.Delivery
	callback_queue string
	shutdown       bool //是否停止整个服务
	serverType     int  //服务类型
}

func NewRabbitAgent(info *conf.Rabbitmq, serverType int) (*RabbitAgent, error) {
	this := new(RabbitAgent)
	this.info = info
	this.serverType = serverType
	this.rCloseError = make(chan *amqp.Error)
	this.wCloseError = make(chan *amqp.Error)
	this.readMsg = make(chan amqp.Delivery, 1)
	this.shutdown = false
	err := this.WConnect()
	if err != nil {
		return nil, err
	}
	err = this.RConnect()
	if err != nil {
		return nil, err
	}
	return this, nil
}
func (this *RabbitAgent) ReadMsg() chan amqp.Delivery {
	return this.readMsg
}
func (this *RabbitAgent) Closed() bool {
	return this.shutdown
}

/**
停止服务
*/
func (this *RabbitAgent) Shutdown() error {
	var err error
	this.shutdown = true
	if this.wchannel != nil {
		this.wchannel.Close()
	}
	if this.wconn != nil {
		this.wconn.Close()
	}
	if this.rchannel != nil {
		this.rchannel.Close()
	}
	if this.rconn != nil {
		this.rconn.Close()
	}
	return err
}

/**
创建一个写连接
*/
func (this *RabbitAgent) WConnect() error {
	var err error
	log.Info("dialing %q", this.info.Uri)
	this.wconn, err = amqp.Dial(this.info.Uri) //打开连接
	if err != nil {
		return err
	}
	this.wconn.NotifyClose(this.wCloseError)
	go this.on_w_disconnect(this.wCloseError)

	return err
}

/**
创建一个读连接
*/
func (this *RabbitAgent) RConnect() error {
	var err error
	log.Info("dialing %q", this.info.Uri)
	this.rconn, err = amqp.Dial(this.info.Uri) //打开连接
	if err != nil {
		return err
	}
	if this.serverType == TypeServer {
		this.ExchangeDeclare() //创建交换器
		this.Queue()           //创建接收请求队列
	} else if this.serverType == TypeClient {
		this.CallbackQueue() //创建接收回调队列
	}
	this.rconn.NotifyClose(this.rCloseError)
	go this.on_r_disconnect(this.rCloseError)
	return err
}

/**
声明一个交换器,提供给rpc server端调用
*/
func (this *RabbitAgent) ExchangeDeclare() error {
	channel, err := this.RChannel()
	if err != nil {
		return err
	}
	return channel.ExchangeDeclare(
		this.info.Exchange,     // name of the exchange
		this.info.ExchangeType, // type
		true,  // durable
		false, // delete when complete
		false, // internal
		false, // noWait
		nil,   // arguments
	)
}

/**
获取读通道
*/
func (this *RabbitAgent) RChannel() (*amqp.Channel, error) {
	var err error
	if this.shutdown {
		//业务停止了
		return nil, fmt.Errorf("rabbit shutdown")
	}
	if this.rchannel == nil {
		this.rchannel, err = this.rconn.Channel()
	}
	return this.rchannel, err
}

/**
获取写通道
*/
func (this *RabbitAgent) WChannel() (*amqp.Channel, error) {
	var err error
	if this.shutdown {
		//业务停止了
		return nil, fmt.Errorf("rabbit shutdown")
	}
	if this.wchannel == nil {
		this.wchannel, err = this.wconn.Channel()
	}
	return this.wchannel, err
}

/**
rpc client写数据
*/
func (this *RabbitAgent) ClientPublish(body []byte) error {
	channel, err := this.WChannel()
	if err != nil {
		return err
	}
	callback_queue, err := this.CallbackQueue()
	if err != nil {
		return err
	}
	return channel.Publish(
		this.info.Exchange,   // publish to an exchange
		this.info.BindingKey, // routing to 0 or more queues
		false,                // mandatory
		false,                // immediate
		amqp.Publishing{
			Headers:         amqp.Table{"reply_to": callback_queue},
			ContentType:     "text/plain",
			ContentEncoding: "",
			Body:            body,
			DeliveryMode:    amqp.Transient, // 1=non-persistent, 2=persistent
			Priority:        0,              // 0-9
			// a bunch of application/implementation-specific fields
		},
	)
}

/**
rpc server写数据
*/
func (this *RabbitAgent) ServerPublish(queueName string, body []byte) error {
	channel, err := this.WChannel()
	if err != nil {
		return err
	}
	return channel.Publish(
		"",        // publish to an exchange
		queueName, // routing to 0 or more queues
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			Headers:         amqp.Table{},
			ContentType:     "text/plain",
			ContentEncoding: "",
			Body:            body,
			DeliveryMode:    amqp.Transient, // 1=non-persistent, 2=persistent
			Priority:        0,              // 0-9
			// a bunch of application/implementation-specific fields
		},
	)
}

/**
创建回调 提供给 rpc client调用
*/
func (this *RabbitAgent) CallbackQueue() (string, error) {
	if this.callback_queue == "" {
		// 声明回调队列，再次声明的原因是，服务器和客户端可能先后开启，该声明是幂等的，多次声明，但只生效一次
		channel, err := this.RChannel()
		if err != nil {
			return "", err
		}
		queue, err := channel.QueueDeclare(
			"",    // name of the queue
			false, // durable	持久化
			true,  // delete when unused
			true,  // exclusive
			false, // noWait
			nil,   // arguments
		)
		if err != nil {
			return "", fmt.Errorf("Queue Declare: %s", err)
		}
		deliveries, err := channel.Consume(
			queue.Name,            // name
			this.info.ConsumerTag, // consumerTag,
			false, // noAck 自动应答
			false, // exclusive
			false, // noLocal
			false, // noWait
			nil,   // arguments
		)
		if err != nil {
			return "", fmt.Errorf("Queue Consume: %s", err)
		}
		this.callback_queue = queue.Name //设置callback_queue
		go this.on_handle(deliveries)
	}
	return this.callback_queue, nil
}

/**
创建回调 提供给 rpc server调用
*/
func (this *RabbitAgent) Queue() error {
	channel, err := this.RChannel()
	if err != nil {
		return err
	}
	queue, err := channel.QueueDeclare(
		this.info.Queue, // name of the queue
		true,            // durable	持久化
		false,           // delete when unused
		false,           // exclusive
		false,           // noWait
		nil,             // arguments
	)
	if err != nil {
		return fmt.Errorf("Queue Declare: %s", err)
	}

	//log.Printf("declared Queue (%q %d messages, %d consumers), binding to Exchange (key %q)",
	//	queue.Name, queue.Messages, queue.Consumers, key)

	if err = channel.QueueBind(
		queue.Name,           // name of the queue
		this.info.BindingKey, // bindingKey
		this.info.Exchange,   // sourceExchange
		false,                // noWait
		nil,                  // arguments
	); err != nil {
		return fmt.Errorf("Queue Bind: %s", err)
	}
	//log.Printf("Queue bound to Exchange, starting Consume (consumer tag %q)", c.tag)
	deliveries, err := channel.Consume(
		queue.Name,            // name
		this.info.ConsumerTag, // consumerTag,
		false, // noAck 自动应答
		false, // exclusive
		false, // noLocal
		false, // noWait
		nil,   // arguments
	)
	if err != nil {
		return fmt.Errorf("Queue Consume: %s", err)
	}
	go this.on_handle(deliveries)
	return nil
}

/**
监听连接中断
*/
func (this *RabbitAgent) on_r_disconnect(closeError chan *amqp.Error) {
	select {
	case <-closeError:
		if this.shutdown {
			//业务停止了，不需要重新连接
			return
		}
		time.Sleep(time.Second * 1) //一秒以后重试
		this.RConnect()
		break
	}
}

/**
监听连接中断
*/
func (this *RabbitAgent) on_w_disconnect(closeError chan *amqp.Error) {
	select {
	case <-closeError:
		if this.shutdown {
			//业务停止了，不需要重新连接
			return
		}
		time.Sleep(time.Second * 1) //一秒以后重试
		this.WConnect()
		break
	}
}

/**
监听消息
*/
func (this *RabbitAgent) on_handle(deliveries <-chan amqp.Delivery) {
	for {
		select {
		case d, ok := <-deliveries:
			if !ok {
				goto LForEnd
			} else {
				this.readMsg <- d
			}
		}
	}
LForEnd:
}
