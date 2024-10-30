package har

import (
	"encoding/json"
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-common/util"
	"github.com/rs/zerolog/log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	HttpScheme = "http"
	Localhost  = "localhost"
)

type PersonallyIdentifiableInformation struct {
	Domain    string `json:"domain,omitempty"`
	AppliesTo string `json:"applies-to,omitempty"`
}

func (pii *PersonallyIdentifiableInformation) ShouldMaskRequest() bool {
	return strings.Contains(pii.AppliesTo, "req")
}

func (pii *PersonallyIdentifiableInformation) ShouldMaskResponse() bool {
	return strings.Contains(pii.AppliesTo, "resp")
}

type PIIMasker interface {
	Mask(domain string, data []byte) ([]byte, error)
}

// Cache contains info about a request coming from browser cache.
//
// See: https://chromedevtools.github.io/devtools-protocol/tot/HAR#type-Cache
type Cache struct {
	BeforeRequest *CacheData `json:"beforeRequest,omitempty" yaml:"beforeRequest,omitempty" mapstructure:"beforeRequest,omitempty"` // State of a cache entry before the request. Leave out this field if the information is not available.
	AfterRequest  *CacheData `json:"afterRequest,omitempty" yaml:"afterRequest,omitempty" mapstructure:"afterRequest,omitempty"`    // State of a cache entry after the request. Leave out this field if the information is not available.
	Comment       string     `json:"comment,omitempty" yaml:"comment,omitempty" mapstructure:"comment,omitempty"`                   // A comment provided by the user or the application.
}

// CacheData describes the cache data for beforeRequest and afterRequest.
//
// See: https://chromedevtools.github.io/devtools-protocol/tot/HAR#type-CacheData
type CacheData struct {
	Expires    string `json:"expires,omitempty" yaml:"expires,omitempty" mapstructure:"expires,omitempty"` // Expiration time of the cache entry.
	LastAccess string `json:"lastAccess" yaml:"lastAccess" mapstructure:"lastAccess"`                      // The last time the cache entry was opened.
	ETag       string `json:"eTag" yaml:"eTag" mapstructure:"eTag"`                                        // Etag
	HitCount   int64  `json:"hitCount" yaml:"hitCount" mapstructure:"hitCount"`                            // The number of times the cache entry has been opened.
	Comment    string `json:"comment,omitempty" yaml:"comment,omitempty" mapstructure:"comment,omitempty"` // A comment provided by the user or the application.
}

// Content describes details about response content (embedded in [Response]
// object).
//
// See: https://chromedevtools.github.io/devtools-protocol/tot/HAR#type-Content
type Content struct {
	Size        int64  `json:"size" yaml:"size" mapstructure:"size"`                                                    // Length of the returned content in bytes. Should be equal to response.bodySize if there is no compression and bigger when the content has been compressed.
	Compression int64  `json:"compression,omitempty" yaml:"compression,omitempty" mapstructure:"compression,omitempty"` // Number of bytes saved. Leave out this field if the information is not available.
	MimeType    string `json:"mimeType" yaml:"mimeType" mapstructure:"mimeType"`                                        // MIME type of the response text (value of the Content-Type response header). The charset attribute of the MIME type is included (if available).
	Text        string `json:"text,omitempty" yaml:"text,omitempty" mapstructure:"text,omitempty"`                      // Response body sent from the server or loaded from the browser cache. This field is populated with textual content only. The text field is either HTTP decoded text or an encoded (e.g. "base64") representation of the response body. Leave out this field if the information is not available.
	Encoding    string `json:"encoding,omitempty" yaml:"encoding,omitempty" mapstructure:"encoding,omitempty"`          // Encoding used for response text field e.g "base64". Leave out this field if the text field is HTTP decoded (decompressed & unchunked), than trans-coded from its original character set into UTF-8.
	Comment     string `json:"comment,omitempty" yaml:"comment,omitempty" mapstructure:"comment,omitempty"`             // A comment provided by the user or the application.
	Data        []byte `json:"-" yaml:"-" mapstructure:"-"`                                                             // the bytes of the text data...
}

func (c *Content) MarshalJSON() ([]byte, error) {
	type content Content
	if len(c.Data) > 0 {
		c.Text = string(c.Data)
	}
	return json.Marshal((*content)(c))
}

// Cookie contains list of all cookies (used in [Request] and [Response]
// objects).
//
// See: https://chromedevtools.github.io/devtools-protocol/tot/HAR#type-Cookie
type Cookie struct {
	Name     string `json:"name" yaml:"name" mapstructure:"name"`                                           // The name of the cookie.
	Value    string `json:"value" yaml:"value" mapstructure:"value"`                                        // The cookie value.
	Path     string `json:"path,omitempty" yaml:"path,omitempty" mapstructure:"path,omitempty"`             // The path pertaining to the cookie.
	Domain   string `json:"domain,omitempty" yaml:"domain,omitempty" mapstructure:"domain,omitempty"`       // The host of the cookie.
	Expires  string `json:"expires,omitempty" yaml:"expires,omitempty" mapstructure:"expires,omitempty"`    // Cookie expiration time. (ISO 8601 - YYYY-MM-DDThh:mm:ss.sTZD, e.g. 2009-07-24T19:20:30.123+02:00).
	HTTPOnly bool   `json:"httpOnly,omitempty" yaml:"httpOnly,omitempty" mapstructure:"httpOnly,omitempty"` // Set to true if the cookie is HTTP only, false otherwise.
	Secure   bool   `json:"secure,omitempty" yaml:"secure,omitempty" mapstructure:"secure,omitempty"`       // True if the cookie was transmitted over ssl, false otherwise.
	Comment  string `json:"comment,omitempty" yaml:"comment,omitempty" mapstructure:"comment,omitempty"`    // A comment provided by the user or the application.
}

// Creator and browser objects share the same structure.
//
// See: https://chromedevtools.github.io/devtools-protocol/tot/HAR#type-Creator
type Creator struct {
	Name    string `json:"name" yaml:"name" mapstructure:"name"`                                        // Name of the application/browser used to export the log.
	Version string `json:"version" yaml:"version" mapstructure:"version"`                               // Version of the application/browser used to export the log.
	Comment string `json:"comment,omitempty" yaml:"comment,omitempty" mapstructure:"comment,omitempty"` // A comment provided by the user or the application.
}

func (c Creator) IsZero() bool {
	return c.Name == "" && c.Version == "" && c.Comment == ""
}

// Entry represents an array with all exported HTTP requests. Sorting entries
// by startedDateTime (starting from the oldest) is preferred way how to export
// data since it can make importing faster. However, the reader application
// should always make sure the array is sorted (if required for the import).
//
// See: https://chromedevtools.github.io/devtools-protocol/tot/HAR#type-Entry
type Entry struct {
	Pageref         string                            `json:"pageref,omitempty" yaml:"pageref,omitempty" mapstructure:"pageref,omitempty"`                         // Reference to the parent page. Leave out this field if the application does not support grouping by pages.
	StartedDateTime string                            `json:"startedDateTime" yaml:"startedDateTime" mapstructure:"startedDateTime"`                               // Date and time stamp of the request start (ISO 8601 - YYYY-MM-DDThh:mm:ss.sTZD).
	StartDateTimeTm time.Time                         `json:"-" yaml:"-" mapstructure:"-"`                                                                         // keep the time in native format to avoid issues in evaluating elapsed...
	Time            float64                           `json:"time" yaml:"time" mapstructure:"time"`                                                                // Total elapsed time of the request in milliseconds. This is the sum of all timings available in the timings object (i.e. not including -1 values) .
	Request         *Request                          `json:"request" yaml:"request" mapstructure:"request"`                                                       // Detailed info about the request.
	Response        *Response                         `json:"response" yaml:"response" mapstructure:"response"`                                                    // Detailed info about the response.
	Cache           *Cache                            `json:"cache" yaml:"cache" mapstructure:"cache"`                                                             // Info about cache usage.
	Timings         *Timings                          `json:"timings" yaml:"timings" mapstructure:"timings"`                                                       // Detailed timing info about request/response round trip.
	ServerIPAddress string                            `json:"serverIPAddress,omitempty" yaml:"serverIPAddress,omitempty" mapstructure:"serverIPAddress,omitempty"` // IP address of the server that was connected (result of DNS resolution).
	Connection      string                            `json:"connection,omitempty" yaml:"connection,omitempty" mapstructure:"connection,omitempty"`                // Unique ID of the parent TCP/IP connection, can be the client or server port number. Note that a port number doesn't have to be unique identifier in cases where the port is shared for more connections. If the port isn't available for the application, any other unique connection ID can be used instead (e.g. connection index). Leave out this field if the application doesn't support this info.
	Comment         string                            `json:"comment,omitempty" yaml:"comment,omitempty" mapstructure:"comment,omitempty"`                         // A comment provided by the user or the application.
	PII             PersonallyIdentifiableInformation `json:"_pii,omitempty"`                                                                                      // Extension field to identify sensistive information handling for logging
	TraceId         string                            `json:"_trace-id,omitempty" yaml:"_trace-id,omitempty" mapstructure:"_trace-id,omitempty"`
}

func (e *Entry) MaskRequestBody(jm PIIMasker) error {

	const semLogContext = "har::mask-request-body"
	if !util.IsNilish(jm) && e.PII.ShouldMaskRequest() && e.Request != nil && e.Request.HasBody() {
		log.Trace().Str("comment", e.Comment).Msg(semLogContext)
		masked, err := jm.Mask(e.PII.Domain, e.Request.PostData.Data)
		if err != nil {
			return err
		}

		e.Request.PostData.Data = masked
	}

	return nil
}

func (e *Entry) MaskResponseBody(jm PIIMasker) error {
	const semLogContext = "har::mask-response-body"

	if !util.IsNilish(jm) && e.PII.ShouldMaskResponse() && e.Response != nil && e.Response.HasBody() {
		log.Trace().Str("comment", e.Comment).Msg(semLogContext)
		masked, err := jm.Mask(e.PII.Domain, e.Response.Content.Data)
		if err != nil {
			return err
		}

		e.Response.Content.Data = masked
	}

	return nil
}

// HAR parent container for HAR log.
//
// See: https://chromedevtools.github.io/devtools-protocol/tot/HAR#type-HAR
type HAR struct {
	Log *Log `json:"log"`
}

// Log represents the root of exported data.
//
// See: https://chromedevtools.github.io/devtools-protocol/tot/HAR#type-Log
type Log struct {
	Version string   `json:"version" yaml:"version" mapstructure:"version"`                               // Version number of the format. If empty, string "1.1" is assumed by default.
	Creator *Creator `json:"creator" yaml:"creator" mapstructure:"creator"`                               // Name and version info of the log creator application.
	Browser *Creator `json:"browser,omitempty" yaml:"browser,omitempty" mapstructure:"browser,omitempty"` // Name and version info of used browser.
	Pages   []*Page  `json:"pages,omitempty" yaml:"pages,omitempty" mapstructure:"pages,omitempty"`       // List of all exported (tracked) pages. Leave out this field if the application does not support grouping by pages.
	Entries []*Entry `json:"entries" yaml:"entries" mapstructure:"entries"`                               // List of all exported (tracked) requests.
	Comment string   `json:"comment,omitempty" yaml:"comment,omitempty" mapstructure:"comment,omitempty"` // A comment provided by the user or the application.
	TraceId string   `json:"_trace-id,omitempty" yaml:"_trace-id,omitempty" mapstructure:"_trace-id,omitempty"`
}

func (log *Log) FindEarliestStartedDateTime() string {
	min := time.Now().Format(time.RFC3339Nano)
	for _, e := range log.Entries {
		if e.StartedDateTime != "" && e.StartedDateTime < min {
			min = e.StartedDateTime
		}
	}

	return min
}

type NameValuePairs []NameValuePair

// NameValuePair describes a name/value pair.
//
// See: https://chromedevtools.github.io/devtools-protocol/tot/HAR#type-NameValuePair
type NameValuePair struct {
	Name    string `json:"name" yaml:"name" mapstructure:"name"`                                        // Name of the pair.
	Value   string `json:"value" yaml:"value" mapstructure:"value"`                                     // Value of the pair.
	Comment string `json:"comment,omitempty" yaml:"comment,omitempty" mapstructure:"comment,omitempty"` // A comment provided by the user or the application.
}

func (nvs NameValuePairs) GetFirst(n string) NameValuePair {

	n = strings.ToLower(n)
	for _, nv := range nvs {
		if strings.ToLower(nv.Name) == n {
			return nv
		}
	}
	return NameValuePair{}
}

// Page represents list of exported pages.
//
// See: https://chromedevtools.github.io/devtools-protocol/tot/HAR#type-Page
type Page struct {
	StartedDateTime string       `json:"startedDateTime" yaml:"startedDateTime" mapstructure:"startedDateTime"`       // Date and time stamp for the beginning of the page load (ISO 8601 - YYYY-MM-DDThh:mm:ss.sTZD, e.g. 2009-07-24T19:20:30.45+01:00).
	ID              string       `json:"id" yaml:"id" mapstructure:"id"`                                              // Unique identifier of a page within the [Log]. Entries use it to refer the parent page.
	Title           string       `json:"title" yaml:"title" mapstructure:"title"`                                     // Page title.
	PageTimings     *PageTimings `json:"pageTimings" yaml:"pageTimings" mapstructure:"pageTimings"`                   // Detailed timing info about page load.
	Comment         string       `json:"comment,omitempty" yaml:"comment,omitempty" mapstructure:"comment,omitempty"` // A comment provided by the user or the application.
}

// PageTimings describes timings for various events (states) fired during the
// page load. All times are specified in milliseconds. If a time info is not
// available appropriate field is set to -1.
//
// See: https://chromedevtools.github.io/devtools-protocol/tot/HAR#type-PageTimings
type PageTimings struct {
	OnContentLoad float64 `json:"onContentLoad,omitempty" yaml:"onContentLoad,omitempty" mapstructure:"onContentLoad,omitempty"` // Content of the page loaded. Number of milliseconds since page load started (page.startedDateTime). Use -1 if the timing does not apply to the current request.
	OnLoad        float64 `json:"onLoad,omitempty" yaml:"onLoad,omitempty" mapstructure:"onLoad,omitempty"`                      // Page is loaded (onLoad event fired). Number of milliseconds since page load started (page.startedDateTime). Use -1 if the timing does not apply to the current request.
	Comment       string  `json:"comment,omitempty" yaml:"comment,omitempty" mapstructure:"comment,omitempty"`                   // A comment provided by the user or the application.
}

// Param list of posted parameters, if any (embedded in [PostData] object).
//
// See: https://chromedevtools.github.io/devtools-protocol/tot/HAR#type-Param
type Params []Param

func (pars Params) GetFirst(n string) Param {
	for _, nv := range pars {
		if nv.Name == n {
			return nv
		}
	}
	return Param{}
}

type Param struct {
	Name        string `json:"name" yaml:"name" mapstructure:"name"`                                                    // name of a posted parameter.
	Value       string `json:"value,omitempty" yaml:"value,omitempty" mapstructure:"value,omitempty"`                   // value of a posted parameter or content of a posted file.
	FileName    string `json:"fileName,omitempty" yaml:"fileName,omitempty" mapstructure:"fileName,omitempty"`          // name of a posted file.
	ContentType string `json:"contentType,omitempty" yaml:"contentType,omitempty" mapstructure:"contentType,omitempty"` // content-type of a posted file.
	Comment     string `json:"comment,omitempty" yaml:"comment,omitempty" mapstructure:"comment,omitempty"`             // A comment provided by the user or the application.
}

// PostData describes posted data, if any (embedded in [Request] object).
//
// See: https://chromedevtools.github.io/devtools-protocol/tot/HAR#type-PostData
type PostData struct {
	MimeType string  `json:"mimeType" yaml:"mimeType" mapstructure:"mimeType"`                            // Mime type of posted data.
	Params   []Param `json:"params" yaml:"params" mapstructure:"params"`                                  // List of posted parameters (in case of URL encoded parameters).
	Text     string  `json:"text" yaml:"text" mapstructure:"text"`                                        // Plain text posted data
	Comment  string  `json:"comment,omitempty" yaml:"comment,omitempty" mapstructure:"comment,omitempty"` // A comment provided by the user or the application.
	Data     []byte  `json:"-" yaml:"-" mapstructure:"-"`                                                 // the bytes of the text data...
}

func (po *PostData) MarshalJSON() ([]byte, error) {
	type postdata PostData
	if po.Data != nil {
		//var b []byte
		//var ok bool
		//var err error
		//if b, ok = po.Data.([]byte); !ok {
		//  b, err = jsoniter.Marshal(po.Data)
		//  if err != nil {
		//    log.Error().Err(err).Send()
		//  }
		//}
		po.Text = string(po.Data)
	}
	return json.Marshal((*postdata)(po))
}

// Request contains detailed info about performed request.
//
// See: https://chromedevtools.github.io/devtools-protocol/tot/HAR#type-Request
type Request struct {
	Method      string         `json:"method,omitempty" yaml:"method,omitempty" mapstructure:"method,omitempty"`                // Request method (GET, POST, ...).
	URL         string         `json:"url,omitempty" yaml:"url,omitempty" mapstructure:"url,omitempty"`                         // Absolute URL of the request (fragments are not included).
	PathParams  []Param        `json:"_pathParams,omitempty" yaml:"_pathParams,omitempty" mapstructure:"_pathParams,omitempty"` // Absolute URL of the request (fragments are not included).
	HTTPVersion string         `json:"httpVersion,omitempty" yaml:"httpVersion,omitempty" mapstructure:"httpVersion,omitempty"` // Request HTTP Version.
	Cookies     []Cookie       `json:"cookies" yaml:"cookies" mapstructure:"cookies"`                                           // List of cookie objects.
	Headers     NameValuePairs `json:"headers,omitempty" yaml:"headers,omitempty" mapstructure:"headers,omitempty"`             // List of header objects.
	QueryString NameValuePairs `json:"queryString" yaml:"queryString" mapstructure:"queryString"`                               // List of query parameter objects.
	PostData    *PostData      `json:"postData,omitempty" yaml:"postData,omitempty" mapstructure:"postData,omitempty"`          // Posted data info.
	HeadersSize int64          `json:"headersSize" yaml:"headersSize" mapstructure:"headersSize"`                               // Total number of bytes from the start of the HTTP request message until (and including) the double CRLF before the body. Set to -1 if the info is not available.
	BodySize    int64          `json:"bodySize" yaml:"bodySize" mapstructure:"bodySize"`                                        // Size of the request body (POST data payload) in bytes. Set to -1 if the info is not available.
	Comment     string         `json:"comment,omitempty" yaml:"comment,omitempty" mapstructure:"comment,omitempty"`             // A comment provided by the user or the application.
}

type UrlBuilder struct {
	Scheme   string `json:"-"`
	Hostname string `json:"-"`
	Port     int    `json:"-"`
	Path     string `json:"-"`
}

func (ub *UrlBuilder) WithPath(p string) {
	ub.Path = p
}

func (ub *UrlBuilder) WithPort(p int) {
	ub.Port = p
}

func (ub *UrlBuilder) WithScheme(p string) {
	ub.Scheme = p
}

func (ub *UrlBuilder) WithHostname(p string) {
	ub.Hostname = p
}

func (ub *UrlBuilder) Url() string {

	var sb strings.Builder

	if ub.Scheme == "" {
		ub.Scheme = HttpScheme
	}

	sb.WriteString(ub.Scheme)
	sb.WriteString("://")

	if ub.Hostname != "" {
		sb.WriteString(ub.Hostname)
	} else {
		sb.WriteString(Localhost)
	}

	if ub.Port != 0 {
		sb.WriteString(":")
		sb.WriteString(strconv.Itoa(ub.Port))
	}

	if ub.Path == "" {
		ub.Path = "/"
	}

	sb.WriteString(ub.Path)

	return sb.String()
}

type RequestOption func(o *Request)

func WithUrl(p string) RequestOption {
	return func(o *Request) {
		o.URL = p
	}
}

func WithMethod(p string) RequestOption {
	return func(o *Request) {
		o.Method = p
	}
}

func WithBody(p []byte) RequestOption {
	return func(o *Request) {
		o.PostData = &PostData{
			MimeType: "application/json",
			Data:     p,
			Params:   []Param{},
		}
	}
}

func WithHeader(h NameValuePair) RequestOption {
	return func(o *Request) {
		o.Headers = append(o.Headers, h)
	}
}

func WithQueryParam(q NameValuePair) RequestOption {
	return func(o *Request) {
		o.QueryString = append(o.QueryString, q)
	}
}

func (req *Request) HasBody() bool {
	return req.PostData != nil && len(req.PostData.Data) > 0
}

func (req *Request) String() string {
	b, err := json.Marshal(req)
	if err != nil {
		return err.Error()
	}

	return string(b)
}

func (req *Request) SetHeader(n string, v string) {

	n = strings.ToLower(n)
	for i, nv := range req.Headers {
		if strings.ToLower(nv.Name) == n {
			req.Headers[i].Value = v
			return
		}
	}

	req.Headers = append(req.Headers, NameValuePair{
		Name:  n,
		Value: v,
	})
}

// NewRequest introduced when migrating the tpm-symphony. Revised the implementation to take care of different postData build
func NewRequest(method string, url string, body []byte, headers http.Header, params []Param) (*Request, error) {

	ct := headers.Get("content-type")

	var hs []NameValuePair
	for n, h := range headers {
		for i := range h {
			hs = append(hs, NameValuePair{Name: n, Value: h[i]})
		}
	}

	var postData *PostData
	bodySize := -1

	switch method {
	case http.MethodGet:
		if len(params) != 0 {
			postData = &PostData{
				MimeType: "",
				Data:     nil,
				Params:   params,
			}
		}
	case http.MethodPut:
		fallthrough
	case http.MethodPost:
		if len(params) != 0 || len(body) > 0 {
			if len(body) > 0 {
				bodySize = len(body)
			}

			if params == nil {
				params = make([]Param, 0)
			}

			postData = &PostData{
				MimeType: ct,
				Data:     body,
				Params:   params,
			}
		}
	}

	req := &Request{
		Method:      method,
		URL:         url,
		PathParams:  params,
		HTTPVersion: "1.1",
		Headers:     hs,
		HeadersSize: -1,
		Cookies:     []Cookie{},
		QueryString: []NameValuePair{},
		BodySize:    int64(bodySize),
		PostData:    postData,
	}

	return req, nil
}

// Response contains detailed info about the response.
//
// See: https://chromedevtools.github.io/devtools-protocol/tot/HAR#type-Response
type Response struct {
	Status      int            `json:"status,omitempty" yaml:"status,omitempty" mapstructure:"status,omitempty"`                // Response status.
	StatusText  string         `json:"statusText,omitempty" yaml:"statusText,omitempty" mapstructure:"statusText,omitempty"`    // Response status description.
	HTTPVersion string         `json:"httpVersion,omitempty" yaml:"httpVersion,omitempty" mapstructure:"httpVersion,omitempty"` // Response HTTP Version.
	Cookies     []Cookie       `json:"cookies" yaml:"cookies" mapstructure:"cookies"`                                           // List of cookie objects.
	Headers     NameValuePairs `json:"headers" yaml:"headers" mapstructure:"headers"`                                           // List of header objects.
	Content     *Content       `json:"content,omitempty" yaml:"content,omitempty" mapstructure:"content,omitempty"`             // Details about the response body.
	RedirectURL string         `json:"redirectURL" yaml:"redirectURL" mapstructure:"redirectURL"`                               // Redirection target URL from the Location response header.
	HeadersSize int            `json:"headersSize" yaml:"headersSize" mapstructure:"headersSize"`                               // Total number of bytes from the start of the HTTP response message until (and including) the double CRLF before the body. Set to -1 if the info is not available.
	BodySize    int64          `json:"bodySize" yaml:"bodySize" mapstructure:"bodySize"`                                        // Size of the received response body in bytes. Set to zero in case of responses coming from the cache (304). Set to -1 if the info is not available.
	Comment     string         `json:"comment,omitempty" yaml:"comment,omitempty" mapstructure:"comment,omitempty"`             // A comment provided by the user or the application.
}

func (resp *Response) HasBody() bool {
	return resp.Content != nil && len(resp.Content.Data) > 0
}

func NewResponse(sc int, sTest string, mimeType string, body []byte, headers NameValuePairs) *Response {

	if headers == nil {
		headers = make([]NameValuePair, 0)
	}

	headers = append(headers, NameValuePair{Name: "Content-type", Value: mimeType})

	r := &Response{
		Status:      sc,
		HTTPVersion: "1.1",
		StatusText:  sTest,
		HeadersSize: -1,
		Headers:     headers,
		Cookies:     []Cookie{},
		BodySize:    int64(len(body)),
		Content: &Content{
			MimeType: mimeType,
			Size:     int64(len(body)),
			Data:     body,
		},
	}

	return r
}

// Timings describes various phases within request-response round trip. All
// times are specified in milliseconds.
//
// See: https://chromedevtools.github.io/devtools-protocol/tot/HAR#type-Timings
type Timings struct {
	Blocked float64 `json:"blocked,omitempty" yaml:"blocked,omitempty" mapstructure:"blocked,omitempty"` // Time spent in a queue waiting for a network connection. Use -1 if the timing does not apply to the current request.
	DNS     float64 `json:"dns,omitempty" yaml:"dns,omitempty" mapstructure:"dns,omitempty"`             // DNS resolution time. The time required to resolve a host name. Use -1 if the timing does not apply to the current request.
	Connect float64 `json:"connect,omitempty" yaml:"connect,omitempty" mapstructure:"connect,omitempty"` // Time required to create TCP connection. Use -1 if the timing does not apply to the current request.
	Send    float64 `json:"send" yaml:"send" mapstructure:"send"`                                        // Time required to send HTTP request to the server.
	Wait    float64 `json:"wait" yaml:"wait" mapstructure:"wait"`                                        // Waiting for a response from the server.
	Receive float64 `json:"receive" yaml:"receive" mapstructure:"receive"`                               // Time required to read entire response from the server (or cache).
	Ssl     float64 `json:"ssl,omitempty" yaml:"ssl,omitempty" mapstructure:"ssl,omitempty"`             // Time required for SSL/TLS negotiation. If this field is defined then the time is also included in the connect field (to ensure backward compatibility with HAR 1.1). Use -1 if the timing does not apply to the current request.
	Comment string  `json:"comment,omitempty" yaml:"comment,omitempty" mapstructure:"comment,omitempty"` // A comment provided by the user or the application.
}
