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
	LenStackBuf=1024
	LogLevel="debug"
	LogPath=""
	LogFlag=0
	RpcExpired=5   //远程访问最后期限值 单位秒[默认5秒] 这个值指定了在客户端可以等待服务端多长时间来应答
	Conf = Config{}
)

func LoadConfig(Path string) {
	// Read config.
	if err := readFileInto(Path); err != nil {
		panic(err)
	}

}

type Config struct {
	Module    map[string][]*ModuleSettings
	Mqtt      Mqtt
}

type Rabbitmq struct {
	Uri          string
	Exchange     string
	ExchangeType string
	Queue	     string
	BindingKey   string	//
	ConsumerTag  string	//消费者TAG
}
type ModuleSettings struct {
	Id 		string
	Host 		string
	Group 		string
	Settings 	map[string]interface{}
	Rabbitmq 	*Rabbitmq
}


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
