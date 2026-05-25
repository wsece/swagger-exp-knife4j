package islazy

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRelativePathForDisplay_underCwd(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	abs := filepath.Join(cwd, "output", "host", "api-docs.json")
	got := RelativePathForDisplay(abs)
	want := filepath.Join("output", "host", "api-docs.json")
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
