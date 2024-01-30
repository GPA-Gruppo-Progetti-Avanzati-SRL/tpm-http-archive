package factory

import (
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-http-archive/hartracing"
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-http-archive/hartracing/filetracer"
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-http-archive/hartracing/logzerotracer"
	"github.com/rs/zerolog/log"
	"io"
	"os"
	"strings"
)

func HarTracerTypeFromEnv() string {
	const semLogContext = "har-tracing::type-from-env"
	trcType := os.Getenv(hartracing.HARTracerTypeEnvName)
	if trcType == "" {
		log.Info().Msgf(semLogContext+" env var %s not set", hartracing.HARTracerTypeEnvName)
	}

	return strings.ToLower(trcType)
}

func IsHarTracerTypeFromEnvSupported() bool {
	const semLogContext = "har-tracing::is-type-from-env-supported"
	trcType := HarTracerTypeFromEnv()
	if trcType == "" {
		log.Info().Msgf(semLogContext+" env var %s not set", hartracing.HARTracerTypeEnvName)
	}

	return trcType == filetracer.HarFileTracerType || trcType == logzerotracer.HarLogZeroTracerType
}

func InitHarTracingFromEnv() (io.Closer, error) {

	const semLogContext = "har-tracing::init-from-env"
	const semLogLabelTracerType = "tracer-type"

	var trc hartracing.Tracer
	var closer io.Closer
	var err error

	trcType := HarTracerTypeFromEnv()
	if trcType == "" {
		return closer, nil
	}

	log.Info().Str(semLogLabelTracerType, trcType).Msg(semLogContext)
	switch strings.ToLower(trcType) {
	case filetracer.HarFileTracerType:
		trc, closer, err = filetracer.NewTracer()
		if err != nil {
			return nil, err
		}

	case logzerotracer.HarLogZeroTracerType:
		trc, closer, err = logzerotracer.NewTracer()
		if err != nil {
			return nil, err
		}

	default:
		log.Info().Str(semLogLabelTracerType, trcType).Msg(semLogContext + " unrecognized tracer type")
	}

	if trc != nil {
		hartracing.SetGlobalTracer(trc)
	}

	return closer, nil
}
