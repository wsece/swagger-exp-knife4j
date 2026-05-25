// scan_shared.go: Bridge functions that convert cmd-level opts to scanrun.SingleParams / WriterConfig.
package cmd

import (
	"swagger-exp-knife4j/pkg/scanrun"
)

// scanWriterConfig parses the target paths for database/CSV/JSONL based on the --write-* flags.
func scanWriterConfig() scanrun.WriterConfig {
	return scanrun.WriterConfig{
		DbURI:     resolvedDbURI(),
		CsvFile:   resolvedCsvFile(),
		JsonlFile: resolvedJsonlFile(),
		DbDebug:   opts.Writer.DbDebug,
	}
}

// scanSingleParams assembles single scan parameters; the caller (e.g. file) will assign a value if inputURL is empty.
func scanSingleParams(inputURL string) scanrun.SingleParams {
	return scanrun.SingleParams{
		InputURL:  inputURL,
		OutputDir: opts.Scan.OutputDir,
		DocsOnly:  opts.Scan.DocsOnly,
		HTTP:      scanHTTPOptions(),
		Writers:   scanWriterConfig(),
	}
}
