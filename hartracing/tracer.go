package hartracing

import (
	"golang.org/x/net/context"
)

const (
	HARTraceIdHeaderName = "har-trace-id"
)

type Tracer interface {
	StartSpan(opts ...SpanOption) Span
	Extract(format string, tmr TextMapReader) (SpanContext, error)
	Inject(s SpanContext, tmr TextMapWriter) error
}

var globalTracer Tracer

func SetGlobalTracer(t Tracer) {
	globalTracer = t
}

func GlobalTracer() Tracer {
	return globalTracer
}

func ContextWithSpan(ctx context.Context, span Span) context.Context {
	return context.WithValue(ctx, HARTraceIdHeaderName, span)
}
