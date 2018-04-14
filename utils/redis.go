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
package utils

import (
	"github.com/gomodule/redigo/redis"
	"time"
)

var factory *RedisFactory

func GetRedisFactory() *RedisFactory {
	if factory == nil {
		factory = &RedisFactory{
			pools: NewBeeMap(),
		}
	}
	return factory
}

type RedisFactory struct {
	pools *BeeMap
}

func (this RedisFactory) GetPool(url string) *redis.Pool {
	if pool, ok := this.pools.Items()[url]; ok {
		return pool.(*redis.Pool)
	}
	pool := &redis.Pool{
		// 最大的空闲连接数，表示即使没有redis连接时依然可以保持N个空闲的连接，而不被清除，随时处于待命状态
		MaxIdle: 10,
		// 最大的激活连接数，表示同时最多有N个连接 ，为0事表示没有限制
		MaxActive: 100,
		//最大的空闲连接等待时间，超过此时间后，空闲连接将被关闭
		IdleTimeout: 240 * time.Second,
		// 当链接数达到最大后是否阻塞，如果不的话，达到最大后返回错误
		Wait: true,
		Dial: func() (redis.Conn, error) {
			c, err := redis.DialURL(url)
			if err != nil {
				return nil, err
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
	if pool != nil {
		this.pools.Set(url, pool)
	}
	return pool
}
func (this RedisFactory) CloseAllPool() {
	for _, pool := range this.pools.Items() {
		pool.(*redis.Pool).Close()
	}
	this.pools.DeleteAll()
}
