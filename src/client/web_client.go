package main

import (
	"github.com/astaxie/beego"
	"os"
	"fmt"
)

type MainController struct {
	beego.Controller
}


func (this *MainController) Get() {
	this.Ctx.WriteString("hello world")
}



func main() {
	beego.Router("/", &MainController{})
	dir,_:=os.Getwd()
	beego.SetStaticPath("/mqant",fmt.Sprintf("%s/views",dir))
	beego.SetStaticPath("/mqant","/work/go/mqant/src/client/views") //golang 工作目录问题还没有搞清楚怎么弄,在编译器编译的工作路径变了
	beego.Run("127.0.0.1:8080")
}
