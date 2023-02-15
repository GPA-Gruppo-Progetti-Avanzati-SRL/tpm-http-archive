package logzerotracer

import (
	"encoding/json"
	"fmt"
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-har/hartracing"
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-har/hartracing/util"
	"github.com/rs/zerolog/log"
	"io"
	"strings"
	"time"
)

type tracerImpl struct {
}

func NewTracer() (hartracing.Tracer, io.Closer) {
	t := &tracerImpl{}
	return t, t
}

func (t *tracerImpl) Close() error {
	return nil
}

func (t *tracerImpl) StartSpan(opts ...hartracing.SpanOption) hartracing.Span {
	const semLogContext = "log-zero-har-tracer::start-har-span"

	spanOpts := hartracing.SpanOptions{}
	for _, o := range opts {
		o(&spanOpts)
	}

	oid := util.NewObjectId().String()
	spanCtx := hartracing.SimpleSpanContext{LogId: oid, ParentId: oid, TraceId: oid}

	if spanOpts.ParentContext != nil {
		if ctxImpl, ok := spanOpts.ParentContext.(hartracing.SimpleSpanContext); ok {
			spanCtx.LogId = ctxImpl.LogId
			spanCtx.ParentId = ctxImpl.TraceId
		} else {
			log.Warn().Msg(semLogContext + " unsupported implementation: wanted internal.spanContextImpl")
		}
	}

	span := spanImpl{
		hartracing.SimpleSpan{
			Tracer:      t,
			SpanContext: spanCtx,
			StartTime:   time.Now(),
		},
	}

	return &span
}

func (t *tracerImpl) Report(s *spanImpl) error {
	const semLogContext = "log-zero-har-tracer::report"

	h, err := s.GetHARData()
	if err != nil {
		log.Error().Err(err).Msg(semLogContext)
		return err
	}

	b, err := json.Marshal(h)
	if err != nil {
		log.Error().Err(err).Msg(semLogContext)
		return err
	}

	fmt.Println(string(b))
	log.Trace().Str("span-id", s.Id()).Str("har", string(b)).Msg(semLogContext)
	return nil
}

func (t *tracerImpl) Extract(format string, tmr hartracing.TextMapReader) (hartracing.SpanContext, error) {

	var spanContext hartracing.SimpleSpanContext
	err := tmr.ForeachKey(func(key, val string) error {
		var err error
		if strings.ToLower(key) == hartracing.HARTraceIdHeaderName {
			spanContext, err = hartracing.ExtractSimpleSpanContextFromString(val)
			return err
		}

		return nil
	})

	if spanContext.IsZero() {
		err = hartracing.ErrSpanContextNotFound
	}

	return spanContext, err
}

func (t *tracerImpl) Inject(s hartracing.SpanContext, tmr hartracing.TextMapWriter) error {
	tmr.Set(hartracing.HARTraceIdHeaderName, s.Id())
	return nil
}
