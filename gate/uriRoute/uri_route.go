package uriRoute

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/liangdas/mqant/gate"
	"github.com/liangdas/mqant/log"
	"github.com/liangdas/mqant/module"
	"github.com/liangdas/mqant/rpc/util"
	"github.com/pkg/errors"
	"net/url"
	"time"
)

/**
服务节点选择函数，可以自定义服务筛选规则
如不指定,默认使用 Scheme作为moduleType,Hostname作为服务节点nodeId
如随机到服务节点Hostname可以用modulus,cache,random等通用规则
例如:
 im://modulus/remove_feeds_member?msg_id=1002
*/
type FSelector func(topic string, u *url.URL) (s module.ServerSession, err error)

/**
指定数据解析函数
返回值如bean！=nil err==nil则会向后端模块传入 func(session,bean)(result, error)
否则使用json或[]byte传参
*/
type FDataParsing func(topic string, u *url.URL, msg []byte) (bean interface{}, err error)

type Option func(*UriRoute)

func NewUriRoute(module module.RPCModule, opts ...Option) *UriRoute {
	route := &UriRoute{
		module:      module,
		CallTimeOut: 3 * time.Second,
	}
	for _, o := range opts {
		o(route)
	}
	return route
}

func Selector(t FSelector) Option {
	return func(o *UriRoute) {
		o.Selector = t
	}
}

func DataParsing(t FDataParsing) Option {
	return func(o *UriRoute) {
		o.DataParsing = t
	}
}

func CallTimeOut(t time.Duration) Option {
	return func(o *UriRoute) {
		o.CallTimeOut = t
	}
}

type UriRoute struct {
	module      module.RPCModule
	Selector    FSelector
	DataParsing FDataParsing
	CallTimeOut time.Duration
}

func (this *UriRoute) OnRoute(session gate.Session, topic string, msg []byte) (bool, interface{}, error) {
	needreturn := true
	u, err := url.Parse(topic)
	if err != nil {
		return needreturn, nil, errors.Errorf("topic is not uri %v", err.Error())
	}
	var ArgsType []string = nil
	var args [][]byte = nil

	_func := u.Path
	m, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return needreturn, nil, errors.Errorf("parse query error %v", err.Error())
	}
	if _, ok := m["msg_id"]; !ok {
		needreturn = false
	}
	ArgsType = make([]string, 2)
	args = make([][]byte, 2)
	session.SetTopic(topic)
	var serverSession module.ServerSession
	if this.Selector != nil {
		ss, err := this.Selector(topic, u)
		if err != nil {
			return needreturn, nil, err
		}
		serverSession = ss
	} else {
		moduleType := u.Scheme
		if u.Hostname() == "modulus" {
			//取模
		} else if u.Hostname() == "cache" {
			//缓存
		} else if u.Hostname() == "random" {
			//随机
		} else {
			//其他规则就是 module://[user:pass@]nodeId/path
			moduleType = fmt.Sprintf("%v@%v", moduleType, u.Hostname())
		}
		ss, err := this.module.GetRouteServer(moduleType)
		if err != nil {
			return needreturn, nil, errors.Errorf("Service(type:%s) not found", moduleType)
		}
		serverSession = ss
	}

	if this.DataParsing != nil {
		bean, err := this.DataParsing(topic, u, msg)
		if err == nil && bean != nil {
			if needreturn {
				ctx, _ := context.WithTimeout(context.TODO(), this.CallTimeOut)
				result, e := serverSession.Call(ctx, _func, session, bean)
				if e != "" {
					return needreturn, result, errors.New(e)
				} else {
					return needreturn, result, nil
				}
			} else {
				e := serverSession.CallNR(_func, session, bean)
				if e != nil {
					log.Warning("Gate RPC", e.Error())
					return needreturn, nil, e
				}
			}
			return needreturn, nil, nil
		}
	}

	//默认参数
	if msg[0] == '{' && msg[len(msg)-1] == '}' {
		//尝试解析为json为map
		var obj interface{} // var obj map[string]interface{}
		err := json.Unmarshal(msg, &obj)
		if err != nil {
			return needreturn, nil, errors.Errorf("The JSON format is incorrect %v", err)
		}
		ArgsType[1] = argsutil.MAP
		args[1] = msg
	} else {
		ArgsType[1] = argsutil.BYTES
		args[1] = msg
	}
	s := session.Clone()
	s.SetTopic(topic)
	if needreturn {
		ArgsType[0] = gate.RPC_PARAM_SESSION_TYPE
		b, err := s.Serializable()
		if err != nil {
			return needreturn, nil, err
		}
		args[0] = b
		ctx, _ := context.WithTimeout(context.TODO(), this.CallTimeOut)
		result, e := serverSession.CallArgs(ctx, _func, ArgsType, args)
		if e != "" {
			return needreturn, result, errors.New(e)
		} else {
			return needreturn, result, nil
		}
	} else {
		ArgsType[0] = gate.RPC_PARAM_SESSION_TYPE
		b, err := s.Serializable()
		if err != nil {
			return needreturn, nil, err
		}
		args[0] = b

		e := serverSession.CallNRArgs(_func, ArgsType, args)
		if e != nil {
			log.Warning("Gate RPC", e.Error())
			return needreturn, nil, e
		}
	}
	return needreturn, nil, nil
}
