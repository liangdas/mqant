# mqant与客户端通信的数据结构
mqant是支持与客户端双向通信的长连接框架,与客户端通信有以下几种情况

1. 客户端主动发起请求

	类似http的Request--Response模式，一问一答。

2. 服务器主动给客户端发送消息

	与app的推送功能相似
	
> 以上两种通信实现原理以及数据结构有所区别,下面将详细说明


## 客户端主动发起请求

1. 首先后端服务器会创建一个模块

	eg. User
2. 然后会在这个模块中定义一系列的功能函数

		eg.
		self.GetServer().RegisterGO("HD_LoginWithPassword", self.hdLoginWithPassword)
		self.GetServer().RegisterGO("HD_LoginWithToken", self.hdLoginWithToken)
		
函数的具体实现如下

	func (m *User) login(session gate.Session, msg map[string]interface{}) (result string, err string) {
		if msg["userName"] == nil || msg["passWord"] == nil {
			result = "userName or passWord cannot be nil"
			return
		}
	
		userName := msg["userName"].(string)
		err = session.Bind(userName)
		if err != "" {
			return
		}
		session.Set("login", "true")
		session.Push() //推送到网关
		return fmt.Sprintf("login success %s", userName), ""
	}

session :客户端在网关的唯一标识

msg     :客户端给该接口发送的数据
	
>msg 可以是map 也可以是 []byte 
	
例如部分客户端与服务端通信可能使用protobuf协议,那么msg就是[]byte,
	函数内可以通过[]byte--> protobuf
	
	func (m *User) login(session gate.Session, msg []byte) ( []byte,  string) {
		//解析客户端发送过来的user.LoginRequest结构体
		request:=&user.LoginRequest{}
		proto.UnmarshalMerge(msg, request)
		/////
	
		//这里开始登陆处理等操作
	
		/////
	
	}

	
##### 服务端回复客户端(Response) *重点*

首先服务端回复客户端结果非常简单,就是执行函数return执行结果即可

eg.

	func (m *Login) login(session gate.Session, msg map[string]interface{}) (result string, err string) {
		if msg["userName"] == nil || msg["passWord"] == nil {
			result = "userName or passWord cannot be nil"
			return
		}
	
		userName := msg["userName"].(string)
		err = session.Bind(userName)
		if err != "" {
			return
		}
		session.Set("login", "true")
		session.Push() //推送到网关
		return fmt.Sprintf("login success %s", userName), ""
	}
	
以上函数返回值有两个  result string, err string

result 代表函数执行正确时返回的结果

	可以是map,[]byte,string等类型

err 代表函数执行错误时返回的错误描述
   
   	err 只能是string类型
   	
#### 如何组织result和err让客户端能正确解析呢？
> 服务端与客户端必须要提前约定好封装的数据结构,只有这样客户端才能正确解析返回结果

##### mqant默认封装规则
> mqant默认封装为json结构体,具体结构如下

	{
		Error string
		Result interface{}
	}
	
流程:

后端模块Error string,Result interface{}--rpc-->Gate【组装为json】【json转为[]byte】---->客户端[]byte转json

##### 默认规则的不足

mqant默认规则使用json来封装，但实际情况下不同的游戏可能需要的封装数据格式有所不同,例如有些游戏倾向于用protobuf封装。因此需要自定义封装规则

### mqant自定义封装规则
> mqant实现自定义的封装规则非常简单，仅需要注册一个全局函数即可


	app.SetProtocolMarshal(func(Result interface{},Error string)(module.ProtocolMarshal,string){
			//下面可以实现你自己的封装规则(数据结构)
			r := &resultInfo{
				Error:  Error,
				Result: Result,
			}
			b,err:= json.Marshal(r)
			if err == nil {
				//解析得到[]byte后用NewProtocolMarshal封装为module.ProtocolMarshal
				return app.NewProtocolMarshal(b),""
			} else {
				return nil,err.Error()
			}
		})

*完活*

### 返回值与性能

如下 mqant默认返回值是这样的 result map[string][string], err string

	func (m *Login) login(session gate.Session, msg map[string]interface{}) (result map[string][string], err string) {
		。。。
		return map[string][string]{
			"info":"login success"
		}, ""
	}
	
rpc通信编码/解密流程:  map—>[]byte ——— []byte—>map—>[]byte—> client

最终发送给客户端的是[]byte类型，中间经历一次 []byte—>map—>[]byte 的无用编解码流程

##### 如何省掉无效的编解码流程呢？

答案: 在后端模块提前编码为[]byte类型，如何实现见以下代码:

	func (this *Login) login(session gate.Session, msg map[string]interface{}) (result module.ProtocolMarshal, err string) {
	。。。
		return this.App.ProtocolMarshal("login success","") 
	}

实现原理:

返回值用 this.App.ProtocolMarshal 函数封装一遍即可，返回值改为了module.ProtocolMarshal类型

## 服务器主动给客户端发送消息
这种通信模式也非常普通,日常开发过程中都随时能遇到

#### mqant如何实现主动给客户端发消息?

答案在 gate.Session这个类里面,见以下代码

	func (m *Login) login(session gate.Session, msg []byte) ( []byte,  string) {
		//解析客户端发送过来的user.LoginRequest结构体
		request:=&user.LoginRequest{}
		proto.UnmarshalMerge(msg, request)
		/////
	
	
		//这里开始登陆处理等操作
	
		/////
	
		//组建处理结果数据包
		datamsg,err:=proto.Marshal(&user.LoginSuccessResponse{})
		if err!=nil{
			log.Error(err.Error())
			return nil, ""
		}
	
		//给客户端主动发送处理结果  
		errstr:=session.Send("Login/Success",datamsg)
		if errstr!=""{
			log.Error(errstr)
			return nil, errstr
		}
	
		return nil, ""
	}
	

session.Send(topic string, datamsg []byte)

topic 代表一个消息体标识，可以约定让客户端知道datamsg是一个什么数据格式

datamsg 序列化过后的二进制数据流


#### 至此mqant与客户端通信数据结构已介绍完了,如有疑问请在[mqant官方论坛](www.mqant.com)或mqant QQ群中留言
