// scan_writers_test.go: Tests resolution of write paths such as resolvedDbURI, resolvedCsvFile, and resolvedJsonlFile.
package cmd

import (
	"testing"

	"swagger-exp-knife4j/pkg/config"
)

func TestResolvedWriters(t *testing.T) {
	t.Parallel()
	saved := *opts
	t.Cleanup(func() { *opts = saved })

	opts = &config.Options{}
	if resolvedDbURI() != "" || resolvedCsvFile() != "" || resolvedJsonlFile() != "" {
		t.Fatal("expected no writers by default")
	}

	opts.Writer.Db = true
	if resolvedDbURI() != defaultSwaggerDBURI {
		t.Fatalf("db %q", resolvedDbURI())
	}

	opts.Writer.Db = false
	opts.Writer.DbURI = "sqlite://custom.db"
	if resolvedDbURI() != "sqlite://custom.db" {
		t.Fatalf("db uri %q", resolvedDbURI())
	}

	opts.Writer.DbURI = ""
	opts.Writer.Csv = true
	if resolvedCsvFile() != defaultScanCSVFile {
		t.Fatalf("csv %q", resolvedCsvFile())
	}

	opts.Writer.Csv = false
	opts.Writer.CsvFile = "out.csv"
	if resolvedCsvFile() != "out.csv" {
		t.Fatalf("csv file %q", resolvedCsvFile())
	}

	opts.Writer.Jsonl = true
	if resolvedJsonlFile() != defaultScanJSONLFile {
		t.Fatalf("jsonl %q", resolvedJsonlFile())
	}
}
