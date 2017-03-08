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
	"fmt"
	"github.com/liangdas/mqant/conf"
	"github.com/liangdas/mqant/log"
	"github.com/liangdas/mqant/utils/x/crypto/ssh"
	"net"
	"os/exec"
	"strconv"
	"strings"
)

const (
	Stop = iota
	Runing
)

func isLocalIP(ip string) (bool, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return false, err
	}
	for i := range addrs {
		intf, _, err := net.ParseCIDR(addrs[i].String())
		if err != nil {
			return false, err
		}
		if net.ParseIP(ip).Equal(intf) {
			return true, nil
		}
	}
	return false, nil
}

type Process struct {
	master    conf.Master
	Process   *conf.Process
	State     int //运行状态
	isLocalIP bool
}

func (p *Process) Init(master conf.Master, process *conf.Process) (err error) {
	p.master = master
	p.Process = process
	b, err := isLocalIP(process.Host)
	if err != nil {
		return
	}
	p.isLocalIP = b
	log.Info("%s isLocal IP :%v", process.Host, p.isLocalIP)
	p.StateUpdate() //执行一次远程进程状态更新

	return nil
}

func (p *Process) Exec(name string, arg ...string) (output string, errput string) {
	if p.isLocalIP {
		return p.OSExec(name, arg...)
	} else {
		return p.SSHExec(name, arg...)
	}
}

func (p *Process) OSExec(name string, arg ...string) (output string, errput string) {
	cmdstr := name + " " + strings.Join(arg, " ")
	log.Info("os exec :%s", cmdstr)

	cmd := exec.Command(name, arg...)
	b, err := cmd.Output()
	if err != nil {
		errput = err.Error()
		return
	}
	output = string(b)
	return
}

func (p *Process) SSHExec(name string, arg ...string) (output string, errput string) {
	sshconf := p.master.GetSSH(p.Process.Host)
	if sshconf == nil {
		errput = fmt.Sprintf("No host() found in SSH configuration ", p.Process.Host)
		return
	}
	client, err := ssh.Dial("tcp", sshconf.GetSSHHost(), &ssh.ClientConfig{
		User: sshconf.User,
		Auth: []ssh.AuthMethod{ssh.Password(sshconf.Password)},
	})
	if err != nil {
		errput = err.Error()
		return
	}

	cmd := name + " " + strings.Join(arg, " ")
	log.Info("ssh[%s] exec :%s", sshconf.GetSSHHost(), cmd)
	session, err := client.NewSession()
	b, err := session.Output(cmd)
	if err != nil {
		errput = err.Error()
		return
	}
	output = string(b)
	defer session.Close()
	return
}

func (p *Process) StateUpdate() (errput string) {
	_, errput = p.FindPID()
	if errput == "" {
		p.State = Runing
	} else {
		p.State = Stop
	}
	return
}

func (p *Process) Start() (output string, errput string) {
	str := []string{}
	if p.Process.Args != nil {
		//自定义参数
		for k, v := range p.Process.Args {
			str = append(str, fmt.Sprintf("-%s=%s", k, v))
		}
	}
	nohup := p.Process.LogDir + "/" + p.Process.ProcessID + ".nohup.log"
	p.Exec("mkdir", p.Process.LogDir)
	output, errput = p.Exec("nohup", p.Process.Execfile, "-pid="+p.Process.ProcessID, strings.Join(str, " "), "-log="+p.Process.LogDir, ">", nohup, "2>&1", "&")
	return
}

func (p *Process) Stop() (output string, errput string) {
	//根据运行文件路径查询PID
	pid, errput := p.FindPID()
	if errput == "" {
		////杀死进程
		output, errput = p.Exec("kill", "-SIGTERM", fmt.Sprintf("%d", pid))
	}
	return
}

func (p *Process) FindPID() (pid int, errput string) {
	//根据运行文件路径查询PID
	//ps -ef|grep ProcessID|grep -v grep|grep -v PPID|awk '{ print $2}'
	output, errput := p.Exec("ps", "-ef|grep", fmt.Sprintf("%s|grep", p.Process.ProcessID), "-v", "grep|grep", "-v", "PPID|awk", "'{ print $2}'")
	if output != "" {
		// 去除空格
		output = strings.Replace(output, " ", "", -1)
		// 去除换行符
		output = strings.Replace(output, "\n", "", -1)
		pid, err := strconv.Atoi(output)
		if err != nil {
			errput = err.Error()
		}
		return pid, errput
	}
	return
}

func (p *Process) Restart() error {
	p.Stop()  //先停止进程
	p.Start() //
	return nil
}
