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
	"github.com/garyburd/redigo/redis"
	"time"
)
var factory *RedisFactory

func GetRedisFactory() *RedisFactory {
	if factory==nil{
		factory=&RedisFactory{
			pools:NewBeeMap(),
		}
	}
	return factory
}
type RedisFactory struct{
	pools *BeeMap
}

func (this RedisFactory)GetPool(url string) (*redis.Pool) {
	if pool,ok:=this.pools.Items()[url];ok{
		return pool.(*redis.Pool)
	}
	pool := &redis.Pool{
		MaxIdle:     30,
		IdleTimeout: 240 * time.Second,
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
	if pool!=nil{
		this.pools.Set(url,pool)
	}
	return pool
}
func (this RedisFactory)CloseAllPool(){
	for _,pool:=range this.pools.Items(){
		pool.(*redis.Pool).Close()
	}
	this.pools.DeleteAll()
}