package dafaultrouter

import (
	"context"
	"encoding/json"
	"github.com/liangdas/mqant/gate"
	"github.com/liangdas/mqant/module"
	argsutil "github.com/liangdas/mqant/rpc/util"
	"github.com/pkg/errors"
	"log"
	"net/url"
	"strings"
)

type Option func(*DefaultRouter)

type DefaultRouter struct {
	module      module.RPCModule
}

func NewDefaultRouter(module module.RPCModule, opts ...Option) *DefaultRouter {
	route := &DefaultRouter{
		module:      module,
	}
	if opts != nil {
		for _, o := range opts {
			o(route)
		}
	}
	return route
}

//OnRoute 默认路由方法
func (this *DefaultRouter) OnRoute(session gate.Session, topic string, msg []byte) (bool, interface{}, error) {
	_, err := url.Parse(topic)
	topics := strings.Split(topic, "/")
	//case1: topics 少于2层结构
	if len(topics) < 2 {
		return true, nil, errors.Errorf("Topic must be [moduleType@moduleID]/[handler]|[moduleType@moduleID]/[handler]/[msgid]")
	}
	var msgid string
	if len(topics) == 3 {
		msgid = topics[2]
	}
	//case2: topics 第二2层不是以HD_开头
	if startsWith := strings.HasPrefix(topics[1], "HD_"); !startsWith {
		return true, nil, errors.Errorf("Method(%s) must begin with 'HD_'", topics[1])
	}
	//case3: 服务器找不到
	var ArgsType []string = make([]string, 2)
	var args [][]byte = make([][]byte, 2)
	serverSession, err := this.module.GetRouteServer(topics[0])
	if err != nil && msgid!="" {
		return true, nil, errors.Errorf("Service(type:%s) not found", topics[0])
	}
	//case4: 判断json格式是否正确
	if len(msg) > 0 && msg[0] == '{' && msg[len(msg)-1] == '}' {
		//尝试解析为json为map
		var obj interface{} // var obj map[string]interface{}
		err := json.Unmarshal(msg, &obj)
		if err != nil && msgid!="" {
			return true, nil, errors.Errorf("The JSON format is incorrect")
		}
		ArgsType[1] = argsutil.MAP
		args[1] = msg
	} else {
		ArgsType[1] = argsutil.BYTES
		args[1] = msg
	}
	session = session.Clone()
	session.SetTopic(topic)

	//case5: RPC发送消息，同时回复
	if msgid != "" {
		ArgsType[0] = gate.RPCParamSessionType
		if b, err := session.Serializable(); err == nil {
			args[0] = b
		}
		ctx, _ := context.WithTimeout(context.TODO(), this.module.GetApp().Options().RPCExpired)
		result, e := serverSession.CallArgs(ctx, topics[1], ArgsType, args)
		return true, result, errors.Errorf(e)
	}

	//case6: RPC发送消息，不回复
	ArgsType[0] = gate.RPCParamSessionType
	if b, err := session.Serializable(); err == nil {
		args[0] = b
	}
	e := serverSession.CallNRArgs(topics[1], ArgsType, args)
	if e != nil {
		log.Println("Gate rpc", e.Error())
	}
	return false, nil, nil
}
