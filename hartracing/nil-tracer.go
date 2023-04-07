package hartracing

import (
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-http-archive/hartracing/util"
	"github.com/rs/zerolog/log"
	"io"
	"strings"
	"time"
)

type nilSpanImpl struct {
	SimpleSpan
}

func (hs *nilSpanImpl) Finish() error {
	const semLogContext = "nil-har-tracer::finish-span"
	return nil
}

type nilTracerImpl struct {
}

func NilTracer() (Tracer, io.Closer) {
	const semLogContext = "nil-har-tracer::new"
	log.Info().Msg(semLogContext)
	t := &nilTracerImpl{}
	return t, t
}

func (t *nilTracerImpl) Close() error {
	return nil
}

func (t *nilTracerImpl) IsNil() bool {
	return true
}

func (t *nilTracerImpl) StartSpan(opts ...SpanOption) Span {
	const semLogContext = "nil-har-tracer::start-har-span"

	spanOpts := SpanOptions{}
	for _, o := range opts {
		o(&spanOpts)
	}

	oid := util.NewTraceId()
	spanCtx := SimpleSpanContext{LogId: oid, ParentId: oid, TraceId: oid, Flag: HARSpanFlagUnSampled}

	if spanOpts.ParentContext != nil {
		if ctxImpl, ok := spanOpts.ParentContext.(SimpleSpanContext); ok {
			spanCtx.LogId = ctxImpl.LogId
			spanCtx.ParentId = ctxImpl.TraceId
		} else {
			log.Warn().Msg(semLogContext + " unsupported implementation: wanted internal.spanContextImpl")
		}
	}

	span := nilSpanImpl{
		SimpleSpan{
			Tracer:      t,
			SpanContext: spanCtx,
			StartTime:   time.Now(),
		},
	}

	return &span
}

func (t *nilTracerImpl) Report(s *nilSpanImpl) error {
	const semLogContext = "nil-har-tracer::report"
	return nil
}

func (t *nilTracerImpl) Extract(format string, tmr TextMapReader) (SpanContext, error) {

	var spanContext SimpleSpanContext
	err := tmr.ForeachKey(func(key, val string) error {
		var err error
		if strings.ToLower(key) == HARTraceIdHeaderName {
			spanContext, err = ExtractSimpleSpanContextFromString(val)
			return err
		}

		return nil
	})

	if spanContext.IsZero() {
		err = ErrSpanContextNotFound
	}

	return spanContext, err
}

func (t *nilTracerImpl) Inject(s SpanContext, tmr TextMapWriter) error {
	tmr.Set(HARTraceIdHeaderName, s.Id())
	return nil
}
