package har_test

import (
	"encoding/json"
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-har/har"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

var harLog = har.HAR{
	Log: &har.Log{
		Version: "1.1",
		Comment: "ex01_orc01",
		Creator: &har.Creator{
			Name:    "tpm-har",
			Version: "0.0.2",
			Comment: "test case",
		},
		Browser: &har.Creator{
			Name:    "",
			Version: "",
			Comment: "",
		},
		Entries: []*har.Entry{
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
		},
	},
}

func TestHarLog(t *testing.T) {

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	b, err := json.Marshal(harLog)
	require.NoError(t, err)

	t.Log(string(b))
}
