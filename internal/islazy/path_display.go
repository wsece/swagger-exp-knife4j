// path_display.go: format filesystem paths for CLI display (relative to cwd when possible).
package islazy

import (
	"os"
	"path/filepath"
	"strings"
)

// RelativePathForDisplay returns path relative to the current working directory.
// Falls back to the cleaned absolute path when rel cannot be computed cleanly.
func RelativePathForDisplay(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	clean := filepath.Clean(path)
	cwd, err := os.Getwd()
	if err != nil {
		return clean
	}
	cwd = filepath.Clean(cwd)
	abs := clean
	if !filepath.IsAbs(abs) {
		if a, err := filepath.Abs(abs); err == nil {
			abs = a
		}
	}
	rel, err := filepath.Rel(cwd, abs)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return clean
	}
	return rel
}
