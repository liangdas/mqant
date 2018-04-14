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
package basemodule

import (
	"github.com/liangdas/mqant/conf"
	"github.com/liangdas/mqant/log"
	"github.com/liangdas/mqant/module"
	"github.com/liangdas/mqant/rpc"
	"github.com/liangdas/mqant/rpc/base"
)

type rpcserver struct {
	settings *conf.ModuleSettings
	server   mqrpc.RPCServer
}

func (s *rpcserver) GetId() string {
	return s.settings.Id
}
func (s *rpcserver) OnInit(module module.Module, app module.App, settings *conf.ModuleSettings) {
	s.settings = settings
	server, err := defaultrpc.NewRPCServer(app, module) //默认会创建一个本地的RPC
	if err != nil {
		log.Warning("Dial: %s", err)
	}

	if settings.Rabbitmq != nil {
		//存在远程rpc的配置
		server.NewRabbitmqRPCServer(settings.Rabbitmq)
	}
	if settings.Redis != nil {
		//存在远程rpc的配置
		server.NewRedisRPCServer(settings.Redis)
	}
	if settings.UDP != nil {
		//存在远程rpc的配置
		server.NewUdpRPCServer(settings.UDP)
	}
	s.server = server
	err = app.RegisterLocalClient(settings.Id, server)
	if err != nil {
		log.Warning("RegisterLocalClient: id(%s) error(%s)", settings.Id, err)
	}
	log.Info("RPCServer init success id(%s) version(%s)", s.settings.Id, module.Version())
}
func (s *rpcserver) OnDestroy() {
	if s.server != nil {
		log.Info("RPCServer closeing id(%s)", s.settings.Id)
		err := s.server.Done()
		if err != nil {
			log.Warning("RPCServer close fail id(%s) error(%s)", s.settings.Id, err)
		} else {
			log.Info("RPCServer close success id(%s)", s.settings.Id)
		}
		s.server = nil
	}
}

func (s *rpcserver) Register(id string, f interface{}) {
	if s.server == nil {
		panic("invalid RPCServer")
	}
	s.server.Register(id, f)
}

func (s *rpcserver) RegisterGO(id string, f interface{}) {
	if s.server == nil {
		panic("invalid RPCServer")
	}
	s.server.RegisterGO(id, f)
}

func (s *rpcserver) GetRPCServer() mqrpc.RPCServer {
	if s.server == nil {
		panic("invalid RPCServer")
	}
	return s.server
}
