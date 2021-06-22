package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/manifoldco/promptui"
)

const Template = `
package ${package}

import (
	"github.com/liangdas/mqant/conf"
	basemodule "github.com/liangdas/mqant/module/base"
)

type ${Template}Server struct {
	*basemodule.BaseModule
}

// //模块版本
func (${t} *${Template}Server) Version() string {
	return "v1.0.0"
}

//  //模块类型
func (${t} *${Template}Server) GetType() string {
	return "${package}"
}

//当App初始化时调用，这个接口不管这个模块是否在这个进程运行都会调用
func (${t} *${Template}Server) OnAppConfigurationLoaded(app *basemodule.BaseModule) {

}

//为以后动态服务发现做准备
func (${t} *${Template}Server) OnConfChanged(settings *conf.ModuleSettings) {

}
func (${t} *${Template}Server) OnInit(app *basemodule.BaseModule, settings *conf.ModuleSettings) {

}
func (${t} *${Template}Server) OnDestroy() {

}
func (${t} *${Template}Server) GetApp() *basemodule.BaseModule {
	return ${t}.BaseModule
}
func (${t} *${Template}Server) Run(closeSig chan bool) {

}
`

var (
	moduleName  = ""
	serviceName = ""
)

const banner = `
________          __           _________                                
\______ \ _____ _/  |______   /   _____/ ______________  __ ___________ 
 |    |  \\__  \\   __\__  \  \_____  \_/ __ \_  __ \  \/ // __ \_  __ \
 |    ,   \/ __ \|  |  / __ \_/        \  ___/|  | \/\   /\  ___/|  | \/
/_______  (____  /__| (____  /_______  /\___  >__|    \_/  \___  >__|
\/     \/          \/        \/     \/                 \/
VERSION: %s
`

func main() {
	fmt.Printf(banner, "v1.0.1")
	// 首先输入服务名
	moduleName = inputSevice("module")
	// 输入服务名
	serviceName = inputSevice("service")

	newModule := strings.ReplaceAll(Template, "${package}", moduleName)
	newModule = strings.ReplaceAll(newModule, "${Template}", serviceName)
	newModule = strings.ReplaceAll(newModule, "${t}", serviceName[0:1])
	os.Chdir("..")
	os.Mkdir(moduleName, os.ModePerm)
	os.Chdir(moduleName)
	os.Create("module.go")
	os.WriteFile("module.go", []byte(newModule), os.ModePerm)
	exec.Command("go fmt")
}

func inputSevice(name string) string {
	validate := func(input string) error {
		_, err := strconv.ParseFloat(input, 64)
		if err != nil {
			return nil
		}
		return errors.New("Invalid number")
	}
	prompt := promptui.Prompt{
		Label:    name,
		Validate: validate,
	}
	r, err := prompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return "0"
	}
	return r
}
