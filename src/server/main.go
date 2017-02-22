package main
import (
	"github.com/liangdas/mqant/conf"
	"github.com/liangdas/mqant/gate"
	"github.com/liangdas/mqant"
	"server/chat"
	"flag"
	"server/login"
)
//func ChatRoute( app module.App,moduleType string,serverId string,Type string) (*module.ServerSession){
//	//演示多个服务路由 默认使用第一个Server
//	log.Debug("Type:%s Id:%s 将要调用 type : %s",moduleType,serverId,Type)
//	servers:=app.GetServersByType(Type)
//	if len(servers)==0{
//		return nil
//	}
//	return servers[0]
//}
func main() {
	confPath:= flag.String("path","conf/server.conf", "")
	conf.LoadConfig(*confPath) //加载配置文件
	app:=mqant.CreateApp()
	app.Configure(conf.Conf)
	//app.Route("Chat",ChatRoute)
	app.Run(gate.Module(),login.Module(),chat.Module())

}
