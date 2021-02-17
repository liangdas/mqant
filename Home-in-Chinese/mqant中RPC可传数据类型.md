## RPC可传数据类型

1. 调用可以使用的参数数据类型

		bool
		int
		long64
		float32
		float64
		[]byte
		string
		map[string]interface{}
		protocol buffer结构体
		自定义结构体

>注意调用参数不能为nil 如:
>result,err:=module.RpcInvoke(“user”,"login","mqant",nil) 会出现异常无法调用

2. 返回值可使用的参数类型

		result:

			bool
			int64
			long64
			float64
			string
			map[string]interface{}
			protocol buffer结构体
			自定义结构体

		err:
			string

		需要注意的是如果正常 err不能为nil要用“”
		pymqant也是不能用None要用""

## RPC自定义参数数据结构
> 我们常常需要在模块之间传输自定义的数据结构，目前mqantRPC支持两种模式使用自定义数据结构

### 方式一 实现protocolbuffer结构体(推荐)
这种方式只能在v1.0.0以上版本使用

#### 注意事项:

    1. proto.Message是protocol buffer约定的数据结构,因此需要双方都能够明确数据结构的类型（可以直接断言的）
    2. 服务函数返回结构一定要用指针(*rpcpb.ResultInfo)否则mqant无法识别 (见下文)

#### 如何使用？

1. 实现带proto.Message接口的数据结构(.proto生成的go结构体就是)


2. 在RPC函数直接使用即可


        func (self *MyModule) OnInit(app module.App, settings *conf.ModuleSettings) {
            self.BaseModule.OnInit(self, app, settings)
            self.GetServer().RegisterGO("testProto", self.testProto)
        }

        //函数返回结构一定要用指针(*rpcpb.ResultInfo)否则mqant无法识别
        func (self *MyModule)testProto(req *rpcpb.ResultInfo) (*rpcpb.ResultInfo,error) {
        	log.Info("testProto %v",req.Error)
        	r:=&rpcpb.ResultInfo{Error:*proto.String("hello Proto返回内容")}
        	return r,nil
        }

3. 调用函数

        protobean:=new(rpcpb.ResultInfo)
        err:=mqrpc.Proto(protobean, func() (reply interface{}, errstr interface{}) {
            return self.RpcInvoke("webapp","testProto",&rpcpb.ResultInfo{Error:*proto.String("hello 我是测试proto编码的")})
        })

        log.Info("RpcInvoke %v , err %v",protobean.Error,err)


### 方式二 实现mqrpc.Marshaler接口(推荐)
这种方式只能在v2.5以上版本使用

#### 注意事项:

    1. mqrpc.Marshaler是请求方和服务方约定的数据结构,因此需要双方都能够明确数据结构的类型（可以直接断言的）
    2. 服务函数返回结构一定要用指针(*rsp)否则mqant无法识别 (见下文)

#### 如何使用？

1. 实现带mqrpc.Marshaler接口的数据结构

        // 请求结构体
        type req struct {
            id string
        }
        func (this *req)Marshal() ([]byte, error){
            return []byte(this.id),nil
        }
        func (this *req)Unmarshal(data []byte) error{
            //可以自由使用自己的编码方式,比如pb,json等
            this.id=string(data)
            return nil
        }
        func (this *req)String() string{
            return "req"
        }

        //响应结构体
        type rsp struct {
            id string
        }
        func (this *rsp)Marshal() ([]byte, error){
            return []byte(this.id),nil
        }
        func (this *rsp)Unmarshal(data []byte) error{
            this.id=string(data)
            return nil
        }
        func (this *rsp)String() string{
            return "rsp"
        }

2. 在RPC函数直接使用即可


        func (self *MyModule) OnInit(app module.App, settings *conf.ModuleSettings) {
            self.BaseModule.OnInit(self, app, settings)
            self.GetServer().RegisterGO("sendMessage", self.sendMessage)
        }

        //函数返回结构一定要用指针(*rsp)否则mqant无法识别
        func (self *MyModule)sendMessage(req req) (*rsp,error) {
            log.Info("sendMessage %v",req.id)
            r:=&rsp{id:req.id}
            return r,nil
        }

3. 调用函数

        r:=new(rsp)
        err:=mqrpc.Marshal(r, func() (reply interface{}, errstr interface{}) {
            return self.RpcInvoke("webapp","sendMessage",&req{id:"hello 我是RpcInvoke"})
        })

        log.Info("RpcInvoke %v , err %v",r.id,err)

        //v2.5以后新增的RPC通信方式,RpcCall可以实现更丰富更灵活的RPC调用
        err=mqrpc.Marshal(r, func() (reply interface{}, errstr interface{}) {
            ctx,_:=context.WithTimeout(context.TODO(),time.Second*3)
            return self.RpcCall(ctx,"webapp","sendMessage",
                mqrpc.Param(&req{id:"hello 我是RpcCall"}),
                selector.WithFilter(func(services []*registry.Service) []*registry.Service {
                    log.Info("WithFilter")
                return services
                }),
            )
        })
        log.Info("RpcCall %v , err %v",r.id,err)

### 方式三 通过注册全局序列化映射表实现

这种方式只能 v1.3.0版本以上版本

#### 如何使用?

1. 实现module.RPCSerialize接口

		/**
		rpc 自定义参数序列化接口
		 */
		type RPCSerialize interface {
			/**
			序列化 结构体-->[]byte
			param 需要序列化的参数值
			@return ptype 当能够序列化这个值,并且正确解析为[]byte时 返回改值正确的类型,否则返回 ""即可
			@return p 解析成功得到的数据, 如果无法解析该类型,或者解析失败 返回nil即可
			@return err 无法解析该类型,或者解析失败 返回错误信息
			 */
			Serialize(param interface{})(ptype string,p []byte, err error)
			/**
			反序列化 []byte-->结构体
			ptype 参数类型 与Serialize函数中ptype 对应
			b   参数的字节流
			@return param 解析成功得到的数据结构
			@return err 无法解析该类型,或者解析失败 返回错误信息
			 */
			Deserialize(ptype string,b []byte)(param interface{},err error)
			/**
			返回这个接口能够处理的所有类型
			 */
			GetTypes()([]string)
		}

		eg

		/**
		自定义rpc参数序列化反序列化  Session
		 */
		func (gate *Gate)Serialize(param interface{})(ptype string,p []byte, err error){
			switch v2:=param.(type) {
			case Session:
				bytes,err:=v2.Serializable()
				if err != nil{
					return "SESSION",nil,err
				}
				return "SESSION",bytes,nil
			default:
				return "", nil,fmt.Errorf("args [%s] Types not allowed",reflect.TypeOf(param))
			}
		}

		func (gate *Gate)Deserialize(ptype string,b []byte)(param interface{},err error){
			switch ptype {
			case "SESSION":
				mps,errs:= NewSession(gate.App,b)
				if errs!=nil{
					return	nil,errs
				}
				return mps,nil
			default:
				return	nil,fmt.Errorf("args [%s] Types not allowed",ptype)
			}
		}

2. 将接口添加到系统中

		//添加Session结构体的序列化操作类
		err:=module.App.AddRPCSerialize("gate",gate)
		if err!=nil{
			log.Warning("Adding session structures failed to serialize interfaces",err.Error())
		}

如此便完成了,以后我们可以通过rpc直接传输gate.Session这个数据结构了
> 注意所有的数据结构名称(ptype)不能重复，rpc已内置了几个类型

	eg.
	result,e:=serverSession.Call("Login",gate.Session,"hello","world")