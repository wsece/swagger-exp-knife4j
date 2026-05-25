package scanner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteReportToJSONL_appendSecondTarget(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.jsonl")
	stats := []APIStatisticsInfo{{Method: "GET", Path: "/a"}}
	reqs := []APIRequestResult{{Method: "GET", Path: "/a", StatusCode: 200}}

	if err := WriteReportToJSONL(path, "https://one.example/", "https://one.example/v3/api-docs", HTTPRequestMeta{}, stats, reqs, false); err != nil {
		t.Fatal(err)
	}
	if err := WriteReportToJSONL(path, "https://two.example/", "https://two.example/v3/api-docs", HTTPRequestMeta{}, stats, reqs, true); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 jsonl lines, got %d: %s", len(lines), string(data))
	}
	if !strings.Contains(lines[0], "one.example") || !strings.Contains(lines[1], "two.example") {
		t.Fatalf("missing targets in lines: %v", lines)
	}
}

func TestWriteReportToCSV_appendSecondTarget(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.csv")
	stats := []APIStatisticsInfo{{Method: "GET", Path: "/a"}}
	reqs := []APIRequestResult{{Method: "GET", Path: "/a", FullURL: "https://h/a", StatusCode: 200}}

	if err := WriteReportToCSV(path, "https://one.example/", stats, reqs, false); err != nil {
		t.Fatal(err)
	}
	if err := WriteReportToCSV(path, "https://two.example/", stats, reqs, true); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 3 { // header + 2 data rows
		t.Fatalf("expected 3 csv lines, got %d:\n%s", len(lines), string(data))
	}
	if !strings.Contains(lines[1], "one.example") || !strings.Contains(lines[2], "two.example") {
		t.Fatalf("missing targets in csv:\n%s", string(data))
	}
}
