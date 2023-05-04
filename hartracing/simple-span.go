package hartracing

import (
	"errors"
	"fmt"
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-http-archive/har"
	"github.com/rs/zerolog/log"
	"os"
	"strings"
	"time"
)

type SimpleSpanContext struct {
	LogId    string
	ParentId string
	TraceId  string
	Flag     string
}

func (spanCtx SimpleSpanContext) Id() string {
	return spanCtx.Encode()
}

func (spanCtx SimpleSpanContext) Encode() string {
	s := fmt.Sprintf("%s:%s:%s:%s", spanCtx.LogId, spanCtx.ParentId, spanCtx.TraceId, spanCtx.Flag)
	return s
}

func (spanCtx SimpleSpanContext) IsZero() bool {
	return spanCtx.LogId == "" && spanCtx.ParentId == "" && spanCtx.TraceId == "" && spanCtx.Flag == ""
}

func (spanCtx SimpleSpanContext) Sampled() bool {
	return spanCtx.Flag == "1"
}

func ExtractSimpleSpanContextFromString(ser string) (SimpleSpanContext, error) {
	sarr := strings.Split(ser, ":")
	if len(sarr) != 4 {
		return SimpleSpanContext{}, fmt.Errorf("invalid span %s", ser)
	}

	sctx := SimpleSpanContext{
		LogId:    sarr[0],
		ParentId: sarr[1],
		TraceId:  sarr[2],
		Flag:     sarr[3],
	}

	return sctx, nil
}

type SimpleSpan struct {
	Tracer      Tracer
	SpanContext SimpleSpanContext
	Creator     har.Creator
	Browser     har.Creator
	Comment     string
	StartTime   time.Time
	Duration    time.Duration
	Finished    bool
	Entries     []*har.Entry
}

func (hs *SimpleSpan) Finish() error {
	panic(errors.New("apparently the Finish method on har-tracing::SimpleSpan has been invoked.... check the implementation"))
	return nil
}

func (hs *SimpleSpan) Id() string {
	return hs.SpanContext.Encode()
}

func (hs *SimpleSpan) Context() SpanContext {
	return hs.SpanContext
}

func (hs *SimpleSpan) Sampled() bool {
	return hs.SpanContext.Sampled()
}

func (hs *SimpleSpan) String() string {
	id := hs.SpanContext.Encode()
	return fmt.Sprintf("[%s] #Entries: %d - start: %s - dur: %d", id, len(hs.Entries), hs.StartTime.Format(time.RFC3339Nano), hs.Duration.Milliseconds())
}

func (hs *SimpleSpan) AddEntry(e *har.Entry) error {
	e.TraceId = hs.Id()
	hs.Entries = append(hs.Entries, e)
	return nil
}

func (hs *SimpleSpan) GetHARData() (*har.HAR, error) {

	const semLogContext = "log-tracer-span::get-har-data"
	podName := os.Getenv("HOSTNAME")
	if podName == "" {
		log.Info().Msg(semLogContext + " HOSTNAME env variable not set.... using localhost")
		podName = "localhost"
	}

	har := har.HAR{
		Log: &har.Log{
			Version: "1.1",
			Creator: &har.Creator{
				Name:    "tpm-har",
				Version: "1.0",
			},
			Browser: &hs.Browser,
			Comment: hs.Comment,
			TraceId: hs.Id(),
		},
	}

	for _, e := range hs.Entries {
		har.Log.Entries = append(har.Log.Entries, e)
	}

	return &har, nil
}
