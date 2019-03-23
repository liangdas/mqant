// Package server is an interface for a micro server
package server

import (
	"context"
	"github.com/pborman/uuid"
	"github.com/liangdas/mqant/module"
	"github.com/liangdas/mqant/conf"
	"github.com/liangdas/mqant/rpc"
)

type Server interface {
	Options() Options
	OnInit(module module.Module, app module.App, settings *conf.ModuleSettings) error
	Init(...Option) error
	SetListener(listener mqrpc.RPCListener)
	Register(id string, f interface{})
	RegisterGO(id string, f interface{})
	ServiceRegister() error
	ServiceDeregister() error
	Start() error
	Stop() error
	OnDestroy() error
	String() string
	Id() string
}

type Message interface {
	Topic() string
	Payload() interface{}
	ContentType() string
}

type Request interface {
	Service() string
	Method() string
	ContentType() string
	Request() interface{}
	// indicates whether the request will be streamed
	Stream() bool
}

// Stream represents a stream established with a client.
// A stream can be bidirectional which is indicated by the request.
// The last error will be left in Error().
// EOF indicated end of the stream.
type Stream interface {
	Context() context.Context
	Request() Request
	Send(interface{}) error
	Recv(interface{}) error
	Error() error
	Close() error
}

type Option func(*Options)


var (
	DefaultAddress        = ":0"
	DefaultName           = "go-server"
	DefaultVersion        = "1.0.0"
	DefaultId             = uuid.NewUUID().String()
)


// NewServer returns a new server with options passed in
func NewServer(opt ...Option) Server {
	return newRpcServer(opt...)
}