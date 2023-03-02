package hartracing

import (
	"golang.org/x/net/context"
)

const (
	HARTraceIdHeaderName       = "har-trace-id"
	HARTracerTypeEnvName       = "HAR_TRACER_TYPE"
	HARTraceOpenTracingTagName = "har.trace.id"
	HARSpanFlagSampled         = "1"
	HARSpanFlagUnSampled       = "0"
)

type Tracer interface {
	StartSpan(opts ...SpanOption) Span
	Extract(format string, tmr TextMapReader) (SpanContext, error)
	Inject(s SpanContext, tmr TextMapWriter) error
	IsNil() bool
}

var globalTracer Tracer

func SetGlobalTracer(t Tracer) {
	globalTracer = t
}

func GlobalTracer() Tracer {
	if globalTracer == nil {
		globalTracer, _ = NilTracer()
	}
	return globalTracer
}

func ContextWithSpan(ctx context.Context, span Span) context.Context {
	return context.WithValue(ctx, HARTraceIdHeaderName, span)
}

func SpanFromContext(ctx context.Context) Span {
	val := ctx.Value(HARTraceIdHeaderName)
	if sp, ok := val.(Span); ok {
		return sp
	}
	return nil
}
