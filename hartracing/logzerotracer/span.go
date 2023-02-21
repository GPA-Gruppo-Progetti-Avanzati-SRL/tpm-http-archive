package logzerotracer

import (
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-http-archive/hartracing"
	"github.com/rs/zerolog/log"
	"time"
)

type logZeroSpanImpl struct {
	hartracing.SimpleSpan
}

func (hs *logZeroSpanImpl) Finish() error {

	const semLogContext = "log-zero-har-tracer::finish-span"

	hs.Duration = time.Since(hs.StartTime)
	if len(hs.Entries) > 0 {
		log.Trace().Str("span-id", hs.Id()).Msg(semLogContext + " reporting span")
		_ = hs.Tracer.(*logZeroTracerImpl).Report(hs)
	} else {
		log.Warn().Str("span-id", hs.Id()).Msg(semLogContext + " no Entries in span....")
	}

	return nil
}
