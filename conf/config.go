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

package conf

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
)

var (
	// LenStackBuf 异常堆栈信息
	LenStackBuf = 1024
	// Conf 配置结构体
	Conf = Config{}
)

// LoadConfig 加载配置
func LoadConfig(Path string) {
	// Read config.
	if err := readFileInto(Path); err != nil {
		panic(err)
	}
	if Conf.RPC.RPCExpired == 0 {
		Conf.RPC.RPCExpired = 3
	}
	if Conf.RPC.MaxCoroutine == 0 {
		Conf.RPC.MaxCoroutine = 100
	}
}

// Config 配置结构体
type Config struct {
	Log      map[string]interface{}
	BI       map[string]interface{}
	OP       map[string]interface{}
	RPC      RPC `json:"rpc"`
	Module   map[string][]*ModuleSettings
	Mqtt     Mqtt
	Settings map[string]interface{}
}

// rpc 进程间通信配置
type RPC struct {
	MaxCoroutine int  //模块同时可以创建的最大协程数量默认是100
	RPCExpired   int  `json:"RpcExpired"` //远程访问最后期限值 单位秒[默认5秒] 这个值指定了在客户端可以等待服务端多长时间来应答
	Log          bool //是否打印RPC的日志
}

// ModuleSettings 模块配置
type ModuleSettings struct {
	ID        string `json:"ID"`
	Host      string
	ProcessID string
	Settings  map[string]interface{}
}

// Mqtt mqtt协议配置
type Mqtt struct {
	WirteLoopChanNum int // Should > 1 	    // 最大写入包队列缓存
	ReadPackLoop     int // 最大读取包队列缓存
	ReadTimeout      int // 读取超时
	WriteTimeout     int // 写入超时
}

func readFileInto(path string) error {
	var data []byte
	buf := new(bytes.Buffer)
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	r := bufio.NewReader(f)
	for {
		line, err := r.ReadSlice('\n')
		if err != nil {
			if len(line) > 0 {
				buf.Write(line)
			}
			break
		}
		if !strings.HasPrefix(strings.TrimLeft(string(line), "\t "), "//") {
			buf.Write(line)
		}
	}
	data = buf.Bytes()
	//fmt.Print(string(data))
	return json.Unmarshal(data, &Conf)
}

// If read the file has an error,it will throws a panic.
func fileToStruct(path string, ptr *[]byte) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	*ptr = data
}
