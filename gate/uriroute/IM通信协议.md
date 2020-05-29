# 通信协议

### http协议

1. 历史消息拉取
2. 群组列表拉取
3. server-to-server接口调用

### mqtt协议

基于tcp,websocket长连接

1. 消息下发


### mqtt协议的使用

mqant只用了mqtt 3.1版本的协议,目前主流开发语言都有相应的客户端开发库。

mqant只用到了mqtt的【发布消息】协议，订阅协议未支持

### mqtt协议简介

mqtt协议包可以简单的理解为就以下的这两个字段，不管是server-to-client 还是client-to-server

    -- topic    string 可以理解为路由
    -- body     []byte  消息体,是任意二进制流

### im模块中mqtt协议的用法

im模块中topic被限定为标准的URI格式

    <scheme>://<user>:<password>@<host>:<port>/<path>;<params>?<query>#<fragment>

### uri资源划分

##### scheme

1. http/https

        用于客户端http流量代理(httpproxy)

        client-mqtt->im-server-http->http-server

        eg.

        topic:

            http://www.wanwan.com/pay/one.json

        body:

            type HttpRequest struct {
                Header	map[string]string
                Body	string
                Method	string
            }

        return:

            type HttpResponse struct {
                Header			http.Header
                Body			interface{}
                Err				error
                StatusCode    	int
            }

2. local

    im模块预留私有协议

3. system

    im模块预留私有协议

4. 业务模块

    1. scheme  用于指定mqant后端微服务的模块ID

    2. hostname 用于指定业务模块节点路由规则

        >业务模块可能有多个节点

        1. modulus

            取模
        2. cache

            缓存
        3. random

            随机
        4. 其他规则

            例如节点ID

    eg.

       topic1:

            im://modulus/remove_feeds_member?msg_id=1002

       代表访问后端的im模块,按取模的规则选择im模块的一个节点

       访问方法 /remove_feeds_member

       msg_id=1002 用来唯一标识一次请求

        body:
             {
                "member":"要移除成员userId",
                "feeds":"feedsid"
             }

       topic2:

           im://d8feff3dc8daf472/agree_join_feeds?msg_id=1003

          代表访问后端的im模块的d8feff3dc8daf472节点

          访问方法 /remove_feeds_member

          msg_id=1002 用来唯一标识一次请求

          body:

             {
             	"applicant":"申请人userId",
             	"feeds":"feedsid"
             }


### 客户端如何接受服务端返回消息

## 请求-响应模式

客户端先发起一次请求,服务器在处理完后返回对应结果


### 客户端请求

topic:

    im://modulus/remove_feeds_member?msg_id=1002

body:

     {
        "member":"要移除成员userId",
        "feeds":"feedsid"
     }


### 服务器响应
> 请看,服务器响应消息时topic跟客户端请求时完全相同,
> 因此客户端只要保证topic每一次请求都是唯一的,并且记住它,
> 那么在服务器响应时就能找到正确的处理代码了
topic:

    im://modulus/remove_feeds_member?msg_id=1002

body:

    {
        "Trace":"94218104768aa033",
        "Error":"",
        "Result":"success"
	}

## 服务器PUSH模式
> 当用户A给用户B发消息是,im工具需要主动的通知用户B消息

这种情况一般我们会提前跟服务器约定好这类消息的topic

eg.

    通知feeds有新消息

    topic:

        impush://modulus/news

    body:

     {
        "feeds":"feeds id"
        "last_seq_id":int //最新消息的序列号
         "message":{
             "MsgId":string	//消息唯一ID
             "Feeds":string	//消息所属消息组
             "SeqId":int	//消息顺序ID
             "MsgType":string	//消息类型 text image 语音：voice 视频：video 小视频：shortvideo 地理位置：location 连接消息：link
             *"From":string	//谁发出的
             "Payload":Object	//谁发出的
             "TimeCreate":int	//消息创建时间(发送)
         }
     }