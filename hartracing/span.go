package hartracing

import (
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-http-archive/har"
)

const (
	HARTraceIdHeaderName = "har-trace-id"
)

type Tracer interface {
	StartSpan(opts ...SpanOption) Span
	Extract(format string, tmr TextMapReader) (SpanContext, error)
	Inject(s SpanContext, tmr TextMapWriter) error
}

type SpanContext interface {
	Id() string
}

type Span interface {
	Id() string
	Context() SpanContext
	AddEntry(e *har.Entry) error
	Finish() error
}

type SpanOptions struct {
	ParentContext SpanContext
	Creator       har.Creator
	Browser       har.Creator
	Comment       string
}

type SpanOption func(opts *SpanOptions)

func ChildOf(parent SpanContext) SpanOption {
	return func(opts *SpanOptions) {
		opts.ParentContext = parent
	}
}

func WithCreator(creator har.Creator) SpanOption {
	return func(opts *SpanOptions) {
		opts.Creator = creator
	}
}

func WithBrowser(browser har.Creator) SpanOption {
	return func(opts *SpanOptions) {
		opts.Browser = browser
	}
}

func WithComment(comment string) SpanOption {
	return func(opts *SpanOptions) {
		opts.Comment = comment
	}
}
