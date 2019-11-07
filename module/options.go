package module

import (
	"github.com/liangdas/mqant/registry"
	"github.com/liangdas/mqant/selector"
	"github.com/nats-io/nats.go"
	"time"
)

type Option func(*Options)

type Options struct {
	Nats        *nats.Conn
	Version     string
	Debug       bool
	WorkDir     string
	ConfPath    string
	LogDir      string
	BIDir       string
	ProcessID   string
	KillWaitTTL time.Duration
	Registry    registry.Registry
	Selector    selector.Selector
	// Register loop interval
	RegisterInterval time.Duration
	RegisterTTL      time.Duration
}

func Version(v string) Option {
	return func(o *Options) {
		o.Version = v
	}
}

//只有是在调试模式下才会在控制台打印日志, 非调试模式下只在日志文件中输出日志
func Debug(t bool) Option {
	return func(o *Options) {
		o.Debug = t
	}
}

func WorkDir(v string) Option {
	return func(o *Options) {
		o.WorkDir = v
	}
}

func Configure(v string) Option {
	return func(o *Options) {
		o.ConfPath = v
	}
}

func LogDir(v string) Option {
	return func(o *Options) {
		o.LogDir = v
	}
}

func ProcessID(v string) Option {
	return func(o *Options) {
		o.ProcessID = v
	}
}

func BILogDir(v string) Option {
	return func(o *Options) {
		o.BIDir = v
	}
}

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

// RegisterInterval specifies the interval on which to re-register
func KillWaitTTL(t time.Duration) Option {
	return func(o *Options) {
		o.KillWaitTTL = t
	}
}
