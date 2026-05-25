package reportserver

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
)

func TestKnife4jUseProxy(t *testing.T) {
	s := &Server{cfg: ServerConfig{Knife4jProxy: false}}

	r := httptest.NewRequest("GET", "/knife4j-openapi.json?session=x&proxy=1", nil)
	if !s.knife4jUseProxy(r) {
		t.Fatal("expected proxy=1")
	}

	r = httptest.NewRequest("GET", "/knife4j-openapi.json?session=x&proxy=0", nil)
	if s.knife4jUseProxy(r) {
		t.Fatal("expected direct with proxy=0")
	}

	s.cfg.Knife4jProxy = true
	r = httptest.NewRequest("GET", "/knife4j-openapi.json?session=x", nil)
	if !s.knife4jUseProxy(r) {
		t.Fatal("expected server default proxy")
	}
}

func TestNormalizeOpenAPISpec_directOrigin(t *testing.T) {
	raw := []byte(`{"openapi":"3.0.1","paths":{}}`)
	out := NormalizeOpenAPISpec(raw, "https://api.example.com")
	var doc map[string]interface{}
	if err := json.Unmarshal(out, &doc); err != nil {
		t.Fatal(err)
	}
	servers, ok := doc["servers"].([]interface{})
	if !ok || len(servers) == 0 {
		t.Fatal("expected servers")
	}
	url := servers[0].(map[string]interface{})["url"].(string)
	if url != "https://api.example.com" {
		t.Fatalf("got %q", url)
	}
}