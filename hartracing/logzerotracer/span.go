package logzerotracer

import (
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-har/hartracing"
	"github.com/rs/zerolog/log"
	"time"
)

type spanImpl struct {
	hartracing.SimpleSpan
}

func (hs *spanImpl) Finish() error {

	const semLogContext = "log-zero-har-tracer::span::finish"

	hs.Duration = time.Since(hs.StartTime)
	if len(hs.Entries) > 0 {
		log.Trace().Str("span-id", hs.Id()).Msg(semLogContext + " reporting span")
		_ = hs.Tracer.(*tracerImpl).Report(hs)
	} else {
		log.Warn().Str("span-id", hs.Id()).Msg(semLogContext + " no Entries in span....")
	}

	return nil
}
