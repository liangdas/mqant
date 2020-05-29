# topic uri路由器
mqant gate网关的默认路由规则为

    [moduleType@moduleID]/[handler]/[msgid]

但这个规则不太灵活,因此设计了一套基于URI规则的topic路由规则。

# 基于uri协议的路由规则

    <scheme>://<user>:<password>@<host>:<port>/<path>;<params>?<query>#<fragment>

1. 可以充分利用uri公共库
2. 资源划分更加清晰明确


## 示例

见 <IM通信协议.md>

# 如何启用模块

## 创建一个UriRoute结构体

    route:=uriRoute.NewUriRoute(this,
        uriRoute.Selector(func(topic string, u *url.URL) (s module.ServerSession, err error) {
            moduleType:=u.Scheme
            nodeId:=u.Hostname()
            //使用自己的
            if nodeId=="modulus"{
                //取模
            }else if nodeId=="cache"{
                //缓存
            }else if nodeId=="random"{
                //随机
            }else{
                //
                //指定节点规则就是 module://[user:pass@]nodeId/path
                //方式1
                // moduleType=fmt.Sprintf("%v@%v",moduleType,u.Hostname())
                //方式2
                return this.GetRouteServer(moduleType,selector.WithFilter(selector.FilterEndpoint(nodeId)))
            }
            return this.GetRouteServer(moduleType)
        }),
        uriRoute.DataParsing(func(topic string, u *url.URL, msg []byte) (bean interface{}, err error) {
            //根据topic解析msg为指定的结构体
            //结构体必须满足mqant的参数传递标准
            //例如mqrpc.Marshaler
            //type Marshaler interface {
            //	Marshal() ([]byte, error)
            //	Unmarshal([]byte) error
            //	String() string
            //}
            return
        }),
        uriRoute.CallTimeOut(3*time.Second),
    )

## 替换默认的gate路由规则

    this.Gate.OnInit(this, app, settings,
        gate.SetRouteHandler(route),
        ...
    )

## 完结

