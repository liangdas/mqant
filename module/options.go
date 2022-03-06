package module

import (
	"github.com/liangdas/mqant/log"
	"time"

	"github.com/liangdas/mqant/registry"
	mqrpc "github.com/liangdas/mqant/rpc"
	rpcpb "github.com/liangdas/mqant/rpc/pb"
	"github.com/liangdas/mqant/selector"
	"github.com/nats-io/nats.go"
)

// Option 配置项
type Option func(*Options)

// Options 应用级别配置项
type Options struct {
	Nats    *nats.Conn
	Version string
	Debug   bool
	Parse   bool //是否由框架解析启动环境变量,默认为true
	// Deprecated: 新版本废弃启动配置。无需指定工作目录
	WorkDir string
	// Deprecated: 新版本废弃启动配置。无需指定配置路径
	ConfPath string
	// Deprecated: 新版本废弃启动配置。无需指定日志路径
	LogDir string
	// Deprecated: 新版本废弃启动配置。无需指定bi路径
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
	// 自定义日志文件名字
	// 主要作用方便k8s映射日志不会被冲突，建议使用k8s pod实现
	LogFileName FileNameHandler
	// 自定义BI日志名字
	BIFileName FileNameHandler
	// 自定义日志组件 兼容1.13版本的重定义日志组件， 只能存在一种
	Logger Logger
}

type Logger func() log.Logger

type FileNameHandler func(logdir, prefix, processID, suffix string) string

// ClientRPCHandler 调用方RPC监控
type ClientRPCHandler func(app App, server registry.Node, rpcinfo *rpcpb.RPCInfo, result interface{}, err string, exec_time int64)

// ServerRPCHandler 服务方RPC监控
type ServerRPCHandler func(app App, module Module, callInfo *mqrpc.CallInfo)

// RpcCompleteHandler 服务方RPC监控
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
// Deprecated: 新版本废弃启动配置。无需指定工作目录
func WorkDir(v string) Option {
	return func(o *Options) {
		o.WorkDir = v
	}
}

// Configure 配置路径
// Deprecated: 新版本废弃启动配置。无需指定配置路径
func Configure(v string) Option {
	return func(o *Options) {
		o.ConfPath = v
	}
}

// LogDir 日志存储路径
// Deprecated: 新版本废弃启动配置。无需指定log工作目录，如需指定请调用WithFile方法指定
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
// Deprecated: 新版本废弃启动配置。无需指定bi工作目录，如需指定请调用WithLogFile方法指定
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

// SetRpcCompleteHandler 服务RPC执行结果监控器
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

// RPCExpired RPC超时时间
func RPCExpired(t time.Duration) Option {
	return func(o *Options) {
		o.RPCExpired = t
	}
}

// RPCMaxCoroutine 单个节点RPC同时并发协程数
func RPCMaxCoroutine(t int) Option {
	return func(o *Options) {
		o.RPCMaxCoroutine = t
	}
}

// WithLogFile 日志文件名称
func WithLogFile(name FileNameHandler) Option {
	return func(o *Options) {
		o.LogFileName = name
	}
}

// WithBIFile Bi日志名称
func WithBIFile(name FileNameHandler) Option {
	return func(o *Options) {
		o.BIFileName = name
	}
}

// WithLogger 自定义日志
func WithLogger(logger Logger) Option {
	return func(o *Options) {
		o.Logger = logger
	}
}
