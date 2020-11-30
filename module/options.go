package module

import (
	"github.com/liangdas/mqant/registry"
	"github.com/liangdas/mqant/rpc"
	"github.com/liangdas/mqant/rpc/pb"
	"github.com/liangdas/mqant/selector"
	"github.com/nats-io/nats.go"
	"time"
)

// Option 配置项
type Option func(*Options)

// Options 应用级别配置项
type Options struct {
	Nats        *nats.Conn
	Version     string
	Debug       bool
	Parse       bool //是否由框架解析启动环境变量,默认为true
	WorkDir     string
	ConfPath    string
	LogDir      string
	BIDir       string
	ProcessID   string
	KillWaitTTL time.Duration
	Registry    registry.Registry
	Selector    selector.Selector
	// Register loop interval
	RegisterInterval   time.Duration
	RegisterTTL        time.Duration
	ClientRPChandler   ClientRPCHandler
	ServerRPCHandler   ServerRPCHandler
	RpcCompleteHandler RpcCompleteHandler
	RPCExpired         time.Duration
	RPCMaxCoroutine    int
}

// ClientRPCHandler 调用方RPC监控
type ClientRPCHandler func(app App, server registry.Node, rpcinfo *rpcpb.RPCInfo, result interface{}, err string, exec_time int64)

// ServerRPCHandler 服务方RPC监控
type ServerRPCHandler func(app App, module Module, callInfo *mqrpc.CallInfo)

// ServerRPCHandler 服务方RPC监控
type RpcCompleteHandler func(app App, module Module, callInfo *mqrpc.CallInfo, input []interface{}, out []interface{}, execTime time.Duration)

// Version 应用版本
func Version(v string) Option {
	return func(o *Options) {
		o.Version = v
	}
}

// Debug 只有是在调试模式下才会在控制台打印日志, 非调试模式下只在日志文件中输出日志
func Debug(t bool) Option {
	return func(o *Options) {
		o.Debug = t
	}
}

// WorkDir 进程工作目录
func WorkDir(v string) Option {
	return func(o *Options) {
		o.WorkDir = v
	}
}

// Configure 配置路径
func Configure(v string) Option {
	return func(o *Options) {
		o.ConfPath = v
	}
}

// LogDir 日志存储路径
func LogDir(v string) Option {
	return func(o *Options) {
		o.LogDir = v
	}
}

// ProcessID 进程分组ID
func ProcessID(v string) Option {
	return func(o *Options) {
		o.ProcessID = v
	}
}

// BILogDir  BI日志路径
func BILogDir(v string) Option {
	return func(o *Options) {
		o.BIDir = v
	}
}

// Nats  nats配置
func Nats(nc *nats.Conn) Option {
	return func(o *Options) {
		o.Nats = nc
	}
}

// Registry sets the registry for the service
// and the underlying components
func Registry(r registry.Registry) Option {
	return func(o *Options) {
		o.Registry = r
		o.Selector.Init(selector.Registry(r))
	}
}

// Selector 路由选择器
func Selector(r selector.Selector) Option {
	return func(o *Options) {
		o.Selector = r
	}
}

// RegisterTTL specifies the TTL to use when registering the service
func RegisterTTL(t time.Duration) Option {
	return func(o *Options) {
		o.RegisterTTL = t
	}
}

// RegisterInterval specifies the interval on which to re-register
func RegisterInterval(t time.Duration) Option {
	return func(o *Options) {
		o.RegisterInterval = t
	}
}

// KillWaitTTL specifies the interval on which to re-register
func KillWaitTTL(t time.Duration) Option {
	return func(o *Options) {
		o.KillWaitTTL = t
	}
}

// SetClientRPChandler 配置调用者监控器
func SetClientRPChandler(t ClientRPCHandler) Option {
	return func(o *Options) {
		o.ClientRPChandler = t
	}
}

// SetServerRPCHandler 配置服务方监控器
func SetServerRPCHandler(t ServerRPCHandler) Option {
	return func(o *Options) {
		o.ServerRPCHandler = t
	}
}

// SetServerRPCCompleteHandler 服务RPC执行结果监控器
func SetRpcCompleteHandler(t RpcCompleteHandler) Option {
	return func(o *Options) {
		o.RpcCompleteHandler = t
	}
}

// Parse mqant框架是否解析环境参数
func Parse(t bool) Option {
	return func(o *Options) {
		o.Parse = t
	}
}

//RPC超时时间
func RPCExpired(t time.Duration) Option {
	return func(o *Options) {
		o.RPCExpired = t
	}
}

//单个节点RPC同时并发协程数
func RPCMaxCoroutine(t int) Option {
	return func(o *Options) {
		o.RPCMaxCoroutine = t
	}
}
