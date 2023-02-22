package filetracer

import (
	"encoding/json"
	"fmt"
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-http-archive/har"
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-http-archive/hartracing"
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-http-archive/hartracing/util"
	"github.com/rs/zerolog/log"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	TargetFolderEnvName = "HAR_FILE_TRACER_FOLDER"
	HarFileTracerType   = "har-file-tracer"
)

type tracerImpl struct {
	targetFolder string
	done         bool
	outCh        chan *har.HAR
}

type tracerOpts struct {
	folder string
}

type Option func(opts *tracerOpts)

func WithFolder(t string) Option {
	return func(opts *tracerOpts) {
		opts.folder = t
	}
}

func NewTracer(opts ...Option) (hartracing.Tracer, io.Closer, error) {

	const semLogContext = "file-har-tracer::new"

	trcOpts := tracerOpts{}
	for _, o := range opts {
		o(&trcOpts)
	}

	if trcOpts.folder == "" {
		trcOpts.folder = os.Getenv(TargetFolderEnvName)
	}

	if trcOpts.folder == "" {
		err := fmt.Errorf("to properly use the tracer need to set the env-var %s with desired target folder", TargetFolderEnvName)
		log.Error().Err(err).Str("env-var", TargetFolderEnvName).Msg(semLogContext)
		return nil, nil, err
	}

	if !util.FolderExists(trcOpts.folder) {
		err := fmt.Errorf("the target folder %s doesn't exist", trcOpts.folder)
		log.Error().Err(err).Str("folder", trcOpts.folder).Msg(semLogContext)
		return nil, nil, err
	}

	t := &tracerImpl{targetFolder: trcOpts.folder, outCh: make(chan *har.HAR, 10)}
	log.Info().Str("tracer-type", HarFileTracerType).Str("folder", trcOpts.folder).Msg(semLogContext + " har tracer initialized")

	go t.processLoop()
	return t, t, nil
}

func (t *tracerImpl) Close() error {

	const semLogContext = "file-har-tracer::close"

	close(t.outCh)
	for !t.done {
		time.Sleep(1 * time.Second)
	}

	log.Info().Msg(semLogContext + " closed")
	return nil
}

func (t *tracerImpl) StartSpan(opts ...hartracing.SpanOption) hartracing.Span {
	const semLogContext = "file-har-tracer::start-har-span"

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
	const semLogContext = "file-har-tracer::report"

	h, err := s.GetHARData()
	if err != nil {
		log.Error().Err(err).Msg(semLogContext)
		return err
	}

	t.outCh <- h
	return nil
}

func (t *tracerImpl) processLoop() error {
	const semLogContext = "file-har-tracer::process-loop"

	log.Info().Msg(semLogContext + " starting loop")

	for h := range t.outCh {
		b, err := json.Marshal(h)
		if err != nil {
			log.Error().Err(err).Msg(semLogContext)
			continue
		}

		fn := t.getFileName(h.Log.TraceId)
		if util.FileExists(fn) {
			h, err = t.merge(h, fn)
			if err != nil {
				log.Error().Err(err).Msg(semLogContext)
				continue
			}

			b, err = json.Marshal(h)
			if err != nil {
				log.Error().Err(err).Msg(semLogContext)
				continue
			}
		}

		fmt.Printf("writing %d bytes\n", len(b))
		err = os.WriteFile(fn, b, fs.ModePerm)
		if err != nil {
			log.Error().Err(err).Msg(semLogContext)
		}
	}

	log.Info().Msg(semLogContext + " ending loop")
	t.done = true
	return nil
}

func (t *tracerImpl) merge(incoming *har.HAR, fileName string) (*har.HAR, error) {

	const semLogContext = "file-har-tracer::merge"
	log.Trace().Str("log-id", incoming.Log.TraceId).Str("fn", fileName).Msg(semLogContext)

	b, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	var another har.HAR
	err = json.Unmarshal(b, &another)
	if err != nil {
		return nil, err
	}

	var mergeResult *har.HAR
	if incoming.Log.TraceId < another.Log.TraceId {
		log.Trace().Str("into-log-id", incoming.Log.TraceId).Str("from-log-id", another.Log.TraceId).Msg(semLogContext + " add file log to current log")
		mergeResult, err = incoming.Merge(&another, harEntryCompare)
	} else {
		log.Trace().Str("from-log-id", incoming.Log.TraceId).Str("into-log-id", another.Log.TraceId).Msg(semLogContext + " add current log to file log")
		mergeResult, err = another.Merge(incoming, harEntryCompare)
	}

	if err != nil {
		return nil, err
	}

	return mergeResult, nil
}

func harEntryCompare(e1, e2 *har.Entry) bool {
	return e1.TraceId < e2.TraceId
}

func (t *tracerImpl) getFileName(traceId string) string {
	const semLogContext = "file-har-tracer::get-filename"

	ctx, err := hartracing.ExtractSimpleSpanContextFromString(traceId)
	if err != nil {
		log.Error().Err(err).Msg(semLogContext)
		return "Undef.har"
	}

	p := filepath.Join(t.targetFolder, fmt.Sprintf("span-%s.har", ctx.LogId))
	return p
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
