# 概述

mqant 2x实现了服务发现,理应支持服务发现的特性功能

# 服务选择(selector)

支持服务发现以后,服务节点可以随意的新增和删除,因此必然需要一套规则来选择需要调用的节点

例如:

1. 按权重轮询
2. 按节点当前负载
3. hash
4. 按节点元数据
5. 其他

## 全局默认规则设置
> mqant的服务发现参考go-micro,因此开发规则基本可以参考go-micro

    app.Options().Selector.Init(selector.SetStrategy(func(services []*registry.Service) selector.Next{
            var nodes []*registry.Node

            // Filter the nodes for datacenter
            for _, service := range services {
                for _, node := range service.Nodes {
                    if node.Metadata["type"] == "helloworld" {
                        nodes = append(nodes, node)
                    }
                }
            }

            var mtx sync.Mutex
            //log.Info("services[0] $v",services[0].Nodes[0])
            return func() (*registry.Node, error) {
                mtx.Lock()
                defer mtx.Unlock()
                index := rand.Intn(len(nodes))
                return nodes[index], nil
            }
        }))

## 单次RPC调用指定selector
> 在开发过程中难免使用某一些特殊规则获取服务节点

    m.RpcInvoke("user","/LoginByToken","token")  使用全局默认selector选择user节点
    m.RpcInvoke("user@nodeId","/LoginByToken","token") 指定调用nodeId节点


    //按自定义规则选择user 节点
    server,err:=m.GetRouteServer("user","",selector.WithStrategy(selector.SetStrategy(func(services []*registry.Service) selector.Next{
        var nodes []*registry.Node

        // Filter the nodes for datacenter
        for _, service := range services {
            for _, node := range service.Nodes {
                if node.Metadata["type"] == "helloworld" {
                    nodes = append(nodes, node)
                }
            }
        }

        var mtx sync.Mutex
        //log.Info("services[0] $v",services[0].Nodes[0])
        return func() (*registry.Node, error) {
            mtx.Lock()
            defer mtx.Unlock()
            index := rand.Intn(len(nodes))
            return nodes[index], nil
        }
    })))
    server.Call("/LoginByToken","token")



# 服务节点自定义服务发现
> 服务节点在初始化时会往服务发现模块注入相关信息,并且会定期同步元数据到服务发现模块中,其中的一些参数是可以设置的

## 设置服务节点同步心跳

    this.BaseModule.OnInit(this, app, settings,gate.Heartbeat(time.Second*10))

> 其他的可设置参数可参见 server.Option

## 更新元数据

元数据可以让其他节点获取到,可以作为节点之间获取对方当前状态的一个通道,元数据默认在下一次心跳时同步到服务发现模块

    self.GetServer().Options().Metadata["weight"]="10"  //节点自身评估可用度
	self.GetServer().Options().Metadata["runs"]="0"     //该节点当前服务的用户数

