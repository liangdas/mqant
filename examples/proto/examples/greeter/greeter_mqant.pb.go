package greeter

// This is a compile-time assertion to ensure that this generated file
// is compatible with the kratos package it is being compiled against.
import (
	"errors"
	client "github.com/liangdas/mqant/module"
	basemodule "github.com/liangdas/mqant/module/base"
	mqrpc "github.com/liangdas/mqant/rpc"
	"golang.org/x/net/context"
)

// generated mqant method
type Greeter interface {
	Hello(in *Request) (out *Response, err error)
	Stream(in *Request) (out *Response, err error)
}

func RegisterGreeterTcpHandler(m *basemodule.BaseModule, ser Greeter) {
	m.GetServer().RegisterGO("hello", ser.Hello)
	m.GetServer().RegisterGO("stream", ser.Stream)
}

// generated proxxy handle
type ClientProxyService struct {
	cli  client.App
	name string
}

var ClientProxyIsNil = errors.New("proxy is nil")

func NewGreeterClient(cli client.App, name string) *ClientProxyService {
	return &ClientProxyService{
		cli:  cli,
		name: name,
	}
}
func (proxy *ClientProxyService) Hello(req *Request) (rsp *Response, err error) {
	if proxy == nil {
		return nil, ClientProxyIsNil
	}
	rsp = &Response{}
	err = mqrpc.Proto(rsp, func() (reply interface{}, err interface{}) {
		return proxy.cli.Call(context.TODO(), proxy.name, "hello", mqrpc.Param(req))
	})
	return rsp, err
}
func (proxy *ClientProxyService) Stream(req *Request) (rsp *Response, err error) {
	if proxy == nil {
		return nil, ClientProxyIsNil
	}
	rsp = &Response{}
	err = mqrpc.Proto(rsp, func() (reply interface{}, err interface{}) {
		return proxy.cli.Call(context.TODO(), proxy.name, "stream", mqrpc.Param(req))
	})
	return rsp, err
}
