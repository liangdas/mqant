### 说明
mqant中的Gate网关模块相对来说非常重要,它有以下几个功能:

1. 与客户端建立长连接,心跳包检测等
2. 通过TCP，websocket通信方式(未来可新增UDP等)
3. MQTT数据包解析和发送
4. 客户端请求的路由工作

## Gate网关模块的使用
gate网关模块包含的功能虽然多,但在实际开发时并不需要做过多的二次开发,开发者只需要继承gate.Gate这个基础模块即可,如下:

	type Gate struct {
	gate.Gate	//继承
	}
	func (gate *Gate) GetType()(string){
		//很关键,需要与配置文件中的Module配置对应
		return "Gate"
	}
	
	func (gate *Gate) OnInit(app module.App,settings *conf.ModuleSettings) {
		//注意这里一定要用 gate.Gate 而不是 module.BaseModule
		gate.Gate.OnInit(gate,app,settings)
		gate.Gate.SetStorageHandler(gate)	//设置持久化处理器
	}

### 客户端消息的路由和转发

[通信协议与客户端支持介绍](https://github.com/liangdas/mqant/wiki/%E9%80%9A%E4%BF%A1%E5%8D%8F%E8%AE%AE%E4%B8%8E%E5%AE%A2%E6%88%B7%E7%AB%AF%E6%94%AF%E6%8C%81%E4%BB%8B%E7%BB%8D)

### Session
> 注意v1.3.0中Session rpc传送已经由map改为protobuf编码过的byte[],并且会由rpc接口自动转换,无需手动gate.NewSession

Session是由Gate网关模块维护的,在客户端向后端业务模块发送消息是,Gate网关模块负责路由并转发,在转发消息时会默认将该连接的Session信息发送给后端业务模块

它的大致字段如下：

	{
	    Userid			string	
	    IP				string
	    Network			string  
	    Sessionid		string				
	    Serverid    	string				
	    Settings		<key-value map> 	
	}

1. Userid			
	需要调用Bind绑定来设置 默认为"" 当客户端登陆以后可以设置该参数,其他业务模块通过判断Userid来判断该连接是否合法
2.	IP		
	客户端IP地址
3.	Network		
	网络类型 TCP websocket ...
4.	Sessionid		
	Gate网关模块生成的该连接唯一ID
5. Serverid    	
	Gate网关模块唯一ID，后端模块可以通过它来与Gate网关进行RPC调用
6. Settings		
	可以给这个连接设置一些参数,例如当用户加入对战房间以后可以设置一个参数
	roomName="mqant"
	
mqant为Session封装了Session类, Session类实现了一系列的方法,开发者可以实例化该类实现各种Session操作:

	session:=gate.NewSession(m.App,sessionMap)
	
	gate.Session实现的部分方法
	
	Bind //绑定UserID
	UnBind //解绑UserID
	Set(key string, value string) //设置一个参数 Push()以后才能发送到网关
	Push()	//将模块设置的Session信息发送到网关
	...
	Send(topic  string,body []byte) //给这个链接发送一个消息
	...

### Session持久化
如果我们不希望客户端网络中断以后导致Session数据丢失,以保证客户端在指定时间内重连以后能继续使用这些数据,我们需要对用户的Session进行持久化

Gate网关模块目前提供一个接口用来控制Session的持久化,但具体的持久化方式需要开发者自己来实现

    1. 实现持久化接口
    
	/**
	Session信息持久化
	 */
	type gate.StorageHandler interface {
		/**
		存储用户的Session信息
		Session Bind Userid以后每次设置 settings都会调用一次Storage
		 */
		Storage(Userid string,settings map[string]interface{})(err error)
		/**
		强制删除Session信息
		 */
		Delete(Userid string)(err error)
		/**
		获取用户Session信息
		Bind Userid时会调用Query获取最新信息
		 */
		Query(Userid string)(settings map[string]interface{},err error)
		/**
		用户心跳,一般用户在线时1s发送一次
		可以用来延长Session信息过期时间
		 */
		Heartbeat(Userid string)
	}
	
	2. 设置持久化处理器
	gate.Gate.SetStorageHandler(gate)

