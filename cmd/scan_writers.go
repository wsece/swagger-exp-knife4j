// scan_writers.go: Parses the --write-db/csv/jsonl flags and their corresponding *-file/-uri flags to determine the actual output paths.
package cmd

const (
	defaultScanCSVFile   = "result.csv"
	defaultScanJSONLFile = "result.jsonl"
)

// resolvedDbURI returns the database URI when --write-db is enabled or --write-db-uri is specified, otherwise returns an empty string.
func resolvedDbURI() string {
	if opts.Writer.DbURI != "" {
		return opts.Writer.DbURI
	}
	if opts.Writer.Db {
		return defaultSwaggerDBURI
	}
	return ""
}

// resolvedCsvFile returns the CSV path when --write-csv is enabled or --write-csv-file is specified, otherwise returns empty.
func resolvedCsvFile() string {
	if opts.Writer.CsvFile != "" {
		return opts.Writer.CsvFile
	}
	if opts.Writer.Csv {
		return defaultScanCSVFile
	}
	return ""
}

// resolvedJsonlFile returns the JSONL path when --write-jsonl is enabled or --write-jsonl-file is specified, otherwise returns empty.
func resolvedJsonlFile() string {
	if opts.Writer.JsonlFile != "" {
		return opts.Writer.JsonlFile
	}
	if opts.Writer.Jsonl {
		return defaultScanJSONLFile
	}
	return ""
}

// hasScanWriters determines if any persistent output (DB/CSV/JSONL) is configured.
func hasScanWriters() bool {
	return resolvedCsvFile() != "" || resolvedJsonlFile() != "" || resolvedDbURI() != ""
}

// scanOutputSatisfiesUser is true when the run produces a useful artifact without --write-*.
func scanOutputSatisfiesUser() bool {
	return opts.Scan.DocsOnly || hasScanWriters()
}
