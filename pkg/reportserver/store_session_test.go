// store_session_test.go: tests for Store listing api-docs sessions.
package reportserver

import (
	"testing"

	"swagger-exp-knife4j/internal/islazy"
)

func TestParseSessionMeta(t *testing.T) {
	t.Parallel()
	host, scope, at := islazy.ParseSessionMeta("mr.example.com")
	if host != "mr.example.com" || scope != "" || !at.IsZero() {
		t.Fatalf("host-only id: %q %q %v", host, scope, at)
	}
	host, scope, at = islazy.ParseSessionMeta("example.com__auth")
	if host != "example.com" || scope != "auth" {
		t.Fatalf("scoped id: %q %q", host, scope)
	}
	host, _, at = islazy.ParseSessionMeta("20260519_032109_mr.bjxulongkeji.cn")
	if host != "mr.bjxulongkeji.cn" || at.IsZero() {
		t.Fatalf("legacy: %q %v", host, at)
	}
}

func TestSessionLabel(t *testing.T) {
	t.Parallel()
	if islazy.SessionLabel("ex.com", islazy.DefaultAPIScope) != "ex.com" {
		t.Fatal("default label")
	}
	if islazy.SessionLabel("ex.com", "auth") != "ex.com / auth" {
		t.Fatal("scoped label")
	}
}
