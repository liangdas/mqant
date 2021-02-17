# 概述

mqant从项目发起时的目标便是一套分布式的微服务架构,但因为精力有限一直没能开发出微服务的标配服务发现功能,
不过从2.0开始mqant也支持服务发现了(^_^)

# mqant 2x 依赖组件

    1. nats
    2. consul 或 etcdv3等服务发现组件

# 服务发现
一开始本人是想直接使用go-micro重新开发一套游戏框架的,但在测试过程中发现go-micro的rpc性能有问题(也许是本人压测方案不对,如有问题请指正)
因此才再次转战会mqant,并简单粗暴的将go-micro优秀的服务发现设计理论和接口直接移植过来了。

# 复用go-micro服务发现插件

因为参考了go-micro的接口设计,mqant几乎可以毫无成本的复用go-micro开发的所有服务发现相关插件

1. registry

        consul  已内置
        etcd
        etcdv3  已内置
        eureka
        kubernetes
        nats
        proxy
        zookeeper

2. selector

        cache   已内置
        blacklist
        label
        shard
        static

# 完整保持了mqant已有接口的兼容性

虽然mqant在1x版本中没有服务发现的相关功能,但整体的设计还是跟go-micro差不太多,因此在移植服务发现时
基本保持了对已有接口的兼容。并且还针对服务发现进行了扩展,以满足各种服务治理的需求

# 只保留了nats作为rpc通道

mqant 1x版本提供了rabbitmq,redis,udp通道,go-micro更是提供了grpc等rpc通道，但mqant在
2x版本中只保留了nats作为唯一的rpc通道,这样做一方面维护成本的考虑,另外一方面是想保持mqant设计理念
的纯净,太多的rpc通道各有差异，导致最终的rpc效果并不理想（我想go-micro性能问题可能正在于此），不过
mqant保留了扩展接口，以后有机会新增通道。

