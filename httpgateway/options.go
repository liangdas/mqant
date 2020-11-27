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

//Package httpgateway 网关配置
package httpgateway

import (
	"errors"
	"fmt"
	"github.com/liangdas/mqant/module"
	"github.com/liangdas/mqant/registry"
	"github.com/liangdas/mqant/selector"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Service represents an API service
type Service struct {
	// hander
	Hander string
	// node
	SrvSession module.ServerSession
}

// DefaultRoute 默认路由规则
var DefaultRoute = func(app module.App, r *http.Request) (*Service, error) {
	if r.URL.Path == "" {
		return nil, errors.New("path is nil")
	}
	handers := strings.Split(r.URL.Path, "/")
	if len(handers) < 2 {
		return nil, errors.New("path is not /[server]/path")
	}
	server := handers[1]
	if server == "" {
		return nil, errors.New("server is nil")
	}
	session, err := app.GetRouteServer(server,
		selector.WithStrategy(func(services []*registry.Service) selector.Next {
			var nodes []*registry.Node

			// Filter the nodes for datacenter
			for _, service := range services {
				for _, node := range service.Nodes {
					nodes = append(nodes, node)
				}
			}

			var mtx sync.Mutex
			//log.Info("services[0] $v",services[0].Nodes[0])
			return func() (*registry.Node, error) {
				mtx.Lock()
				defer mtx.Unlock()
				if len(nodes) == 0 {
					return nil, fmt.Errorf("no node")
				}
				index := rand.Intn(int(len(nodes)))
				return nodes[index], nil
			}
		}),
	)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	return &Service{SrvSession: session, Hander: r.URL.Path}, err
}

// Route 路由器定义
type Route func(app module.App, r *http.Request) (*Service, error)

// Option 配置
type Option func(*Options)

// Options 网关配置项
type Options struct {
	TimeOut time.Duration
	Route   Route
}

// NewOptions 创建配置
func NewOptions(app module.App, opts ...Option) Options {
	opt := Options{
		Route:   DefaultRoute,
		TimeOut: app.Options().RPCExpired,
	}

	for _, o := range opts {
		o(&opt)
	}

	return opt
}

// SetRoute 设置路由器
func SetRoute(s Route) Option {
	return func(o *Options) {
		o.Route = s
	}
}

// TimeOut 设置网关超时时间
func TimeOut(s time.Duration) Option {
	return func(o *Options) {
		o.TimeOut = s
	}
}
