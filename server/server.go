// Package server is an interface for a micro server
package server

import (
	"context"
	"github.com/liangdas/mqant/conf"
	"github.com/liangdas/mqant/module"
	"github.com/liangdas/mqant/rpc"
	"github.com/pborman/uuid"
)

// Server Server
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
	ID() string
	// Deprecated: 因为命名规范问题函数将废弃,请用ID代替
	Id() string
}

// Message RPC消息头
type Message interface {
	Topic() string
	Payload() interface{}
	ContentType() string
}

// Request Request
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

// Option Option
type Option func(*Options)

var (
	// DefaultAddress DefaultAddress
	DefaultAddress = ":0"
	// DefaultName DefaultName
	DefaultName = "go-server"
	// DefaultVersion DefaultVersion
	DefaultVersion = "1.0.0"
	// DefaultID DefaultID
	DefaultID = uuid.NewUUID().String()
)

// NewServer returns a new server with options passed in
func NewServer(opt ...Option) Server {
	return newRPCServer(opt...)
}
