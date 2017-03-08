// Copyright 2014 mqantserver Author. All Rights Reserved.
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
package master
import (
	"github.com/liangdas/mqant/log"
	"fmt"
	"io/ioutil"
	"os/exec"
)

func cmd(remoteHost string, sourceDir string,targetDir string){
	// 执行系统命令
	// 第一个参数是命令名称
	// 后面参数可以有多个，命令参数
	//scp  -r /work/go/mqantserver/bin/* root@123.56.166.90:/opt/go/mqantserver
	cmd := exec.Command("scp", sourceDir, fmt.Sprintf("%s:%s",remoteHost,targetDir))
	// 获取输出对象，可以从该对象中读取输出结果
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Warning("",err)
	}
	// 保证关闭输出流
	defer stdout.Close()
	// 运行命令
	if err := cmd.Start(); err != nil {
		log.Warning("",err)
	}
	// 读取输出结果
	opBytes, err := ioutil.ReadAll(stdout)
	if err != nil {
		log.Warning("",err)
	}
	log.Info(string(opBytes))
}
