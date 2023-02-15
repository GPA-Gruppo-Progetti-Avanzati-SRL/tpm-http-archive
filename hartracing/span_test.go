package hartracing_test

import (
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-har/har"
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-har/hartracing"
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-har/hartracing/filetracer"
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-har/hartracing/logzerotracer"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
	"net/http"
	"os"
	"testing"
	"time"
)

var harEntries = []*har.Entry{
	{
		StartedDateTime: "2023-02-12T20:07:02.147874+01:00",
		Time:            5,
		Request: &har.Request{
			Method:      "POST",
			URL:         "/examples/example-001/api/v1/orc-001",
			HTTPVersion: "1.1",
			Cookies:     []har.Cookie{},
			Headers: []har.NameValuePair{
				{
					Name:  "Requestid",
					Value: "a-resquest-id",
				},
				{
					Name:  "Content-Type",
					Value: "application/json",
				},
			},
			QueryString: har.NameValuePairs{},
			PostData: &har.PostData{
				MimeType: "application/json",
				Params:   har.Params{},
				Text:     "\"{\\n  \\\"canale\\\": \\\"APPP\\\",\\n  \\\"ordinante\\\": {\\n    \\\"natura\\\": \\\"PP\\\",\\n    \\\"tipologia\\\": \\\"ALIAS\\\",\\n    \\\"numero\\\": \\\"10724279\\\",\\n    \\\"codiceFiscale\\\": \\\"77626979028\\\",\\n    \\\"intestazione\\\": \\\"string\\\"\\n  }\\n}\"",
				Comment:  "",
				Data:     nil,
			},
			HeadersSize: -1,
			BodySize:    101,
			Comment:     "",
		},
		Response: &har.Response{
			Status:      503,
			StatusText:  "execution error",
			HTTPVersion: "1.1",
			Cookies:     []har.Cookie{},
			Headers: []har.NameValuePair{
				{
					Name:  "Content-Type",
					Value: "application/json",
				},
			},
			Content: &har.Content{
				Size:        82,
				Compression: 0,
				MimeType:    "application/json",
				Text:        "{\\\"ambit\\\":\\\"endpoint01\\\",\\\"step\\\":\\\"endpoint01\\\",\\\"timestamp\\\":\\\"2023-02-12T20:07:02+01:00\\\"}",
				Encoding:    "",
				Comment:     "",
				Data:        nil,
			},
			RedirectURL: "",
			HeadersSize: -1,
			BodySize:    82,
			Comment:     "",
		},
		Cache: nil,
		Timings: &har.Timings{
			Blocked: -1,
			DNS:     -1,
			Connect: -1,
			Send:    -1,
			Wait:    5,
			Receive: -1,
			Ssl:     -1,
			Comment: "",
		},
		ServerIPAddress: "",
		Connection:      "",
		Comment:         "",
		TraceId:         "",
	},
	{
		StartedDateTime: "2023-02-12T20:07:02.149311+01:00",
		Time:            3,
		Request: &har.Request{
			Method:      "GET",
			URL:         "http://localhost:3004/example-01/api/v1/orc-01/endpoint-01/10724279",
			HTTPVersion: "1.1",
			Cookies:     []har.Cookie{},
			Headers: []har.NameValuePair{
				{
					Name:  "Requestid",
					Value: "a-resquest-id",
				},
				{
					Name:  "Content-Type",
					Value: "application/json",
				},
			},
			QueryString: har.NameValuePairs{},
			HeadersSize: -1,
			BodySize:    -1,
			Comment:     "",
		},
		Response: &har.Response{
			Status:      503,
			StatusText:  "Service Unavailable GEN",
			HTTPVersion: "1.1",
			Cookies:     []har.Cookie{},
			Headers:     []har.NameValuePair{},
			Content: &har.Content{
				Size:     0,
				MimeType: "",
			},
			RedirectURL: "",
			HeadersSize: -1,
			BodySize:    0,
			Comment:     "",
		},
		Cache: nil,
		Timings: &har.Timings{
			Blocked: -1,
			DNS:     -1,
			Connect: -1,
			Send:    -1,
			Wait:    3,
			Receive: -1,
			Ssl:     -1,
			Comment: "",
		},
		ServerIPAddress: "",
		Connection:      "",
		Comment:         "endpoint01-1",
		TraceId:         "",
	},
}

func TestLogZeroTracer(t *testing.T) {

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	headers := http.Header{
		"Content-type": []string{"application/json"},
	}

	tracer, _ := logzerotracer.NewTracer()

	// the extraction should fail. the header is not there
	sctx, err := tracer.Extract("", hartracing.HTTPHeadersCarrier(headers))
	require.Error(t, err)

	s := tracer.StartSpan()
	s.Finish()

	// inject the header
	err = tracer.Inject(s.Context(), hartracing.HTTPHeadersCarrier(headers))
	require.NoError(t, err)

	// the extraction should work. the header has been injected
	sctx, err = tracer.Extract("", hartracing.HTTPHeadersCarrier(headers))
	require.NoError(t, err)
	require.Equal(t, sctx.Id(), s.Id())

	ns := tracer.StartSpan(hartracing.ChildOf(s.Context()))
	ns.Finish()

	spanContext, err := tracer.Extract("", hartracing.StringCarrier(s.Id()))
	require.NoError(t, err)

	ns2 := tracer.StartSpan(hartracing.ChildOf(spanContext))
	ns2.AddEntry(harEntries[0])
	ns2.AddEntry(harEntries[1])
	ns2.Finish()

}

func TestFileTracer(t *testing.T) {

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	headers := http.Header{
		"Content-type": []string{"application/json"},
	}

	tracer, c := filetracer.NewTracer("")
	defer c.Close()

	// the extraction should fail. the header is not there
	sctx, err := tracer.Extract("", hartracing.HTTPHeadersCarrier(headers))
	require.Error(t, err)

	// get something like.. 63ed27f8c936e45136000001::63ed27f8c936e45136000001
	s := tracer.StartSpan()
	s.Finish()

	// inject the header
	err = tracer.Inject(s.Context(), hartracing.HTTPHeadersCarrier(headers))
	require.NoError(t, err)

	// the extraction should work. the header has been injected
	sctx, err = tracer.Extract("", hartracing.HTTPHeadersCarrier(headers))
	require.NoError(t, err)
	require.Equal(t, sctx.Id(), s.Id())

	// get something like 63ed27f8c936e45136000001:63ed27f8c936e45136000001:63ed27f8c936e45136000002
	ns := tracer.StartSpan(hartracing.ChildOf(s.Context()))
	ns.Finish()

	spanContext, err := tracer.Extract("", hartracing.StringCarrier(ns.Id()))
	require.NoError(t, err)

	// get something like 63ed27f8c936e45136000001:63ed27f8c936e45136000002:63ed27f8c936e45136000003
	ns2 := tracer.StartSpan(hartracing.ChildOf(spanContext))
	var ns2Entries = []har.Entry{*harEntries[0], *harEntries[1]}
	ns2.AddEntry(&ns2Entries[0])
	ns2.AddEntry(&ns2Entries[1])
	ns2.Finish()

	// get something like 63ed27f8c936e45136000001:63ed27f8c936e45136000002:63ed27f8c936e45136000004
	ns3 := tracer.StartSpan(hartracing.ChildOf(spanContext))
	var ns3Entries = []har.Entry{*harEntries[0], *harEntries[1]}
	ns3.AddEntry(&ns3Entries[0])
	ns3.AddEntry(&ns3Entries[1])
	ns3.Finish()

	time.Sleep(10 * time.Second)
}
