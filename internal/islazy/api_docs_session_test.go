// api_docs_session_test.go: Tests derivation of session paths such as APIScopeDir and SessionID.
package islazy

import (
	"path/filepath"
	"testing"
)

func TestAPIScopeDir(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"https://example.com/v3/api-docs":          DefaultAPIScope,
		"https://example.com/auth/v3/api-docs":     "auth",
		"https://example.com/house/api-docs.json":  "house",
		"https://example.com/auth/v1/swagger.json": "auth/v1",
	}
	for raw, want := range cases {
		if got := APIScopeDir(raw); got != want {
			t.Fatalf("%s: want %q got %q", raw, want, got)
		}
	}
}

func TestEncodeDecodeSessionID(t *testing.T) {
	t.Parallel()
	rel := "example.com/auth"
	id := EncodeSessionID(rel)
	if id != "example.com__auth" {
		t.Fatalf("encode: %q", id)
	}
	if filepath.ToSlash(DecodeSessionDir(id)) != rel {
		t.Fatalf("decode: %q", DecodeSessionDir(id))
	}
}
