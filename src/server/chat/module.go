/**
一定要记得在confin.json配置这个模块的参数,否则无法使用
 */
package chat
import (
	"github.com/liangdas/mqant/module"
	"encoding/json"
	"github.com/liangdas/mqant/gate"
	"github.com/liangdas/mqant/conf"
	"github.com/liangdas/mqant/log"
)
var Module = func() (module.Module){
	chat := new(Chat)
	return chat
}


type Chat struct {
	app	module.App
	server *module.Server
	chats  map[string][]*gate.Session
}
func (m *Chat) GetType()(string){
	//很关键,需要与配置文件中的Module配置对应
	return "Chat"
}
func (m *Chat) GetServer() (*module.Server){
	if m.server==nil{
		m.server = new(module.Server)
	}
	return m.server
}

func (m *Chat) OnInit(app module.App,settings *conf.ModuleSettings) {
	//初始化模块
	m.app=app
	m.chats=map[string][]*gate.Session{}

	//创建一个远程调用的RPC
	m.GetServer().OnInit(app,settings)
	//注册远程调用的函数
	m.GetServer().RegisterGO("Handler_JoinChat",m.joinChat) //我们约定所有对客户端的请求都以Handler_开头
	m.GetServer().RegisterGO("Handler_Say",m.say) //我们约定所有对客户端的请求都以Handler_开头

}

func (m *Chat) Run(closeSig chan bool) {
	//运行模块
}

func (m *Chat) OnDestroy() {
	//注销模块
	//一定别忘了关闭RPC
	m.GetServer().OnDestroy()
}

func (m *Chat) joinChat(s map[string]interface{},msg map[string]interface{})(result string,err string) {
	if msg["roomName"]==nil{
		result="roomName cannot be nil"
		return
	}
	session:=gate.NewSession(m.app,s)
	log.Debug("session %v",session.ExportMap())
	if session.Userid==""{
		err="Not Logined"
		return
	}
	roomName:=msg["roomName"].(string)

	r,_:=m.app.RpcInvoke("Login","getRand",roomName)

	log.Debug("演示模块间RPC调用 :",r)

	userList:=m.chats["roomName"]
	if userList==nil{
		//添加一个新的房间
		userList=[]*gate.Session{session}
		m.chats["roomName"]=userList
	}else{
		for i,user:=range userList{
			if user.Userid==session.Userid{
				//已经加入过这个聊天室了 不过这里还是替换一下session 因此用户可能是重连的
				err="Already in this chat room"
				userList[i]=session
				return
			}
		}
		//添加这个用户进入聊天室
		userList=append(userList,session)
		m.chats["roomName"]=userList
	}

	rmsg:=map[string]string{}
	rmsg["roomName"]=roomName
	rmsg["userName"]=session.Userid
	b,_:=json.Marshal(rmsg)
	//广播添加用户信息到该房间的所有用户
	for _,user:=range userList{
		//这个不保证信息是否真的发送成功
		user.SendNR("Chat/OnJoin",b)

		//err:=user.Send("Chat/OnJoin",b)
		//if err!=nil{
		//	//信息没有发送成功
		//}
	}
	return "join success",""
}

func (m *Chat) say(s map[string]interface{},msg map[string]interface{})(result string,err string){
	if msg["roomName"]==nil||msg["say"]==nil{
		result="roomName or say cannot be nil"
		return
	}
	session:=gate.NewSession(m.app,s)
	if session.Userid==""{
		err="Not Logined"
		return
	}
	roomName:=msg["roomName"].(string)
	say:=msg["say"].(string)
	userList:=m.chats["roomName"]
	if userList==nil{
		err="No room"
		return
	}else{
		isJion:=false
		for _,user:=range userList{
			if user.Userid==session.Userid{
				//已经加入过这个聊天室了
				isJion=true
				break
			}
		}
		if !isJion{
			err="You haven't been in the room yet"
			return
		}

		rmsg:=map[string]string{}
		rmsg["roomName"]=roomName
		rmsg["userName"]=session.Userid
		rmsg["say"]=say
		b,_:=json.Marshal(rmsg)
		//广播添加用户信息到该房间的所有用户
		for _,user:=range userList{
			user.SendNR("Chat/OnSay",b)
		}
	}
	result="say success"
	return
}


