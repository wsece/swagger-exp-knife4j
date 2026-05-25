// targets.go: Reads batch scan target URLs from text files (one per line, supports # comments).
package islazy

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ReadTargetURLs reads target URLs from the specified path, one per line.
// Empty lines and lines starting with # are skipped.
// Returns a list of URLs; returns an error if the file is empty or contains no valid lines.
func ReadTargetURLs(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open targets file: %w", err)
	}
	defer f.Close()

	var urls []string
	sc := bufio.NewScanner(f)
	lineNo := 0
	for sc.Scan() {
		lineNo++
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		urls = append(urls, line)
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("read targets file line %d: %w", lineNo, err)
	}
	if len(urls) == 0 {
		return nil, fmt.Errorf("no targets found in %s", path)
	}
	return urls, nil
}
