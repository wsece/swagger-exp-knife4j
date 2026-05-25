// knife4j_proxy_test.go: tests for Knife4j proxy and TargetOrigin resolution.
package reportserver

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"swagger-exp-knife4j/internal/islazy"
)

func TestTargetOriginForSession_fromManifest(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	scopeDir := filepath.Join(dir, "api.apearth.com", "swagger", "v1")
	if err := os.MkdirAll(scopeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := islazy.DocManifest{
		Host:      "api.apearth.com",
		Scope:     "swagger/v1",
		JSONURL:   "https://api.apearth.com/swagger/v1/swagger.json",
		ScannedAt: time.Now(),
	}
	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(scopeDir, islazy.ManifestFileName), data, 0o644); err != nil {
		t.Fatal(err)
	}
	store := &Store{cfg: Config{APIDocPath: dir}}
	sessionID := islazy.EncodeSessionID("api.apearth.com/swagger/v1")
	got := store.TargetOriginForSession(sessionID)
	if got != "https://api.apearth.com" {
		t.Fatalf("got %q want https://api.apearth.com", got)
	}
}
