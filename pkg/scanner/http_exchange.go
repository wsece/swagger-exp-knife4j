// http_exchange.go: Burp-style HTTP capture (meta + request + response), serialization, legacy row compat.
package scanner

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

// HTTPHeaderKV is one HTTP header (name/value).
type HTTPHeaderKV struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// HTTPMessagePart is the body of one HTTP request or response.
type HTTPMessagePart struct {
	Method      string         `json:"method,omitempty"`
	URL         string         `json:"url,omitempty"`
	HTTPVersion string         `json:"http_version,omitempty"`
	Status      int            `json:"status,omitempty"`
	StatusLine  string         `json:"status_line,omitempty"`
	Headers     []HTTPHeaderKV `json:"headers,omitempty"`
	Body        string         `json:"body,omitempty"`
	BodyLength  int            `json:"body_length,omitempty"`
	ContentType string         `json:"content_type,omitempty"`
}

// HTTPExchangeMeta is context for one API probe (Swagger path, duration, failure reason, etc.).
type HTTPExchangeMeta struct {
	SwaggerURL       string `json:"swagger_url,omitempty"`
	ResolvedJSONURL  string `json:"resolved_json_url,omitempty"`
	Host             string `json:"host,omitempty"`
	Path             string `json:"path,omitempty"`
	Method           string `json:"method,omitempty"`
	ParameterNames   string `json:"parameter_names,omitempty"`
	ParametersSummary string `json:"parameters_summary,omitempty"`
	DurationMs       int64  `json:"duration_ms,omitempty"`
	Failed           bool   `json:"failed"`
	Error            string `json:"error,omitempty"`
	UnauthorizedRisk bool   `json:"unauthorized_risk,omitempty"`
}

// HTTPExchange is a Burp-style single-object capture: meta + request + response.
type HTTPExchange struct {
	Meta     HTTPExchangeMeta `json:"meta"`
	Request  HTTPMessagePart  `json:"request"`
	Response HTTPMessagePart  `json:"response"`
}

func headersToKV(h http.Header) []HTTPHeaderKV {
	if len(h) == 0 {
		return nil
	}
	out := make([]HTTPHeaderKV, 0, len(h))
	for name, vals := range h {
		for _, v := range vals {
			out = append(out, HTTPHeaderKV{Name: name, Value: v})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		ni, nj := strings.ToLower(out[i].Name), strings.ToLower(out[j].Name)
		if ni != nj {
			return ni < nj
		}
		return out[i].Value < out[j].Value
	})
	return out
}

func statusLine(code int) string {
	if code == 0 {
		return ""
	}
	text := http.StatusText(code)
	if text == "" {
		return strconv.Itoa(code)
	}
	return fmt.Sprintf("%d %s", code, text)
}

// BuildHTTPExchange builds a full HTTPExchange from a report row and request result.
func BuildHTTPExchange(rec SwaggerReportRecord, req APIRequestResult) HTTPExchange {
	failed := rec.Failed || strings.TrimSpace(req.Error) != ""
	ex := HTTPExchange{
		Meta: HTTPExchangeMeta{
			SwaggerURL:        rec.SwaggerURL,
			ResolvedJSONURL:   rec.ResolvedJSONURL,
			Host:              rec.Host,
			Path:              rec.Path,
			Method:            rec.Method,
			ParameterNames:    rec.ParameterNames,
			ParametersSummary: rec.ParametersSummary,
			DurationMs:        req.DurationMs,
			Failed:            failed,
			Error:             rec.Error,
			UnauthorizedRisk:  rec.UnauthorizedRisk,
		},
		Request: HTTPMessagePart{
			Method:      strings.ToUpper(rec.Method),
			URL:         firstNonEmpty(req.FinalURL, req.FullURL, rec.FullURL),
			HTTPVersion: "HTTP/1.1",
			Headers:     req.RequestHeaders,
			Body:        requestBodyForPacket(req, rec),
			ContentType: headerValue(req.RequestHeaders, "Content-Type"),
		},
		Response: HTTPMessagePart{
			HTTPVersion: "HTTP/1.1",
			Status:      req.StatusCode,
			StatusLine:  statusLine(req.StatusCode),
			Headers:     req.ResponseHeaders,
			Body:        firstNonEmpty(req.Response, rec.Response),
			ContentType: firstNonEmpty(req.ContentType, rec.ContentType, headerValue(req.ResponseHeaders, "Content-Type")),
		},
	}
	if ex.Request.Body != "" {
		ex.Request.BodyLength = len(ex.Request.Body)
	}
	if ex.Response.Body != "" {
		ex.Response.BodyLength = len(ex.Response.Body)
	}
	if len(ex.Request.Headers) == 0 && (rec.Cookie != "" || rec.UserAgent != "" || rec.Authorization != "") {
		ex.Request.Headers = legacyAuthHeaders(rec)
	}
	return ex
}

func requestBodyForPacket(req APIRequestResult, rec SwaggerReportRecord) string {
	if req.RequestBody != "" {
		return req.RequestBody
	}
	if rec.RequestBody != "" {
		return rec.RequestBody
	}
	if req.RequestParams != "" {
		return ""
	}
	return ""
}

func legacyAuthHeaders(rec SwaggerReportRecord) []HTTPHeaderKV {
	var h []HTTPHeaderKV
	if rec.Cookie != "" {
		h = append(h, HTTPHeaderKV{Name: "Cookie", Value: rec.Cookie})
	}
	if rec.UserAgent != "" {
		h = append(h, HTTPHeaderKV{Name: "User-Agent", Value: rec.UserAgent})
	}
	if rec.Authorization != "" {
		h = append(h, HTTPHeaderKV{Name: "Authorization", Value: rec.Authorization})
	}
	return h
}

func headerValue(headers []HTTPHeaderKV, name string) string {
	lower := strings.ToLower(name)
	for _, h := range headers {
		if strings.ToLower(h.Name) == lower {
			return h.Value
		}
	}
	return ""
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

// MarshalHTTPExchange serializes the capture to JSON for models.SwaggerAPIRecord.PacketJSON.
func MarshalHTTPExchange(ex HTTPExchange) (string, error) {
	b, err := json.Marshal(ex)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// UnmarshalHTTPExchange deserializes a capture from DB PacketJSON.
func UnmarshalHTTPExchange(raw string) (HTTPExchange, error) {
	var ex HTTPExchange
	if raw == "" {
		return ex, nil
	}
	err := json.Unmarshal([]byte(raw), &ex)
	return ex, err
}

// HTTPExchangeFromLegacyRecord rebuilds a capture from legacy flat columns (no PacketJSON) for report UI.
func HTTPExchangeFromLegacyRecord(
	inputURL, method, host, api string,
	cookie, userAgent, authorization string,
	requestBody, requestParams, responseBody string,
	statusCode int,
	failed bool,
	failedReason string,
) HTTPExchange {
	fullURL := ""
	if host != "" && api != "" {
		fullURL = "https://" + host + api
		if requestParams != "" {
			if strings.Contains(api, "?") {
				fullURL += "&" + requestParams
			} else {
				fullURL += "?" + requestParams
			}
		}
	}
	reqHeaders := legacyAuthHeaders(SwaggerReportRecord{
		Cookie: cookie, UserAgent: userAgent, Authorization: authorization,
	})
	rec := SwaggerReportRecord{
		SwaggerURL: inputURL,
		Method:     method,
		Path:       api,
		Host:       host,
		RequestBody: requestBody,
		RequestParams: requestParams,
		FullURL:    fullURL,
		StatusCode: statusCode,
		Response:   responseBody,
		Failed:     failed,
		Error:      failedReason,
		UnauthorizedRisk: !failed && statusCode != 0 && statusCode != 401,
	}
	req := APIRequestResult{
		Method: method, Path: api, Host: host,
		FullURL: fullURL, FinalURL: fullURL,
		RequestBody: requestBody, RequestParams: requestParams,
		StatusCode: statusCode, Response: responseBody,
		Error: failedReason,
	}
	if requestBody != "" {
		reqHeaders = append(reqHeaders, HTTPHeaderKV{Name: "Content-Type", Value: "application/json"})
	}
	req.RequestHeaders = reqHeaders
	return BuildHTTPExchange(rec, req)
}
