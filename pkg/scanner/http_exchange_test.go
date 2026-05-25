// http_exchange_test.go: unit tests for BuildHTTPExchange and statusLine.
package scanner

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestBuildHTTPExchange(t *testing.T) {
	t.Parallel()
	rec := SwaggerReportRecord{
		SwaggerURL:  "https://api.example.com/index.html",
		Method:      "POST",
		Path:        "/api/test/ok",
		Host:        "api.example.com",
		UnauthorizedRisk: true,
	}
	req := APIRequestResult{
		Method:   "POST",
		Path:     "/api/test/ok",
		FinalURL: "https://api.example.com/api/test/ok",
		RequestBody: `{"id":1}`,
		RequestHeaders: []HTTPHeaderKV{
			{Name: "Content-Type", Value: "application/json"},
			{Name: "Host", Value: "api.example.com"},
		},
		StatusCode: 200,
		Response:   `{"ok":true}`,
		ResponseHeaders: []HTTPHeaderKV{
			{Name: "Content-Type", Value: "application/json"},
		},
		DurationMs: 42,
	}
	ex := BuildHTTPExchange(rec, req)
	if ex.Request.URL != req.FinalURL {
		t.Fatalf("request url %q", ex.Request.URL)
	}
	if ex.Request.Body != `{"id":1}` {
		t.Fatalf("request body %q", ex.Request.Body)
	}
	if ex.Response.Status != 200 {
		t.Fatalf("status %d", ex.Response.Status)
	}
	if ex.Meta.DurationMs != 42 {
		t.Fatalf("duration %d", ex.Meta.DurationMs)
	}
	b, err := json.Marshal(ex)
	if err != nil {
		t.Fatal(err)
	}
	var round HTTPExchange
	if err := json.Unmarshal(b, &round); err != nil {
		t.Fatal(err)
	}
	if round.Response.Body != `{"ok":true}` {
		t.Fatal("round-trip body")
	}
}

func TestStatusLine(t *testing.T) {
	t.Parallel()
	if statusLine(http.StatusNotFound) != "404 Not Found" {
		t.Fatal(statusLine(http.StatusNotFound))
	}
}
