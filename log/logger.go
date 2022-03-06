package log

import (
	beegolog "github.com/liangdas/mqant/log/beego"
	mqanttools "github.com/liangdas/mqant/utils"
)

// Fields type, used to pass to `WithFields`.
type Fields map[string]interface{}

type Logger interface {
	TPLogger
	Info(format string, keyvals ...interface{})
	Debug(format string, keyvals ...interface{})
	Warning(format string, keyvals ...interface{})
	Error(format string, keyvals ...interface{})
}

type TPLogger interface {
	TInfo(span TraceSpan, format string, keyvals ...interface{})
	TDebug(span TraceSpan, format string, keyvals ...interface{})
	TWarning(span TraceSpan, format string, keyvals ...interface{})
	TError(span TraceSpan, format string, keyvals ...interface{})
}

func New(logger *beegolog.BeeLogger) Logger {
	return &Beego{
		BeeLogger: logger,
	}
}

type Beego struct {
	*beegolog.BeeLogger
}

// CreateRootTrace CreateRootTrace
func rootTrace() TraceSpan {
	return &TraceSpanImp{
		Trace: mqanttools.GenerateID().String(),
		Span:  mqanttools.GenerateID().String(),
	}
}

// Info 输出普通日志
func (b *Beego) Info(format string, keyvals ...interface{}) {
	ts := rootTrace()
	tc := &beegolog.BeegoTraceSpan{
		Trace: ts.TraceId(),
		Span:  ts.SpanId(),
	}
	if b.BeeLogger != nil {
		b.BeeLogger.Info(tc, format, keyvals)
	}
}

// Debug 输出调试日志
func (b *Beego) Debug(format string, keyvals ...interface{}) {
	if b.BeeLogger != nil {
		b.BeeLogger.Debug(nil, format, keyvals)
	}
}

// Warning 输出警告日志
func (b *Beego) Warning(format string, keyvals ...interface{}) {
	ts := rootTrace()
	tc := &beegolog.BeegoTraceSpan{
		Trace: ts.TraceId(),
		Span:  ts.SpanId(),
	}
	if b.BeeLogger != nil {
		b.BeeLogger.Warning(tc, format, keyvals)
	}
}

// Error 输出错误日志
func (b *Beego) Error(format string, keyvals ...interface{}) {
	ts := rootTrace()
	tc := &beegolog.BeegoTraceSpan{
		Trace: ts.TraceId(),
		Span:  ts.SpanId(),
	}
	if b.BeeLogger != nil {
		b.BeeLogger.Error(tc, format, keyvals)
	}
}

// TInfo 输出普通日志
func (b *Beego) TInfo(span TraceSpan, format string, keyvals ...interface{}) {
	ts := rootTrace()
	tc := &beegolog.BeegoTraceSpan{
		Trace: ts.TraceId(),
		Span:  ts.SpanId(),
	}
	if b.BeeLogger != nil {
		b.BeeLogger.Info(tc, format, keyvals)
	}
}

// TError 输出错误日志
func (b *Beego) TError(span TraceSpan, format string, keyvals ...interface{}) {
	ts := rootTrace()
	tc := &beegolog.BeegoTraceSpan{
		Trace: ts.TraceId(),
		Span:  ts.SpanId(),
	}
	if b.BeeLogger != nil {
		b.BeeLogger.Error(tc, format, keyvals)
	}
}

// TDebug 输出调试日志
func (b *Beego) TDebug(span TraceSpan, format string, keyvals ...interface{}) {
	ts := rootTrace()
	tc := &beegolog.BeegoTraceSpan{
		Trace: ts.TraceId(),
		Span:  ts.SpanId(),
	}
	if b.BeeLogger != nil {
		b.BeeLogger.Debug(tc, format, keyvals)
	}
}

// TWarning 输出警告日志
func (b *Beego) TWarning(span TraceSpan, format string, keyvals ...interface{}) {
	ts := rootTrace()
	tc := &beegolog.BeegoTraceSpan{
		Trace: ts.TraceId(),
		Span:  ts.SpanId(),
	}
	if b.BeeLogger != nil {
		b.BeeLogger.Warning(tc, format, keyvals)
	}
}
