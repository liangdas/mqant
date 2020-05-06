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
package gate

import (
	"github.com/liangdas/mqant/server"
	"time"
)

type Option func(*Options)

type Options struct {
	ConcurrentTasks int
	BufSize         int
	MaxPackSize     int
	Tls				bool
	TcpAddr			string
	WsAddr			string
	CertFile		string
	KeyFile			string
	Heartbeat       time.Duration
	OverTime        time.Duration
	RouteHandler    RouteHandler
	StorageHandler  StorageHandler
	AgentLearner    AgentLearner
	SessionLearner  SessionLearner
	GateHandler     GateHandler
	SendMessageHook SendMessageHook
	Opts			[]server.Option
}

func NewOptions(opts ...Option) Options {
	opt := Options{
		Opts:[]server.Option{},
		ConcurrentTasks: 20,
		BufSize:         2048,
		MaxPackSize:     65535,
		Heartbeat:       time.Minute,
		OverTime:        time.Second * 10,
		Tls:false,
	}

	for _, o := range opts {
		o(&opt)
	}

	return opt
}

func ConcurrentTasks(s int) Option {
	return func(o *Options) {
		o.ConcurrentTasks = s
	}
}
func BufSize(s int) Option {
	return func(o *Options) {
		o.BufSize = s
	}
}
func MaxPackSize(s int) Option {
	return func(o *Options) {
		o.MaxPackSize = s
	}
}
func Heartbeat(s time.Duration) Option {
	return func(o *Options) {
		o.Heartbeat = s
	}
}

func OverTime(s time.Duration) Option {
	return func(o *Options) {
		o.OverTime = s
	}
}

func SetRouteHandler(s RouteHandler) Option {
	return func(o *Options) {
		o.RouteHandler = s
	}
}
func SetStorageHandler(s StorageHandler) Option {
	return func(o *Options) {
		o.StorageHandler = s
	}
}
func SetAgentLearner(s AgentLearner) Option {
	return func(o *Options) {
		o.AgentLearner = s
	}
}
func SetGateHandler(s GateHandler) Option {
	return func(o *Options) {
		o.GateHandler = s
	}
}

func SetSessionLearner(s SessionLearner) Option {
	return func(o *Options) {
		o.SessionLearner = s
	}
}

func SetSendMessageHook(s SendMessageHook) Option {
	return func(o *Options) {
		o.SendMessageHook = s
	}
}

func Tls(s bool) Option {
	return func(o *Options) {
		o.Tls = s
	}
}

func TcpAddr(s string) Option {
	return func(o *Options) {
		o.TcpAddr = s
	}
}

func WsAddr(s string) Option {
	return func(o *Options) {
		o.WsAddr = s
	}
}

func CertFile(s string) Option {
	return func(o *Options) {
		o.CertFile = s
	}
}

func KeyFile(s string) Option {
	return func(o *Options) {
		o.KeyFile = s
	}
}

func ServerOpts(s []server.Option) Option {
	return func(o *Options) {
		o.Opts = s
	}
}
