// targets_test.go: Unit tests for ReadTargetURLs
package islazy

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadTargetURLs(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "targets.txt")
	content := "# comment\n\nhttps://a.example.com\n  https://b.example.com  \n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	urls, err := ReadTargetURLs(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(urls) != 2 || urls[0] != "https://a.example.com" || urls[1] != "https://b.example.com" {
		t.Fatalf("got %v", urls)
	}
}
